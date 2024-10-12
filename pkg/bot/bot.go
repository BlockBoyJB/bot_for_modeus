package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"os"
	"sync"
)

// Default settings
const (
	defaultParseMode = "HTML"
)

type Logger interface {
	Printf(format string, args ...any)
}

type Bot struct {
	client     *tgbotapi.BotAPI
	wg         *sync.WaitGroup
	ctx        context.Context
	parseMode  string
	routers    router
	middleware []MiddlewareFunc
	storage    storage
	logger     Logger
	stop       chan bool
	isWebhook  bool
}

type Settings struct {
	Token     string
	IsWebhook bool
	Ctx       context.Context
}

func NewBot(s *Settings, opts ...Option) (*Bot, error) {
	client, err := tgbotapi.NewBotAPI(s.Token)
	if err != nil {
		return nil, err
	}
	if s.Ctx == nil {
		s.Ctx = context.Background()
	}
	b := &Bot{
		client:    client,
		wg:        new(sync.WaitGroup),
		ctx:       s.Ctx,
		parseMode: defaultParseMode,
		routers:   newRouter(),
		storage:   newMemoryStorage(),
		logger:    log.New(os.Stdout, "/bot", 4),
		stop:      make(chan bool),
		isWebhook: s.IsWebhook,
	}
	for _, option := range opts {
		if err = option(b); err != nil {
			return nil, err
		}
	}
	return b, nil
}

func (b *Bot) ListenAndServe() {
	var updates tgbotapi.UpdatesChannel
	if b.isWebhook {
		updates = b.client.ListenForWebhook("/")
		go func() {
			if err := http.ListenAndServe("0.0.0.0:8000", nil); err != nil {
				b.logger.Printf("/ListenAndService error handle request: %s", err)
			}
		}()
	} else {
		config := tgbotapi.NewUpdate(0)
		config.Timeout = 60
		updates = b.client.GetUpdatesChan(config)
	}
	for {
		select {
		case u := <-updates:
			b.wg.Add(1)
			go b.processMessage(u)
		case <-b.stop:
			return
		}
	}
}

func (b *Bot) processMessage(u tgbotapi.Update) {
	defer b.wg.Done()
	c := b.NewContext(u)

	// на коллбэк (нажатие инлайн кнопки) нужно ответить пустым, чтобы убрать анимацию "ожидания" на кнопке
	if u.CallbackQuery != nil {
		b.answerEmptyCallback(u.CallbackQuery)
	}

	f, ok := b.handle(c, u)
	if !ok {
		return
	}
	if b.middleware != nil {
		f = applyMiddleware(f, b.middleware...)
	}
	if err := f(c); err != nil {
		b.logger.Printf("/processMessage handle message func err: %s", err)
	}
}

// Ищем нужную ручку для обработки...
func (b *Bot) handle(c Context, u tgbotapi.Update) (HandlerFunc, bool) {
	if u.Message != nil {
		if u.Message.IsCommand() {
			f, ok := b.routers[OnCommand].find(c, u.Message.Text)
			return f, ok
		}
		// Может быть нажатие с обычной клавиатуры
		if f, ok := b.routers[OnMessage].find(c, u.Message.Text); ok {
			return f, ok
		}
	}
	if u.CallbackQuery != nil {
		if f, ok := b.routers[OnCallback].find(c, u.CallbackQuery.Data); ok {
			return f, ok
		}
	}
	// Если условия выше ничего не вернули, значит это либо обычное сообщение от пользователя (не /команда) (попросили его что-то ввести),
	// либо в инлайн кнопке на коллбэк есть какое-то значение, которое надо обработать отдельно от ручки коллбэков.
	// Соответственно, при таких вариантах это какое-то состояние пользователя
	state, _ := b.storage.getState(c.UserId())
	f, ok := b.routers[OnState].find(c, state)
	return f, ok
}

func (b *Bot) Shutdown() {
	b.stop <- true
	b.client.StopReceivingUpdates()
	b.wg.Wait()
}

func (b *Bot) answerEmptyCallback(c *tgbotapi.CallbackQuery) {
	cb := tgbotapi.NewCallback(c.ID, "")
	_, _ = b.client.Request(cb)
}
