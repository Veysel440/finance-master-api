package obs

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func InitLogger() {
	lvl := strings.ToLower(os.Getenv("LOG_LEVEL"))
	level, err := zerolog.ParseLevel(lvl)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.TimeFieldFormat = time.RFC3339Nano

	if strings.EqualFold(os.Getenv("LOG_FORMAT"), "console") {
		Log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
			Level(level).With().Timestamp().Logger()
		return
	}
	Log = zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()
}
