package utils

import (
	"log/slog"
	"os"
	"sync"
)

var lock = &sync.Mutex{}
var logger *slog.Logger

func GetLogger() *slog.Logger {
	if logger == nil {
		lock.Lock()
		defer lock.Unlock()
		if logger == nil {
			logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
		}
	}

	return logger
}
