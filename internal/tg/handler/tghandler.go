package handler

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/service"
	"context"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
)

const (
	txtDefault           = "В главное меню: /menu\nПомощь: /help\nПолучить расписание на день: /day_schedule\nПолучить расписание на неделю: /week_schedule\nУзнать текущие баллы: /grades\nНастройки: /settings"
	txtUnknownMessage    = "Ой! Кажется, я не знаю такой команды...\nПомощь: /help"
	txtParserError       = "Ой! Кажется, у нас произошла ошибка при сборе данных с модеуса! Пожалуйста, попробуйте позже или узнайте расписание самостоятельно: https://utmn.modeus.org/schedule-calendar/"
	txtPermissionDenied  = "Ой! Кажется, Вы не указали логин и пароль (или указали неправильный) от модеуса, чтобы получить возможность просматривать оценки. Для этого перейдите в /settings -> добавить логин и пароль"
	txtIncorrectFullName = "Ой! Кажется, Вы указали ФИО с ошибкой! (или вас нет в системе модеус?).\nЕсли вы хотели узнать свое расписание, то перейдите в настройки /settings -> Изменить ФИО"
	txtGetSubjectInfo    = "Чтобы посмотреть баллы за каждую пройденную встречу, нажмите на одну из кнопок:"
)

var (
	stateUserCreate         = "stateUserCreate"
	stateAddLoginPassword   = "stateAddLoginPassword"
	stateUpdateFullName     = "stateUpdateFullName"
	stateChooseSubject      = "stateChooseSubject"
	stateActionAfterCreate  = "stateActionAfterCreate"
	stateOtherStudentInput  = "stateOtherStudentInput"
	stateOtherStudentAction = "stateOtherStudentAction"
)

type MessageSender interface {
	SendMessage(id int64, text string) error
	DeleteMessage(chatId int64, messageId int) error
	SendMessageWithReturn(id int64, text string) (tgmodel.Message, error)
	SendMessageWithRemoveKeyboard(id int64, text string) error
	SendMessageWithInlineKeyboard(id int64, text string, buttons []tgmodel.RowInlineButtons) error
}

type Handler struct {
	ctx      context.Context
	bot      MessageSender
	services *service.Services
	storage  *storage
}

func NewHandler(ctx context.Context, bot MessageSender, services *service.Services) *Handler {
	return &Handler{
		ctx:      ctx,
		bot:      bot,
		services: services,
		storage:  newStorage(),
	}
}

func (h *Handler) IncomingMessage(msg tgmodel.Message) error {
	if msg.IsCommand {
		// Work with commands
		return h.commandHandler(msg)
	} else if msg.IsCallback {
		// callback
		return h.callbackHandler(msg)
	} else { // it can be a message with state or just incorrect command
		state := h.storage.getState(msg.UserId)
		if len(state) != 0 {
			// state handle
			return h.stateHandler(state, msg)
		} else {
			// unknown message
			return h.unknownMessage(msg)
		}
	}
}

func (h *Handler) unknownMessage(msg tgmodel.Message) error {
	return h.bot.SendMessage(msg.UserId, txtUnknownMessage)
}

// Эти функции вызывает broker consumer. На данный момент не используются,
// тк я посчитал их работу (+реализацию) ненужным усложнением всего проекта в целом

func (h *Handler) SendDaySchedule(userId int64, messageId int) error {
	s, err := h.services.DaySchedule(h.ctx, userId)
	if err != nil {
		if errors.Is(err, service.ErrUserIncorrectFullName) {
			return h.bot.SendMessage(userId, txtIncorrectFullName)
		}
		if errors.Is(err, service.ErrUserNotFound) {
			return h.bot.SendMessage(userId, txtUserNotFound)
		}
		if errors.Is(err, service.ErrUserPermissionDenied) {
			return h.bot.SendMessage(userId, txtPermissionDenied)
		}
		return h.bot.SendMessage(userId, txtParserError)
	}
	_ = h.bot.DeleteMessage(userId, messageId)
	return h.bot.SendMessage(userId, s)
}

func (h *Handler) SendWeekSchedule(userId int64, messageId int) error {
	s, err := h.services.WeekSchedule(h.ctx, userId)
	if err != nil {
		if errors.Is(err, service.ErrUserIncorrectFullName) {
			return h.bot.SendMessage(userId, txtIncorrectFullName)
		}
		if errors.Is(err, service.ErrUserNotFound) {
			return h.bot.SendMessage(userId, txtUserNotFound)
		}
		if errors.Is(err, service.ErrUserPermissionDenied) {
			return h.bot.SendMessage(userId, txtPermissionDenied)
		}
		return h.bot.SendMessage(userId, txtParserError)
	}
	_ = h.bot.DeleteMessage(userId, messageId)
	return h.bot.SendMessage(userId, s)
}

func (h *Handler) SendUserGrades(userId int64, messageId int) error {
	subjects, txt, err := h.services.UserGrades(h.ctx, userId)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return h.bot.SendMessage(userId, txtUserNotFound)
		}
		if errors.Is(err, service.ErrUserPermissionDenied) {
			return h.bot.SendMessage(userId, txtPermissionDenied)
		}
		return h.bot.SendMessage(userId, txtParserError)
	}
	if err = h.storage.updateData(userId, "user_subjects", subjects); err != nil {
		log.Errorf("SendUserGrades encode subject map error: %s", err)
		return h.bot.SendMessage(userId, txtUnknownError)
	}
	if err = h.storage.updateData(userId, "amount_subjects", len(subjects)); err != nil {
		log.Errorf("SendUserGrades encode subject map error: %s", err)
		return h.bot.SendMessage(userId, txtUnknownError)
	}

	_ = h.bot.DeleteMessage(userId, messageId)
	h.storage.setState(userId, stateChooseSubject)
	if err = h.bot.SendMessage(userId, txt); err != nil {
		return h.bot.SendMessage(userId, txtParserError)
	}
	return h.bot.SendMessageWithInlineKeyboard(userId, txtGetSubjectInfo, tgmodel.NumbersButtons(len(subjects), 4))
}

func (h *Handler) SendSubjectGradesInfo(userId int64, messageId int) error {
	b := h.storage.getData(userId, "subject")
	var subjectIndex int
	if err := json.Unmarshal(b, &subjectIndex); err != nil {
		log.Errorf("SendSubjectGradesInfo encode subject map error: %s", err)
		return h.bot.SendMessage(userId, txtUnknownError)
	}
	b = h.storage.getData(userId, "amount_subjects")
	var amountSubjects int
	if err := json.Unmarshal(b, &amountSubjects); err != nil {
		log.Errorf("SendSubjectGradesInfo encode subject map error: %s", err)
		return h.bot.SendMessage(userId, txtUnknownError)
	}
	txt, err := h.services.SubjectGradesInfo(h.ctx, userId, subjectIndex)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return h.bot.SendMessage(userId, txtUserNotFound)
		}
		if errors.Is(err, service.ErrUserPermissionDenied) {
			return h.bot.SendMessage(userId, txtPermissionDenied)
		}
		return h.bot.SendMessage(userId, txtParserError)
	}
	_ = h.bot.DeleteMessage(userId, messageId)
	if err = h.bot.SendMessage(userId, txt); err != nil {
		return h.bot.SendMessage(userId, txtParserError)
	}
	return h.bot.SendMessageWithInlineKeyboard(userId, txtGetSubjectInfo, tgmodel.NumbersButtons(amountSubjects, 4))
}
