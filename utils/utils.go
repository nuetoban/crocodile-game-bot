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

package utils

import (
	"math"
	"time"
)

func CalculateTimeDiff(t1, t2 time.Time) (h, m, s int) {
	hs := t1.Sub(t2).Hours()
	hs, mf := math.Modf(hs)
	ms := mf * 60

	ms, sf := math.Modf(ms)
	ss := sf * 60

	h = int(hs)
	m = int(ms)
	s = int(ss)

	return
}

func DetectCaseAnswers(i int) string {
	i %= 100
	if 11 <= i && i <= 19 {
		return "ответов"
	}

	i %= 10
	switch i {
	case 0, 5, 6, 7, 8, 9:
		return "ответов"
	case 1:
		return "ответ"
	case 2, 3, 4:
		return "ответа"
	}
	return "ответов"
}

func DetectCaseForGames(i int) string {
	i %= 100
	if 11 <= i && i <= 19 {
		return "игр"
	}

	i %= 10
	switch i {
	case 0, 5, 6, 7, 8, 9:
		return "игр"
	case 1:
		return "игра"
	case 2, 3, 4:
		return "игры"
	}
	return "игр"
}
