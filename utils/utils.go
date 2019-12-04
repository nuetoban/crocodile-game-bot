package utils

import (
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
