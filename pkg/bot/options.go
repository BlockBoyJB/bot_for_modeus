package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

type Option func(bot *Bot) error

func SetCommands(commands []tgbotapi.BotCommand) Option {
	return func(bot *Bot) error {
		cmd := tgbotapi.NewSetMyCommands(commands...)
		if _, err := bot.client.Request(cmd); err != nil {
			return err
		}
		return nil
	}
}

func RedisStorage(ctx context.Context, redis *redis.Client) Option {
	return func(bot *Bot) error {
		bot.storage = newRedisStorage(ctx, redis)
		return nil
	}
}

func SetLogger(logger Logger) Option {
	return func(bot *Bot) error {
		bot.logger = logger
		return nil
	}
}
