package crocodile

import (
	"io"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
)

// WordsProviderReader takes content from reader, converts to string, splits by "\n" and returns random word
type WordsProviderReader struct {
	wordsList []string
}

// NewWordsProviderReader returns new instance of WordsProviderReader
func NewWordsProviderReader(r io.Reader) (*WordsProviderReader, error) {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	contentString := strings.TrimSpace(string(content))

	return &WordsProviderReader{
		wordsList: strings.Split(contentString, "\n"),
	}, nil
}

// GetWord returns random word
func (w *WordsProviderReader) GetWord() (string, error) {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(w.wordsList))
	return strings.TrimSpace(w.wordsList[index]), nil
}
