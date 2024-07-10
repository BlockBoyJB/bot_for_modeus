package v2

import (
	"bot_for_modeus/pkg/bot"
	log "github.com/sirupsen/logrus"
)

func LoggingMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(c bot.Context) error {
		u := c.Update()
		if err := next(c); err != nil {
			log.Errorf("Message from %d is handled with error: %s", u.SentFrom().ID, err)
			return err
		}
		return nil
	}
}
