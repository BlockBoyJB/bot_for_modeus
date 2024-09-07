package v2

import (
	"bot_for_modeus/pkg/bot"
	log "github.com/sirupsen/logrus"
)

func loggingMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(c bot.Context) error {
		u := c.Update()
		if err := next(c); err != nil {
			log.Errorf("Message from %d is handled with error: %s", u.SentFrom().ID, err)
			return err
		}
		return nil
	}
}

func errorMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(c bot.Context) error {
		if err := next(c); err != nil {
			return c.SendMessage(txtError)
		}
		return nil
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
