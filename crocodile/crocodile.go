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
	"math"
	"sync"
	"time"

	"github.com/looplab/fsm"

	"gitlab.com/pviktor/crocodile-game-bot/utils"
)

const (
	// ErrGameAlreadyStarted is error when user tries to start a game in chat, but there is the one already
	ErrGameAlreadyStarted      = "Game already started"
	ErrWaitingForWinnerRespond = "Waiting for winner respond"
)

// Game stores methods to manage one game
type Game interface {
	StartNewGameAndReturnWord(host int64) (string, error)
	SetNewRandomWord() (string, error)
	GetWord() string
	GetHost() string
	CheckWord(word string) bool
}

// WordsProvider should return random word
type WordsProvider interface {
	GetWord() (string, error)
}

// Storage aims to save FSM state somewhere (e.g. in Redis)
type Storage interface {
	Save(Game) error
}

// Machine stores state of game in one chat
type Machine struct {
	MesID int

	// ChatID where the game is started
	ChatID int64

	// Word which users should guess
	Word string

	// UserID of the user that should explain the word
	Host int

	// UserID of user who guessed the word
	Winner int

	StartedTime time.Time
	GuessedTime time.Time

	// Technical data
	Storage       Storage
	WordsProvider WordsProvider
	FSM           *fsm.FSM
	mutex         *sync.Mutex
}

// MachineFabric aims to produce new machines with freezed Storage and WordsProvider
type MachineFabric struct {
	Storage       Storage
	WordsProvider WordsProvider
}

// NewMachine returns Machine with freezed Storage and WordsProvider
func (m *MachineFabric) NewMachine(chatID int64, mesID int) *Machine {
	return NewMachine(m.Storage, m.WordsProvider, chatID, mesID)
}

// NewMachineFabric returns MachineFabric
func NewMachineFabric(storage Storage, wp WordsProvider) *MachineFabric {
	return &MachineFabric{
		Storage:       storage,
		WordsProvider: wp,
	}
}

// NewMachine returns new Machine instance
func NewMachine(storage Storage, wp WordsProvider, chatID int64, mesID int) *Machine {
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
	}

	return m
}

// StartNewGameAndReturnWord sets m.Word to new words and returns it
func (m *Machine) StartNewGameAndReturnWord(host int) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.FSM.Cannot("new_game") {
		return "", errors.New(ErrGameAlreadyStarted)
	}

	_, _, ss := utils.CalculateTimeDiff(time.Now(), m.GetGuessedTime())
	if host != m.GetWinner() && m.GetWinner() != 0 && ss < 5 {
		return "", errors.New(ErrWaitingForWinnerRespond)
	}

	var err error
	m.Word, err = m.WordsProvider.GetWord()
	if err != nil {
		return "", err
	}

	m.Host = host
	m.StartedTime = time.Now()
	m.FSM.Event("new_game")

	return m.Word, nil
}

// SetNewRandomWord generates new word
func (m *Machine) SetNewRandomWord() (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var err error

	m.Word, err = m.WordsProvider.GetWord()
	if err != nil {
		return "", err
	}

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
func (m *Machine) CheckWordAndSetWinner(word string, potentialWinner int) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.FSM.Current() != "game_started" {
		return false
	}

	if m.CheckWord(word) {
		m.FSM.Event("stop_game")
		m.Winner = potentialWinner
		m.GuessedTime = time.Now()
		return true
	}

	return false
}

func (m *Machine) StopGame() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.FSM.Event("stop_game")
	return nil
}
