/*
 * This file is part of Crocodile Game Bot.
 * Copyright (C) 2019  Viktor
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/nuetoban/crocodile-game-bot/crocodile"
	"github.com/nuetoban/crocodile-game-bot/model"
	"github.com/nuetoban/crocodile-game-bot/storage"
	"github.com/nuetoban/crocodile-game-bot/utils"
)

var (
	mutexFabric  *redsync.Redsync
	locks        map[int64]*sync.Mutex
	lockForLocks *sync.Mutex // ну а хули нам
	machines     map[int64]*crocodile.Machine
	fabric       *crocodile.MachineFabric
	bot          *tb.Bot
	redisPool    *redis.Pool

	textUpdatesRecieved float64
	startTotal          float64
	ratingTotal         float64
	globalRatingTotal   float64
	cstatTotal          float64
	chatsRatingTotal    float64

	updatesProcessed int

	wordsInlineKeys   [][]tb.InlineButton
	newGameInlineKeys [][]tb.InlineButton
	ratingGetter      RatingGetter
	statisticsGetter  StatisticsGetter

	rateLimiter *RateLimiter

	DEBUG = false
)

func init() {
	lockForLocks = &sync.Mutex{}
	locks = make(map[int64]*sync.Mutex)

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = ":6379"
	}
	redisPool = newPool(redisHost)
	mutexFabric = redsync.New([]redsync.Pool{redisPool})
	cleanupHook()
}

// https://github.com/pete911/examples-redigo
func newPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     200,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				panic(err)
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				panic(err)
			}
			return err
		},
	}
}

// https://github.com/pete911/examples-redigo
func cleanupHook() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		redisPool.Close()
		os.Exit(0)
	}()
}

type RatingGetter interface {
	GetRating(chatID int64) ([]model.UserInChat, error)
	GetGlobalRating() ([]model.UserInChat, error)
	GetChatsRating() ([]model.ChatStatistics, error)
}

type StatisticsGetter interface {
	GetStatistics() (model.Statistics, error)
}

type dbCredentials struct {
	Host,
	User,
	Pass,
	Name string

	Port int
	KW   storage.KW
}

func loggerMiddlewarePoller(upd *tb.Update) bool {
	if upd.Message != nil && upd.Message.Chat != nil && upd.Message.Sender != nil {
		log.Debugf(
			"Received update, chat: %d, chatTitle: \"%s\", user: %d",
			upd.Message.Chat.ID,
			upd.Message.Chat.Title,
			upd.Message.Sender.ID,
		)
	}
	updatesProcessed++
	return true
}

func getDbCredentialsFromEnv() (dbCredentials, error) {
	prefix := "CROCODILE_GAME_DB_"
	ret := dbCredentials{}
	ret.KW = storage.KW{}
	envList := os.Environ()
	env := make(map[string]string)
	for _, v := range envList {
		kv := strings.Split(v, "=")
		env[kv[0]] = kv[1]
	}
	var err error

	ret.Port, err = strconv.Atoi(env[prefix+"PORT"])
	if err != nil {
		return ret, err
	}
	delete(env, prefix+"PORT")

	ret.Host = env[prefix+"HOST"]
	ret.User = env[prefix+"USER"]
	ret.Pass = env[prefix+"PASS"]
	ret.Name = env[prefix+"NAME"]
	delete(env, prefix+"HOST")
	delete(env, prefix+"USER")
	delete(env, prefix+"PASS")
	delete(env, prefix+"NAME")

	for k, v := range env {
		if strings.HasPrefix(k, prefix) {
			ret.KW[strings.ToLower(strings.TrimPrefix(k, prefix))] = v
		}
	}

	return ret, nil
}

func main() {
	logInit()
	if os.Getenv("CROCODILE_GAME_DEV") != "" {
		DEBUG = true
		setLogLevel("TRACE")
	}

	if os.Getenv("CROCODILE_GAME_LOGLEVEL") != "" {
		setLogLevel(os.Getenv("CROCODILE_GAME_LOGLEVEL"))
	}

	log.Info("Loading words")
	f, err := os.Open("dictionaries/word_rus_min.txt")
	if err != nil {
		log.Fatalf("Cannot open dictionary: %v", err)
	}
	wordsProvider, _ := crocodile.NewWordsProviderReader(f)

	log.Info("Readind DB env variables")
	creds, err := getDbCredentialsFromEnv()
	if err != nil {
		log.Fatalf("Cannot get database credentials from ENV: %v", err)
	}

	log.Info("Connecting to the database")
	pg, err := storage.NewStorage(storage.NewConnString(
		creds.Host, creds.User,
		creds.Pass, creds.Name,
		creds.Port, creds.KW,
	), redisPool, storage.WrapLogrus(log))
	if err != nil {
		log.Fatalf("Cannot connect to database (%s, %s) on host %s: %v", creds.User, creds.Name, creds.Host, err)
	}

	ratingGetter = pg
	statisticsGetter = pg

	log.Info("Creating games fabric")
	fabric = crocodile.NewMachineFabric(pg, wordsProvider, log)
	machines = make(map[int64]*crocodile.Machine)

	rateLimiter = NewRateLimiter(redisPool)

	log.Info("Connecting to Telegram API")
	var poller tb.Poller
	if os.Getenv("CROCODILE_GAME_WEBHOOK") != "" {
		poller = &tb.Webhook{
			Endpoint: &tb.WebhookEndpoint{
				PublicURL: os.Getenv("CROCODILE_GAME_WEBHOOK"),
			},
			Listen: "0.0.0.0:9999",
		}
	} else {
		poller = &tb.LongPoller{Timeout: 5 * time.Second}
	}

	mp := tb.NewMiddlewarePoller(poller, loggerMiddlewarePoller)
	mp.Capacity = 10000

	settings := tb.Settings{
		Token:   os.Getenv("CROCODILE_GAME_BOT_TOKEN"),
		Poller:  mp,
		Updates: 10000,
	}
	bot, err = tb.NewBot(settings)
	if err != nil {
		log.Fatalf("Cannot connect to Telegram API: %v", err)
	}
	// pg.SetBotID(bot.Me.ID)

	log.Info("Binding handlers")
	bot.Handle(tb.OnText, logDuration(mustLock(textHandler)))
	bot.Handle("/start", logDuration(mustLock(startNewGameHandler)))
	bot.Handle("/rating", logDuration(ratingHandler))
	bot.Handle("/globalrating", logDuration(globalRatingHandler))
	bot.Handle("/cancel", func(m *tb.Message) {})
	bot.Handle("/cstat", logDuration(statsHandler))
	bot.Handle("/rules", logDuration(rulesHandler))
	bot.Handle("/chatrating", logDuration(chatsRatingHandler))
	bindButtonsHandlers(bot)

	collector := newMetricsCollector(pg)
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())

	log.Info("Starting metrics exporter server")
	go http.ListenAndServe(":8080", nil)

	go func() {
		var err error
		for {
			time.Sleep(time.Second * 15)

			if updatesProcessed == 0 {
				err = ioutil.WriteFile("/tmp/crocostatus", []byte("BAD"), 0644)
			} else {
				err = ioutil.WriteFile("/tmp/crocostatus", []byte("GOOD"), 0644)
			}

			updatesProcessed = 0
			if err != nil {
				panic(err)
			}
		}
	}()

	log.Info("Starting the bot")
	bot.Start()
}

// Decorator for logging duration of function execution
func logDuration(f func(*tb.Message)) func(*tb.Message) {
	return func(m *tb.Message) {
		start := time.Now()
		f(m)
		diff := time.Now().Sub(start)
		if diff.Seconds() > 1 {
			log.Warnf("Took %s time to complete update processing", time.Time{}.Add(diff).Format("04:05.000"))
		} else {
			log.Tracef("Took %s time to complete update processing", time.Time{}.Add(diff).Format("04:05.000"))
		}
	}
}

// Decorator for logging duration of function execution (callback handlers)
func logDurationCallback(f func(*tb.Callback)) func(*tb.Callback) {
	return func(c *tb.Callback) {
		start := time.Now()
		f(c)
		diff := time.Now().Sub(start)
		if diff.Seconds() > 1 {
			log.Warnf("Took %s time to complete update processing (cb)", time.Time{}.Add(diff).Format("04:05.000"))
		} else {
			log.Tracef("Took %s time to complete update processing (cb)", time.Time{}.Add(diff).Format("04:05.000"))
		}
	}
}

// Decorator for distributed lock for chat (messages handlers)
func mustLock(f func(*tb.Message)) func(*tb.Message) {
	return func(m *tb.Message) {
		go func() {
			m := m
			log.Tracef("Locking chat %d", m.Chat.ID)
			lockChat(m.Chat.ID)

			f(m)

			log.Tracef("Unlocking chat %d", m.Chat.ID)
			unlockChat(m.Chat.ID)
		}()
	}
}

// Decorator for distributed lock for chat (callback handlers)
func mustLockCallback(f func(*tb.Callback)) func(*tb.Callback) {
	return func(c *tb.Callback) {
		go func() {
			c := c
			log.Tracef("Locking chat %d", c.Message.Chat.ID)
			lockChat(c.Message.Chat.ID)

			f(c)

			log.Tracef("Unlocking chat %d", c.Message.Chat.ID)
			unlockChat(c.Message.Chat.ID)
		}()
	}
}

func globalRatingHandler(m *tb.Message) {
	globalRatingTotal++
	rating, err := ratingGetter.GetGlobalRating()
	if err != nil {
		log.Errorf("globalRatingHandler: cannot get rating %v:", err)
		return
	}

	ratingString := buildRating("Топ-25 <b>игроков в крокодила</b> во всех чатах 🐊", rating)

	err = sendMessage(m.Chat, m.Chat.ID, ratingString)
	if err != nil {
		log.Errorf("globalRatingHandler: cannot send rating: %v", err)
	}
}

func buildRating(header string, data []model.UserInChat) string {
	if len(data) < 1 {
		return "Данных пока недостаточно!"
	}

	out := header + "\n\n"
	for k, v := range data {
		out += fmt.Sprintf(
			"<b>%d</b>. %s — %d %s.\n",
			k+1,
			html.EscapeString(v.Name),
			v.Guessed,
			utils.DetectCaseAnswers(v.Guessed),
		)
	}

	return out
}

func buildRatingChatStatistics(header string, data []model.ChatStatistics) string {
	if len(data) < 1 {
		return "Данных пока недостаточно!"
	}

	out := header + "\n\n"
	for k, v := range data {
		out += fmt.Sprintf(
			"<b>%d</b>. %s — %d %s.\n",
			k+1,
			html.EscapeString(v.Title),
			v.Guessed,
			utils.DetectCaseForGames(v.Guessed),
		)
	}

	return out
}

func ratingHandler(m *tb.Message) {
	ratingTotal++
	rating, err := ratingGetter.GetRating(m.Chat.ID)
	if err != nil {
		log.Errorf("ratingHandler: cannot get rating %v:", err)
		return
	}

	ratingString := buildRating("Топ-25 <b>игроков в крокодила</b> 🐊", rating)

	err = sendMessage(m.Chat, m.Chat.ID, ratingString)
	if err != nil {
		log.Errorf("ratingHandler: cannot send rating: %v", err)
	}
}

func sendMessage(s tb.Recipient, chatID int64, text string) error {
	err := rateLimiter.Limit(chatID,
		func() error { _, err := bot.Send(s, text, tb.ModeHTML, tb.NoPreview); return err },
		func() error {
			_, err := bot.Send(s, "Достигнут лимит по количеству сообщений в минуту!")
			return err
		},
		func() error { return nil })
	return err
}

func statsHandler(m *tb.Message) {
	cstatTotal++
	stats, err := statisticsGetter.GetStatistics()
	if err != nil {
		log.Errorf("statsHandler: cannot get stats %v:", err)
		return
	}

	outString := "<b>Статистика крокодила</b> 🐊\n\n"
	outString += fmt.Sprintf("Количество чатов: %d\n", stats.Chats)
	outString += fmt.Sprintf("Количество игроков: %d\n", stats.Users)
	outString += fmt.Sprintf("Всего игр: %d\n", stats.GamesPlayed)

	err = sendMessage(m.Chat, m.Chat.ID, outString)
	if err != nil {
		log.Errorf("statsHandler: cannot send stats: %v", err)
	}
}

func lockChat(chatID int64) error {
	if _, ok := locks[chatID]; ok {
		if locks[chatID] != nil {
			locks[chatID].Lock()
		} else {
			lockForLocks.Lock()
			locks[chatID] = &sync.Mutex{}
			lockForLocks.Unlock()
			locks[chatID].Lock()
		}
	} else {
		lockForLocks.Lock()
		locks[chatID] = &sync.Mutex{}
		lockForLocks.Unlock()
		locks[chatID].Lock()
	}
	return nil
}

func unlockChat(chatID int64) {
	// mutexFabric.NewMutex("mutex/" + strconv.Itoa(int(chatID))).Unlock()
	locks[chatID].Unlock()
}

func startNewGameHandler(m *tb.Message) {
	if m.Private() {
		sendMessage(m.Sender, m.Chat.ID, "Добавить бота в чат: https://t.me/Crocodile_Game_Bot?startgroup=a ")
		return
	}

	startTotal++

	machine := fabric.NewMachine(m.Chat.ID, m.ID)

	username := strings.TrimSpace(m.Sender.FirstName + " " + m.Sender.LastName)

	_, err := machine.StartNewGameAndReturnWord(m.Sender.ID, username, m.Chat.Title)

	if err != nil {
		if err.Error() == crocodile.ErrGameAlreadyStarted {
			_, ms, _ := utils.CalculateTimeDiff(time.Now(), machine.GetStartedTime())

			if ms < 2 {
				sendMessage(m.Chat, m.Chat.ID, "Игра уже начата! Ожидайте 2 минуты")
				return
			} else {
				machine.StopGame()
				_, err = machine.StartNewGameAndReturnWord(m.Sender.ID, username, m.Chat.Title)
				if err != nil {
					log.Println(err)
				}
			}
		} else {
			log.Println(err)
			return
		}
	}

	bot.Send(
		m.Chat,
		fmt.Sprintf(
			`<a href="tg://user?id=%d">%s</a> объясняет слово`,
			m.Sender.ID, html.EscapeString(m.Sender.FirstName)),
		tb.ModeHTML,
		&tb.ReplyMarkup{InlineKeyboard: wordsInlineKeys},
	)
}

func startNewGameHandlerCallback(c *tb.Callback) {
	m := c.Message

	// If machine for this chat has been created already
	ma := fabric.NewMachine(m.Chat.ID, m.ID)

	username := strings.TrimSpace(c.Sender.FirstName + " " + c.Sender.LastName)
	_, err := ma.StartNewGameAndReturnWord(c.Sender.ID, username, m.Chat.Title)

	if err != nil {
		if err.Error() == crocodile.ErrGameAlreadyStarted {
			_, ms, _ := utils.CalculateTimeDiff(time.Now(), ma.GetStartedTime())

			if ms < 2 {
				bot.Respond(c, &tb.CallbackResponse{Text: "Игра уже начата! Ожидайте 2 минуты"})
				return
			} else {
				ma.StopGame()
				_, err = ma.StartNewGameAndReturnWord(c.Sender.ID, username, m.Chat.Title)
				if err != nil {
					log.Println(err)
				}
				bot.Respond(c, &tb.CallbackResponse{
					Text:      fmt.Sprintf("Ты — ведущий, твое слово — %s", ma.GetWord()),
					ShowAlert: true,
				})
			}
		} else if err.Error() == crocodile.ErrWaitingForWinnerRespond {
			bot.Respond(c, &tb.CallbackResponse{Text: "У победителя есть 5 секунд на решение!"})
			return
		} else {
			log.Println(err)
			bot.Respond(c, &tb.CallbackResponse{Text: "."})
			return
		}
	}

	bot.Respond(c, &tb.CallbackResponse{
		Text:      fmt.Sprintf("Ты — ведущий, твое слово — %s", ma.GetWord()),
		ShowAlert: true,
	})
	bot.Send(
		m.Chat,
		fmt.Sprintf(
			`<a href="tg://user?id=%d">%s</a> объясняет слово`,
			c.Sender.ID, html.EscapeString(c.Sender.FirstName)),
		tb.ModeHTML,
		&tb.ReplyMarkup{InlineKeyboard: wordsInlineKeys},
	)
}

func textHandler(m *tb.Message) {
	textUpdatesRecieved++

	ma := fabric.NewMachine(m.Chat.ID, m.ID)

	if ma.GetHost() != m.Sender.ID || DEBUG {
		username := strings.TrimSpace(m.Sender.FirstName + " " + m.Sender.LastName)
		if word, ok := ma.CheckWordAndSetWinner(m.Text, m.Sender.ID, username); ok {
			bot.Send(
				m.Chat,
				fmt.Sprintf(
					"%s отгадал(а) слово <b>%s</b>",
					username, word,
				),
				tb.ModeHTML,
				&tb.ReplyMarkup{InlineKeyboard: newGameInlineKeys},
			)
		}
	}
}

func seeWordCallbackHandler(c *tb.Callback) {
	m := fabric.NewMachine(c.Message.Chat.ID, c.Message.ID)
	var message string

	if c.Sender.ID != m.GetHost() {
		message = "Это слово предназначено не для тебя!"
	} else {
		message = m.GetWord()
	}

	bot.Respond(c, &tb.CallbackResponse{Text: message, ShowAlert: true})
}

func nextWordCallbackHandler(c *tb.Callback) {
	m := fabric.NewMachine(c.Message.Chat.ID, c.Message.ID)
	var message string
	var err error

	if c.Sender.ID != m.GetHost() {
		message = "Это слово предназначено не для тебя!"
	} else {
		message, err = m.SetNewRandomWord()
		if err != nil {
			log.Errorf("nextWordCallbackHandler: cannot get word: %v", err)
			bot.Respond(c, &tb.CallbackResponse{Text: message, ShowAlert: true})
			return
		}
	}

	bot.Respond(c, &tb.CallbackResponse{Text: message, ShowAlert: true})
}

func bindButtonsHandlers(bot *tb.Bot) {
	seeWord := tb.InlineButton{Unique: "see_word", Text: "Посмотреть слово"}
	nextWord := tb.InlineButton{Unique: "next_word", Text: "Следующее слово"}
	newGame := tb.InlineButton{Unique: "new_game", Text: "Хочу быть ведущим!"}

	wordsInlineKeys = [][]tb.InlineButton{[]tb.InlineButton{seeWord}, []tb.InlineButton{nextWord}}
	newGameInlineKeys = [][]tb.InlineButton{[]tb.InlineButton{newGame}}

	bot.Handle(&newGame, logDurationCallback(mustLockCallback(startNewGameHandlerCallback)))
	bot.Handle(&seeWord, logDurationCallback(mustLockCallback(seeWordCallbackHandler)))
	bot.Handle(&nextWord, logDurationCallback(mustLockCallback(nextWordCallbackHandler)))
}

func rulesHandler(m *tb.Message) {
	sendMessage(m.Chat, m.Chat.ID, `
<b>ПРАВИЛА ИГРЫ В КРОКОДИЛА</b>

Есть ведущий и есть игроки, которые отгадывают слова.

После нажатия /start@Crocodile_Game_Bot задача ведущего — нажать кнопку "Посмотреть слово" и объяснить его, не используя однокоренные слова.
Если слово не нравится, то можно нажать "Следующее слово".
Задача игроков — отгадать загаданное слово, для этого нужно просто писать их в чат, по одному слову в сообщении.
`)
}

func chatsRatingHandler(m *tb.Message) {
	chatsRatingTotal++
	rating, err := ratingGetter.GetChatsRating()
	if err != nil {
		log.Errorf("chatsRatingHandler: cannot get rating %v:", err)
		return
	}

	ratingString := buildRatingChatStatistics("Топ-25 <b>чатов по количеству игр в крокодила</b>🐊", rating)

	err = sendMessage(m.Chat, m.Chat.ID, ratingString)
	if err != nil {
		log.Errorf("chatsRatingHandler: cannot send rating: %v", err)
	}
}
