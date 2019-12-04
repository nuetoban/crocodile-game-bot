package crocodile

import (
	"github.com/looplab/fsm"
)

// Game stores methods to manage one game
type Game interface {
	StartNewGameAndReturnWord() (string, error)
	SetNewRandomWord() (string, error)
	GetWord() string
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
	ChatID        int64
	Word          string
	FSM           *fsm.FSM
	Storage       Storage
	WordsProvider WordsProvider
}

// MachineFabric aims to produce new machines with freezed Storage and WordsProvider
type MachineFabric struct {
	Storage       Storage
	WordsProvider WordsProvider
}

// NewMachine returns Machine with freezed Storage and WordsProvider
func (m *MachineFabric) NewMachine(chatID int64) *Machine {
	return NewMachine(m.Storage, m.WordsProvider, chatID)
}

// NewMachineFabric returns MachineFabric
func NewMachineFabric(storage Storage, wp WordsProvider) *MachineFabric {
	return &MachineFabric{
		Storage:       storage,
		WordsProvider: wp,
	}
}

// NewMachine returns new Machine instance
func NewMachine(storage Storage, wp WordsProvider, chatID int64) *Machine {
	fsm := fsm.NewFSM("init", fsm.Events{}, fsm.Callbacks{})

	return &Machine{
		ChatID:        chatID,
		FSM:           fsm,
		Storage:       storage,
		WordsProvider: wp,
	}
}

// StartNewGameAndReturnWord sets m.Word to new words and returns it
func (m *Machine) StartNewGameAndReturnWord() (string, error) {
	var err error
	m.Word, err = m.WordsProvider.GetWord()
	if err != nil {
		return "", err
	}

	return m.Word, nil
}

// SetNewRandomWord generates new word
func (m *Machine) SetNewRandomWord() (string, error) {
	return "", nil
}

// GetWord returns m.Word
func (m *Machine) GetWord() string {
	return ""
}

// CheckWord checks if m.Word == provided word
func (m *Machine) CheckWord(word string) bool {
	return false
}
