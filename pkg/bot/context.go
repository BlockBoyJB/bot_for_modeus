package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

type Context interface {
	Bot() *Bot
	Update() tgbotapi.Update
	Context() context.Context

	UserId() int64
	Text() string

	SendMessage(text string) error
	SendMessageWithInlineKB(text string, kb [][]tgbotapi.InlineKeyboardButton) error
	EditMessage(text string) error
	EditMessageWithInlineKB(text string, kb [][]tgbotapi.InlineKeyboardButton) error
	DeleteInlineKB() error

	SetState(state string) error
	GetState() (string, error)
	SetData(key string, v any) error
	SetTempData(key string, v any, d time.Duration) error
	GetData(key string, v any) error
	Clear() error
}

type nativeContext struct {
	bot    *Bot
	update tgbotapi.Update
}

func (b *Bot) NewContext(u tgbotapi.Update) Context {
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

func (c *nativeContext) Context() context.Context {
	return c.bot.ctx
}

func (c *nativeContext) UserId() int64 {
	if u := c.update.SentFrom(); u != nil {
		return u.ID
	}
	c.bot.logger.Printf("/Context/UserId error get user id from request")
	return 0
}

func (c *nativeContext) Text() string {
	if c.update.Message != nil {
		return c.update.Message.Text
	}
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery.Data
	}
	c.bot.logger.Printf("/Context/Text error get user input from request")
	return ""
}

func (c *nativeContext) SendMessage(text string) error {
	msg := tgbotapi.NewMessage(c.UserId(), text)
	msg.ParseMode = c.bot.parseMode
	return c.sendMessage(msg)
}

func (c *nativeContext) SendMessageWithInlineKB(text string, kb [][]tgbotapi.InlineKeyboardButton) error {
	msg := tgbotapi.NewMessage(c.UserId(), text)
	msg.ParseMode = c.bot.parseMode
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(kb...)
	return c.sendMessage(msg)
}

func (c *nativeContext) EditMessage(text string) error {
	msg := tgbotapi.NewEditMessageText(c.UserId(), c.lastMessageId(), text)
	msg.ParseMode = c.bot.parseMode
	return c.sendMessage(msg)
}

func (c *nativeContext) EditMessageWithInlineKB(text string, kb [][]tgbotapi.InlineKeyboardButton) error {
	msg := tgbotapi.NewEditMessageText(c.UserId(), c.lastMessageId(), text)
	msg.ParseMode = c.bot.parseMode

	r := tgbotapi.NewInlineKeyboardMarkup(kb...)
	msg.ReplyMarkup = &r

	return c.sendMessage(msg)
}

func (c *nativeContext) DeleteInlineKB() error {
	msg := tgbotapi.NewEditMessageText(c.UserId(), c.lastMessageId(), c.update.CallbackQuery.Message.Text)
	msg.ParseMode = c.bot.parseMode
	return c.sendMessage(msg)
}

func (c *nativeContext) sendMessage(msg tgbotapi.Chattable) error {
	if _, err := c.bot.client.Send(msg); err != nil {
		c.bot.logger.Printf("/Context/sendMessage error send message to user: %s", err)
		return err
	}
	return nil
}

func (c *nativeContext) lastMessageId() int {
	if c.update.Message != nil {
		return c.update.Message.MessageID
	}
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery.Message.MessageID
	}
	c.bot.logger.Printf("/Context/lastMessageId error find last message id from request")
	return 0
}

func (c *nativeContext) SetState(state string) error {
	return c.bot.storage.setState(c.UserId(), state)
}

func (c *nativeContext) GetState() (string, error) {
	return c.bot.storage.getState(c.UserId())
}

func (c *nativeContext) SetData(key string, v any) error {
	return c.bot.storage.setData(c.UserId(), key, v)
}

func (c *nativeContext) SetTempData(key string, v any, d time.Duration) error {
	return c.bot.storage.setTempData(c.UserId(), key, v, d)
}

func (c *nativeContext) GetData(key string, v any) error {
	return c.bot.storage.getData(c.UserId(), key, v)
}

func (c *nativeContext) Clear() error {
	return c.bot.storage.clear(c.UserId())
}
