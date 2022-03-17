package logger

import (
	"sync"

	"go.uber.org/zap"
)

var (
	st      Logger = &noop{}
	stGuard sync.Once
	stMux   sync.RWMutex
)

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
}

func Init(log Logger) {
	stGuard.Do(func() {
		stMux.Lock()
		st = log
		defer stMux.Unlock()
	})
}

func Instance() Logger {
	stMux.RLock()
	defer stMux.RUnlock()

	return st
}
