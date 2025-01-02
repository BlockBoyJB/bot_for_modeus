package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
)

func loggingMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(c bot.Context) error {
		u := c.Update()
		if err := next(c); err != nil {
			log.Errorf("Message from %d is handled with error: %s", u.SentFrom().ID, err)
			return c.SendMessage(txtError)
		}
		return nil
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
				log.Warnf("[PANIC RECOVER] %s", r)
			}
		}()
		return next(c)
	}
}
