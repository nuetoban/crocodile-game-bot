package storage

import (
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Print(...interface{})
}

type logger struct {
	log *logrus.Logger
}

func (l *logger) Print(v ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			l.log.Error("Recovered in ", r)
		}
	}()

	if v[0] == "sql" {
		l.log.Debugf("Database event: type: \"%v\", pos: \"%v\", duration: \"%v\", query: \"%v\"", v[0], v[1], v[2], v[3])
	} else {
		l.log.Errorf("Database error: type: \"%v\", pos: \"%v\", error: \"%v\"", v[0], v[1], v[2])
	}
}

func WrapLogrus(log *logrus.Logger) *logger {
	return &logger{log: log}
}
