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

package storage

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"fmt"

	"github.com/nuetoban/crocodile-game-bot/model"
)

type Postgres struct {
	db *gorm.DB

	log Logger
}

type KW map[string]interface{}

func NewConnString(host, user, pass, dbname string, port int, kw KW) string {
	baseString := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s",
		host, port, user, dbname, pass,
	)

	for k, v := range kw {
		baseString += fmt.Sprintf(" %s=%v", k, v)
	}

	return baseString
}

func NewPostgres(conn string, logger Logger) (*Postgres, error) {
	db, err := gorm.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	db.SetLogger(logger)
	db.LogMode(true)

	return &Postgres{
		db: db,
	}, nil
}

func (p *Postgres) IncrementUserStats(chat model.Chat, givenUser ...model.UserInChat) error {
	tx := p.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	for _, u := range givenUser {
		var (
			user model.UserInChat
			err  error
		)

		err = tx.FirstOrCreate(&model.Chat{}, model.Chat{ID: chat.ID}).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.FirstOrCreate(&user, model.UserInChat{
			ID:     u.ID,
			ChatID: u.ChatID,
		}).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Table("chats").
			Where("id = ?", chat.ID).
			Updates(map[string]interface{}{
				"title":     chat.Title,
			}).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Table("user_in_chats").
			Where("id = ? AND chat_id = ?", u.ID, u.ChatID).
			Updates(map[string]interface{}{
				"name":     u.Name,
				"was_host": user.WasHost + u.WasHost,
				"success":  user.Success + u.Success,
				"guessed":  user.Guessed + u.Guessed,
			}).Error
		if err != nil {
			tx.Rollback()
			return err
		}

	}

	return tx.Commit().Error
}

func (p *Postgres) GetRating(chatID int64) ([]model.UserInChat, error) {
	var users []model.UserInChat
	p.db.Where("guessed > 0 AND chat_id = ?", chatID).Limit(25).Order("guessed desc").Find(&users)
	return users, nil
}

func (p *Postgres) GetGlobalRating() ([]model.UserInChat, error) {
	var users []model.UserInChat

	rows, err := p.db.Table("user_in_chats").
		Select("sum(\"guessed\") as guessed, (array_agg(\"name\"))[1] as name, \"id\"").
		Group("id").
		Limit(25).
		Order("guessed desc").
		Having("sum(\"guessed\") > ?", 0).
		Rows()
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user model.UserInChat
		p.db.ScanRows(rows, &user)
		users = append(users, user)
	}

	return users, nil
}

func (p *Postgres) GetStatistics() (model.Statistics, error) {
	result := model.Statistics{}

	p.db.Raw(`SELECT
                 (SELECT COUNT(DISTINCT("chat_id")) FROM user_in_chats WHERE "id" != "chat_id") AS chats,
                 (SELECT COUNT(DISTINCT("id")) FROM user_in_chats) AS users,
                 (SELECT SUM("was_host") FROM user_in_chats) AS games_played;`).
		Scan(&result)

	return result, nil
}

func (p *Postgres) GetChatsRating() ([]model.ChatStatistics, error) {
	var chats []model.ChatStatistics

	rows, err := p.db.Table("chats").
		Select("sum(user_in_chats.\"guessed\") as guessed, chats.title").
		Joins("inner join user_in_chats on user_in_chats.chat_id = chats.id").
		Where("chats.title != ''").
		Group("chats.id").
		Order("guessed desc").
		Limit(25).
		Having("sum(\"guessed\") > ?", 0).
		Rows()
	if err != nil {
		return chats, err
	}
	defer rows.Close()

	for rows.Next() {
		var chat model.ChatStatistics
		p.db.ScanRows(rows, &chat)
		chats = append(chats, chat)
	}

	return chats, nil
}
