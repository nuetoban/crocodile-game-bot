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
	"sync"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/nuetoban/crocodile-game-bot/model"
)

var p *Postgres

func TestMain(m *testing.M) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&model.UserInChat{})

	p = &Postgres{
		db:    db,
		mutex: &sync.Mutex{},
	}

	m.Run()
}

func TestPostgresIncrementUserStats(t *testing.T) {
	c := func(name string, expected, got interface{}) {
		if expected != got {
			t.Errorf("Wrong \"%s\": Got: %v, expected: %v", name, got, expected)
		}
	}

	u := model.UserInChat{
		ID:     69,
		ChatID: 420,
	}

	user := model.UserInChat{
		ID:      69,
		ChatID:  420,
		Name:    "test-name",
		WasHost: 1,
		Success: 2,
		Guessed: 3,
	}

	// Start testing
	p.IncrementUserStats(user)
	p.db.First(&u)
	t.Logf("User: %#v\n", u)

	c("WasHost - first", 1, u.WasHost)
	c("Success - first", 2, u.Success)
	c("Guessed - first", 3, u.Guessed)
	c("Name - first", "test-name", u.Name)

	// Do the same one more time
	p.IncrementUserStats(user)
	p.db.First(&u)
	t.Logf("User: %#v\n", u)

	c("WasHost - second", 2, u.WasHost)
	c("Success - second", 4, u.Success)
	c("Guessed - second", 6, u.Guessed)
	c("Name - second", "test-name", u.Name)
}
