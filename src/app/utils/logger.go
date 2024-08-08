package utils

import (
	"log/slog"
	"os"
	"sync"
)

var logger *slog.Logger

func GetLogger() *slog.Logger {
	if logger == nil {
		logger = sync.OnceValue(func() *slog.Logger {
			return slog.New(slog.NewJSONHandler(os.Stdout, nil))
		})()
	}

	return logger
}
