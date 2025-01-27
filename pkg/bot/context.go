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

	// Param возвращает значение параметра в маршруте
	Param(name string) string

	SendMessage(text string) error
	SendMessageWithInlineKB(text string, kb [][]tgbotapi.InlineKeyboardButton) error
	SendMessageWithReplyKB(text string, kb [][]tgbotapi.KeyboardButton) error

	EditMessage(text string) error
	EditMessageWithInlineKB(text string, kb [][]tgbotapi.InlineKeyboardButton) error

	DeleteLastMessage() error
	DeleteInlineKB() error

	SetState(state string) error
	GetState() (string, error)

	SetData(key string, v any) error
	SetTempData(key string, v any, d time.Duration) error
	SetCommonData(key string, v any, d time.Duration) error

	GetData(key string, v any) error
	GetCommonData(key string, v any) error

	DelData(keys ...string) error
	DelCommonData(keys ...string) error
	Clear() error
}

type nativeContext struct {
	bot    *Bot
	update tgbotapi.Update
	params map[string]string
}

func (b *Bot) NewContext(u tgbotapi.Update) Context {
	return &nativeContext{
		bot:    b,
		update: u,
		params: map[string]string{},
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

func (c *nativeContext) setParam(name, param string) {
	c.params[name] = param
}

func (c *nativeContext) Param(name string) string {
	return c.params[name]
}

func (c *nativeContext) SendMessage(text string) error {
	msg := tgbotapi.NewMessage(c.UserId(), text)
	msg.ParseMode = c.bot.parseMode
	return c.request(msg)
}

func (c *nativeContext) SendMessageWithInlineKB(text string, kb [][]tgbotapi.InlineKeyboardButton) error {
	msg := tgbotapi.NewMessage(c.UserId(), text)
	msg.ParseMode = c.bot.parseMode
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(kb...)
	return c.request(msg)
}

func (c *nativeContext) SendMessageWithReplyKB(text string, kb [][]tgbotapi.KeyboardButton) error {
	msg := tgbotapi.NewMessage(c.UserId(), text)
	msg.ParseMode = c.bot.parseMode
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(kb...)
	return c.request(msg)
}

func (c *nativeContext) EditMessage(text string) error {
	msg := tgbotapi.NewEditMessageText(c.UserId(), c.lastMessageId(), text)
	msg.ParseMode = c.bot.parseMode
	return c.request(msg)
}

func (c *nativeContext) EditMessageWithInlineKB(text string, kb [][]tgbotapi.InlineKeyboardButton) error {
	msg := tgbotapi.NewEditMessageText(c.UserId(), c.lastMessageId(), text)
	msg.ParseMode = c.bot.parseMode

	r := tgbotapi.NewInlineKeyboardMarkup(kb...)
	msg.ReplyMarkup = &r

	return c.request(msg)
}

func (c *nativeContext) DeleteLastMessage() error {
	msg := tgbotapi.NewDeleteMessage(c.UserId(), c.lastMessageId())
	return c.request(msg)
}

func (c *nativeContext) DeleteInlineKB() error {
	msg := tgbotapi.NewEditMessageText(c.UserId(), c.lastMessageId(), c.update.CallbackQuery.Message.Text)
	msg.ParseMode = c.bot.parseMode
	return c.request(msg)
}

func (c *nativeContext) request(msg tgbotapi.Chattable) error {
	if _, err := c.bot.client.Request(msg); err != nil {
		c.bot.logger.Printf("/Context/request error send message to user: %s", err)
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

func (c *nativeContext) SetCommonData(key string, v any, d time.Duration) error {
	return c.bot.storage.setCommonData(key, v, d)
}

func (c *nativeContext) GetData(key string, v any) error {
	return c.bot.storage.getData(c.UserId(), key, v)
}

func (c *nativeContext) GetCommonData(key string, v any) error {
	return c.bot.storage.getCommonData(key, v)
}

func (c *nativeContext) DelData(keys ...string) error {
	return c.bot.storage.delData(c.UserId(), keys...)
}

func (c *nativeContext) DelCommonData(keys ...string) error {
	return c.bot.storage.delCommonData(keys...)
}

func (c *nativeContext) Clear() error {
	return c.bot.storage.clear(c.UserId())
}
