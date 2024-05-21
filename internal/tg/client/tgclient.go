package client

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/tg/handler"
	"bytes"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

const (
	clientPrefixLog  = "/tg/client/tgclient"
	defaultParseMode = "HTML"
)

type HandlerFunc func(u tgbotapi.Update, c *Client, h *handler.Handler)

func (h HandlerFunc) Run(update tgbotapi.Update, client *Client, handler *handler.Handler) {
	h(update, client, handler)
}

type Client struct {
	client  *tgbotapi.BotAPI
	handler HandlerFunc
}

func NewClient(token string, handler HandlerFunc) (*Client, error) {
	c, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	if err = setBotCommands(c, tgmodel.UICommands); err != nil {
		return nil, err
	}
	return &Client{
		client:  c,
		handler: handler,
	}, nil
}

func (c *Client) ListenAndServe(h *handler.Handler) {
	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 60
	updates := c.client.GetUpdatesChan(cfg)
	for u := range updates {
		go c.handler.Run(u, c, h)
	}
}

func ProcessingMessage(u tgbotapi.Update, c *Client, h *handler.Handler) {
	if u.Message != nil {
		if err := h.IncomingMessage(tgmodel.Message{
			UserId:    u.Message.From.ID,
			MessageId: u.Message.MessageID,
			Text:      u.Message.Text,
			IsCommand: u.Message.IsCommand(),
		}); err != nil {
			log.Errorf("%s/ProcessingMessage error handle message: %s", clientPrefixLog, err)
		}
	}
	if u.CallbackQuery != nil {
		cb := tgbotapi.NewCallback(u.CallbackQuery.ID, "") // пустой ответ на коллбэк, чтобы убрать ожидание коллбэка
		if _, err := c.client.Request(cb); err != nil {
			log.Errorf("%s/ProcessinMessage error send empty callback answer: %s", clientPrefixLog, err)
		}
		if err := c.deleteInlineKeyboard(u.CallbackQuery.From.ID, u.CallbackQuery.Message.MessageID, u.CallbackQuery.Message.Text); err != nil {
			log.Errorf("processing message error delete inline kb")
		}
		if err := h.IncomingMessage(tgmodel.Message{
			UserId:     u.CallbackQuery.From.ID,
			MessageId:  u.CallbackQuery.Message.MessageID,
			Text:       u.CallbackQuery.Data,
			IsCallback: true,
		}); err != nil {
			log.Errorf("%s/ProcessigMessage error handle callback: %s", clientPrefixLog, err)
		}
	}
}

func (c *Client) SendMessage(id int64, text string) error {
	if len(text) > 4096 {
		for _, s := range splitString(text, 4096) {
			msg := tgbotapi.NewMessage(id, s)
			msg.ParseMode = defaultParseMode
			if _, err := c.client.Send(msg); err != nil {
				return err
			}
		}
		return nil
	}
	msg := tgbotapi.NewMessage(id, text)
	msg.ParseMode = defaultParseMode
	if _, err := c.client.Send(msg); err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteMessage(chatId int64, messageId int) error {
	msg := tgbotapi.NewDeleteMessage(chatId, messageId)
	if r, err := c.client.Request(msg); err != nil || !r.Ok {
		return err
	}
	return nil
}

func (c *Client) SendMessageWithReturn(id int64, text string) (tgmodel.Message, error) {
	msg := tgbotapi.NewMessage(id, text)
	msg.ParseMode = defaultParseMode
	newMsg, err := c.client.Send(msg)
	if err != nil {
		return tgmodel.Message{}, err
	}
	return tgmodel.Message{
		UserId:    id,
		MessageId: newMsg.MessageID,
		Text:      text,
	}, nil
}

func (c *Client) SendMessageWithInlineKeyboard(id int64, text string, buttons []tgmodel.RowInlineButtons) error {
	kb := make([][]tgbotapi.InlineKeyboardButton, len(buttons))
	for i := 0; i < len(buttons); i++ {
		row := buttons[i]
		kb[i] = make([]tgbotapi.InlineKeyboardButton, len(row))
		for j := 0; j < len(row); j++ {
			btn := row[j]
			kb[i][j] = tgbotapi.NewInlineKeyboardButtonData(btn.Key, btn.Value)
		}
	}
	replyMarkup := tgbotapi.NewInlineKeyboardMarkup(kb...)
	newMsg := tgbotapi.NewMessage(id, text)
	newMsg.ReplyMarkup = replyMarkup
	newMsg.ParseMode = defaultParseMode
	if _, err := c.client.Send(newMsg); err != nil {
		return err
	}
	return nil
}

func (c *Client) deleteInlineKeyboard(id int64, msgId int, text string) error {
	msg := tgbotapi.NewEditMessageText(id, msgId, text)
	if _, err := c.client.Send(msg); err != nil {
		return err
	}
	return nil
}

func (c *Client) SendMessageWithRemoveKeyboard(id int64, text string) error {
	msg := tgbotapi.NewMessage(id, text)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	msg.ParseMode = defaultParseMode
	if _, err := c.client.Send(msg); err != nil {
		return err
	}
	return nil
}

// Функция делит строку по заданной длине подстрок. Телеграм имеет ограничение на максимальную длину сообщения,
// а именно 4096 символов. Некоторые возвращаемые с парсера строки могут доходить и до 10к символов, особенно для функции
// SubjectGradesInfo в конце семестра, где количество встреч 25+
func splitString(str string, size int) []string {
	subStr := ""
	result := make([]string, 0)
	runes := bytes.Runes([]byte(str))
	for i, r := range runes {
		subStr += string(r)
		if (i+1)%size == 0 {
			result = append(result, subStr)
			subStr = ""
		} else if (i + 1) == len(runes) {
			result = append(result, subStr)
		}
	}
	return result
}

// custom ui commands in user side menu
func setBotCommands(client *tgbotapi.BotAPI, commands []tgmodel.UIBotCommand) error {
	cmdList := make([]tgbotapi.BotCommand, len(commands))
	for i := 0; i < len(commands); i++ {
		cmdList[i] = tgbotapi.BotCommand{
			Command:     commands[i].Command,
			Description: commands[i].Description,
		}
	}
	cmd := tgbotapi.NewSetMyCommands(cmdList...)
	if _, err := client.Request(cmd); err != nil {
		return err
	}
	return nil
}
