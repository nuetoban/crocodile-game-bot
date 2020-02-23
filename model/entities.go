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

package model

type UserInChat struct {
	ID     int
	ChatID int64

	Name string

	// When user was a Host
	WasHost int
	Success int

	// When user was a guesser
	Guessed int
}

type Statistics struct {
	Chats       int64
	Users       int64
	GamesPlayed int64
}

type Chat struct {
	ID    int64
	Title string
}

type ChatStatistics struct {
	Title string

	// How many games have been done
	Guessed int
}
