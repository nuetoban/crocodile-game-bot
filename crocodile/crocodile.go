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

package crocodile

import (
	"errors"
	"sync"
	"time"

	"github.com/looplab/fsm"

	"gitlab.com/pviktor/crocodile-game-bot/model"
	"gitlab.com/pviktor/crocodile-game-bot/utils"
)

const (
	// ErrGameAlreadyStarted is error when user tries to start a game in chat, but there is the one already
	ErrGameAlreadyStarted = "Game already started"

	// ErrWaitingForWinnerRespond is error when the game have been played, but winner did't start a new one
	ErrWaitingForWinnerRespond = "Waiting for winner respond"
)

// WordsProvider should return random word
type WordsProvider interface {
	GetWord() (string, error)
}

// Storage aims to save FSM state somewhere (e.g. in Redis)
type Storage interface {
	IncrementUserStats(model.UserInChat) error
}

// Machine stores state of game in one chat
type Machine struct {
	MesID int

	// ChatID where the game is started
	ChatID int64

	// Word which users should guess
	Word string

	// UserID of the user that should explain the word
	Host     int
	HostName string

	// UserID of user who guessed the word
	Winner int

	StartedTime time.Time
	GuessedTime time.Time

	// Technical data
	Storage       Storage       `json:"-"`
	WordsProvider WordsProvider `json:"-"`
	FSM           *fsm.FSM      `json:"-"`
	Log           Logger        `json:"-"`

	mutex *sync.Mutex
}

// MachineFabric aims to produce new machines with freezed Storage and WordsProvider
type MachineFabric struct {
	Storage       Storage
	WordsProvider WordsProvider
	Log           Logger
}

// NewMachine returns Machine with freezed Storage and WordsProvider
func (m *MachineFabric) NewMachine(chatID int64, mesID int) *Machine {
	return NewMachine(m.Storage, m.WordsProvider, m.Log, chatID, mesID)
}

// NewMachineFabric returns MachineFabric
func NewMachineFabric(storage Storage, wp WordsProvider, log Logger) *MachineFabric {
	return &MachineFabric{
		Storage:       storage,
		WordsProvider: wp,
		Log:           log,
	}
}

// NewMachine returns new Machine instance
func NewMachine(storage Storage, wp WordsProvider, log Logger, chatID int64, mesID int) *Machine {
	fsm := fsm.NewFSM(
		"init",
		fsm.Events{
			{Name: "new_game", Src: []string{"init", "done"}, Dst: "game_started"},
			{Name: "stop_game", Src: []string{"game_started"}, Dst: "done"},
		},
		fsm.Callbacks{},
	)

	m := &Machine{
		ChatID:        chatID,
		FSM:           fsm,
		Storage:       storage,
		WordsProvider: wp,
		MesID:         mesID,
		mutex:         &sync.Mutex{},
		StartedTime:   time.Now(),
		GuessedTime:   time.Now(),
		Log:           log,
	}

	return m
}

// StartNewGameAndReturnWord sets m.Word to new words and returns it
func (m *Machine) StartNewGameAndReturnWord(host int, hostName string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Log.Debugf("Starting new game, host: %d, hostName: %s", host, hostName)

	if m.FSM.Cannot("new_game") {
		m.Log.Debugf("StartNewGameAndReturnWord: already started, machine: %+v", m)
		return "", errors.New(ErrGameAlreadyStarted)
	}

	_, _, ss := utils.CalculateTimeDiff(time.Now(), m.GetGuessedTime())
	if host != m.GetWinner() && m.GetWinner() != 0 && ss < 5 {
		m.Log.Debug("StartNewGameAndReturnWord: waiting for winner respond")
		return "", errors.New(ErrWaitingForWinnerRespond)
	}

	var err error
	m.Word, err = m.WordsProvider.GetWord()
	if err != nil {
		m.Log.Warningf("StartNewGameAndReturnWord: error during getting word: %v", err)
		return "", err
	}

	m.Host = host
	m.StartedTime = time.Now()
	m.FSM.Event("new_game")

	m.Storage.IncrementUserStats(model.UserInChat{
		ID:      m.Host,
		ChatID:  m.ChatID,
		WasHost: 1,
		Name:    hostName,
	})
	m.HostName = hostName

	m.Log.Debugf("StartNewGameAndReturnWord: returning word: \"%s\"", m.Word)
	return m.Word, nil
}

// SetNewRandomWord generates new word
func (m *Machine) SetNewRandomWord() (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var err error

	m.Word, err = m.WordsProvider.GetWord()
	if err != nil {
		m.Log.Warningf("SetNewRandomWord: error during getting word: %v", err)
		return "", err
	}

	m.Log.Tracef("SetNewRandomWord: setting word for chat (%d): %s", m.ChatID, m.Word)

	return m.Word, nil
}

// GetWord is getter for m.Word
func (m *Machine) GetWord() string { return m.Word }

// GetHost is getter for m.Host
func (m *Machine) GetHost() int { return m.Host }

// GetStartedTime is getter for m.StartedTime
func (m *Machine) GetStartedTime() time.Time { return m.StartedTime }

// GetGuessedTime is getter for m.GuessedTime
func (m *Machine) GetGuessedTime() time.Time { return m.GuessedTime }

// GetWinner is getter for m.Winner
func (m *Machine) GetWinner() int { return m.Winner }

// CheckWord checks if m.Word == provided word
func (m *Machine) CheckWord(word string) bool { return word == m.Word }

// CheckWordAndSetWinner sets m.Winner and returns true if m.CheckWord() returns true, otherwise ret. false
func (m *Machine) CheckWordAndSetWinner(word string, potentialWinner int, winnerName string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Log.Debugf(
		"CheckWordAndSetWinner: checking word: %s, potentialWinner: %d, winnerName: %s, chatID: %d",
		word, potentialWinner, winnerName, m.ChatID,
	)

	if m.FSM.Current() != "game_started" {
		m.Log.Debugf("CheckWordAndSetWinner: game is not in state \"game_started\", chatID: %d", m.ChatID)
		return false
	}

	if m.CheckWord(word) {
		m.Log.Debugf("CheckWordAndSetWinner: stopping game, chatID: %d", m.ChatID)
		m.FSM.Event("stop_game")
		m.Winner = potentialWinner
		m.GuessedTime = time.Now()

		winner := model.UserInChat{
			ID:      m.Winner,
			ChatID:  m.ChatID,
			Guessed: 1,
			Name:    winnerName,
		}
		err := m.Storage.IncrementUserStats(winner)
		if err != nil {
			m.Log.Errorf("CheckWordAndSetWinner: cannot increment winner stats: %+v", winner)
		}

		host := model.UserInChat{
			ID:      m.Host,
			ChatID:  m.ChatID,
			Success: 1,
			Name:    m.HostName,
		}
		err = m.Storage.IncrementUserStats(host)
		if err != nil {
			m.Log.Errorf("CheckWordAndSetWinner: cannot increment host stats: %+v", host)
		}

		return true
	}

	return false
}

// StopGame sends stop_game event to FSM
func (m *Machine) StopGame() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Log.Debugf("Stopping game, machine: %+v", m)

	m.FSM.Event("stop_game")
	return nil
}
