package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

const (
	contextPrefixLog = "bot/context"
)

type Context interface {
	Bot() *Bot
	Update() tgbotapi.Update
	UserId() int64
	Text() string
	Context() context.Context
	SendMessage(text string) error
	SendMessageWithInlineKB(text string, kb [][]tgbotapi.InlineKeyboardButton) error

	SetState(state string) error
	GetState() (string, error)
	UpdateData(key string, value interface{}) error
	GetData(key string, v interface{}) error
	Clear() error
}

type nativeContext struct {
	bot    *Bot
	update tgbotapi.Update
}

func (b *Bot) newContext(u tgbotapi.Update) Context {
	return &nativeContext{
		bot:    b,
		update: u,
	}
}

func (c *nativeContext) Bot() *Bot {
	return c.bot
}

func (c *nativeContext) Update() tgbotapi.Update {
	return c.update
}

func (c *nativeContext) UserId() int64 {
	if c.update.Message != nil {
		return c.update.Message.From.ID
	}
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery.From.ID
	}
	log.Errorf("%s/UserId error get user id from incoming update", contextPrefixLog)
	return 0 // TODO panic???
}

func (c *nativeContext) Text() string {
	if c.update.Message != nil {
		return c.update.Message.Text
	}
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery.Data
	}
	log.Errorf("%s/Text error get user message from incoming update", contextPrefixLog)
	return "" // TODO panic???
}

func (c *nativeContext) Context() context.Context {
	return c.bot.ctx
}

func (c *nativeContext) SendMessage(text string) error {
	msg := tgbotapi.NewMessage(c.UserId(), text)
	msg.ParseMode = c.bot.parseMode
	if _, err := c.bot.client.Send(msg); err != nil {
		log.Errorf("%s/SendMessage error send message: %s", contextPrefixLog, err)
		return err
	}
	return nil
}

func (c *nativeContext) SendMessageWithInlineKB(text string, kb [][]tgbotapi.InlineKeyboardButton) error {
	msg := tgbotapi.NewMessage(c.UserId(), text)
	msg.ParseMode = c.bot.parseMode
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(kb...)
	if _, err := c.bot.client.Send(msg); err != nil {
		log.Errorf("%s/SendMessageWithInlineKB error send message with inline kb: %s", contextPrefixLog, err)
		return err
	}
	return nil
}

func (c *nativeContext) SetState(state string) error {
	return c.bot.storage.SetState(c.Context(), c.UserId(), state)
}

func (c *nativeContext) GetState() (string, error) {
	return c.bot.storage.GetState(c.Context(), c.UserId())
}

func (c *nativeContext) UpdateData(key string, value interface{}) error {
	return c.bot.storage.UpdateData(c.Context(), c.UserId(), key, value)
}

func (c *nativeContext) GetData(key string, v interface{}) error {
	return c.bot.storage.GetData(c.Context(), c.UserId(), key, v)
}

func (c *nativeContext) Clear() error {
	return c.bot.storage.Clear(c.Context(), c.UserId())
}
