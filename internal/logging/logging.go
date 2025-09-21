package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

var (
	once   sync.Once
	logger *slog.Logger
)

// Logger returns a singleton slog.Logger that writes both to stdout and a
// persistent log file inside the application's working directory. The
// function is safe for concurrent use.
func Logger() *slog.Logger {
	once.Do(func() {
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})

		file, err := openLogFile()
		if err != nil {
			logger = slog.New(handler)
			return
		}

		multi := io.MultiWriter(os.Stdout, file)
		handler = slog.NewTextHandler(multi, &slog.HandlerOptions{Level: slog.LevelInfo})
		logger = slog.New(handler)
	})

	return logger
}

func openLogFile() (*os.File, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(cwd, "scraper.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	return file, nil
}
