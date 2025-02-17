package bot

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
)

func setLogger(level, output string) *zerolog.Logger {
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.DebugLevel
	}
	var out io.Writer
	if output == "stdout" {
		out = os.Stdout
	} else {
		file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			log.Fatal().Err(err).Msg("setup logger error")
		}
		out = file
	}
	log.Logger = zerolog.New(out).Level(logLevel).With().Timestamp().Logger()
	return &log.Logger
}
