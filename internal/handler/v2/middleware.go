package v2

import (
	"bot_for_modeus/internal/metrics"
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func loggingMiddleware() bot.MiddlewareFunc {
	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()

	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(c bot.Context) error {
			start := time.Now()
			err := next(c)
			logger.Err(err).Int64("user_id", c.UserId()).Float64("duration", time.Since(start).Seconds()).Msg("update from user")
			if err != nil {
				return c.SendMessage(txtError)
			}
			return nil
		}
	}
}

func errorMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(c bot.Context) error {
		err := next(c)
		switch {
		case errors.Is(err, ErrIncorrectInput):
			return c.SendMessage(txtWarn)

		case errors.Is(err, parser.ErrModeusUnavailable):
			return c.SendMessageWithInlineKB(txtModeusUnavailable, tgmodel.ScheduleLink)

		case errors.Is(err, parser.ErrIncorrectLoginPassword):
			return c.SendMessage(txtIncorrectLoginPass)

		case errors.Is(err, parser.ErrStudentsNotFound):
			return c.SendMessage(fmt.Sprintf(txtStudentNotFound, c.Text()))

		case errors.Is(err, service.ErrUserNotFound):
			return c.SendMessage(txtUserNotFound)
		}
		return err
	}
}

// Panic-recovery мидлварь, чтобы в случае непредвиденной ошибки бот не падал, а писал лог
func recoverMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(c bot.Context) error {
		defer func() {
			if r := recover(); r != nil {
				log.Error().Interface("recover", r).Msg("[PANIC RECOVER]")
			}
		}()
		return next(c)
	}
}

func metricsMiddleware(t string) bot.MiddlewareFunc {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(c bot.Context) error {
			start := time.Now()
			defer func() {
				metrics.RequestDuration(t, time.Since(start))
				metrics.RequestTotal(t)
			}()
			if err := next(c); err != nil {
				metrics.ErrorsTotal(t)
				return err
			}
			return nil
		}
	}
}
