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
