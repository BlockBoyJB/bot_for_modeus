package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

// Default settings
const (
	defaultParseMode = "HTML"
	botPrefixLog     = "bot/bot"
)

// Incoming message types
const (
	OnCommand  = "COMMAND"
	OnCallback = "CALLBACK"
	OnState    = "STATE"
)

type Bot struct {
	client     *tgbotapi.BotAPI
	wg         *sync.WaitGroup
	ctx        context.Context
	parseMode  string
	routers    map[string]*Group
	middleware []MiddlewareFunc
	storage    *storage
	stop       chan interface{} // TODO chan error???
	isWebhook  bool
}

type Settings struct {
	Token     string
	IsWebhook bool
	Redis     *redis.Client
	Ctx       context.Context
}

func NewBot(s Settings, opts ...Option) (*Bot, error) {
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
		routers:   make(map[string]*Group),
		storage:   newStorage(s.Redis),
		stop:      make(chan interface{}),
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
				log.Errorf("%s/ListenAndServe error handle incoming webhook message: %s", botPrefixLog, err)
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
			go b.ProcessMessage(u)
		case <-b.stop:
			return
		}
	}
}

func (b *Bot) Shutdown() {
	b.stop <- "stop"
	b.client.StopReceivingUpdates()
	b.wg.Wait()
}

func (b *Bot) ProcessMessage(u tgbotapi.Update) {
	defer b.wg.Done()
	c := b.newContext(u)
	if u.CallbackQuery != nil {
		cb := tgbotapi.NewCallback(u.CallbackQuery.ID, "")
		if _, err := b.client.Request(cb); err != nil {
			log.Errorf("%s/ProcessMessage error answer empty callback: %s", botPrefixLog, err)
		}
		if err := b.removeMessageInlineKB(u.CallbackQuery); err != nil {
			log.Errorf("%s/ProcessMessage error remove message inline kb: %s", botPrefixLog, err)
		}
	}
	f, ok := b.handle(u)
	if ok {
		if b.middleware != nil {
			f = applyMiddleware(f, b.middleware...)
		}
		if err := f(c); err != nil {
			log.Errorf("%s/ProcessMessage error handle message: %s", botPrefixLog, err)
		}
	}
}

func (b *Bot) handle(u tgbotapi.Update) (HandlerFunc, bool) {
	var userId int64
	if u.Message != nil {
		if u.Message.IsCommand() {
			g := b.routers[OnCommand]
			f, ok := g.routes[u.Message.Text]
			if ok && g.middleware != nil {
				return applyMiddleware(f, g.middleware...), true
			}
			return f, ok
		}
		userId = u.Message.From.ID
	}
	if u.CallbackQuery != nil {
		g := b.routers[OnCallback]
		f, ok := g.routes[u.CallbackQuery.Data]
		if ok {
			if g.middleware != nil {
				return applyMiddleware(f, g.middleware...), true
			}
			return f, true
		}
		userId = u.CallbackQuery.From.ID
	}
	g := b.routers[OnState]
	state, _ := b.storage.GetState(b.ctx, userId)
	f, ok := g.routes[state]
	if ok && g.middleware != nil {
		return applyMiddleware(f, g.middleware...), true
	}
	return f, ok
}

func (b *Bot) removeMessageInlineKB(query *tgbotapi.CallbackQuery) error {
	msg := tgbotapi.NewEditMessageText(query.From.ID, query.Message.MessageID, query.Message.Text)
	if _, err := b.client.Send(msg); err != nil {
		return err
	}
	return nil
}
