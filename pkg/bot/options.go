package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Option func(bot *Bot) error

func ParseMode(mode string) Option {
	return func(bot *Bot) error {
		bot.parseMode = mode
		return nil
	}
}

func SetCommands(commands []tgbotapi.BotCommand) Option {
	return func(bot *Bot) error {
		cmd := tgbotapi.NewSetMyCommands(commands...)
		if _, err := bot.client.Request(cmd); err != nil {
			return err
		}
		return nil
	}
}
