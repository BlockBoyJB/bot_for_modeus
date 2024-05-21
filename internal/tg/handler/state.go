package handler

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/service"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

const (
	txtUnknownError = "Ой! Кажется, у нас произошла непредвиденная ошибка. Попробуйте воспользоваться функцией позже"
	txtUserCreated  = "Пользователь успешно создан!\n" +
		"Хотите добавить логин и пароль от своей учетной записи модеус?\n" +
		"Вы получите возможность просматривать раздел оценок. Вы можете нажать кнопку \"Нет\", если не хотите указывать\n" +
		"Вам так же будет доступна возможность смотреть расписание, однако раздел оценок будет недоступен."
	txtUserCreateError        = "Ой! Произошла ошибка при создании пользователя! Пожалуйста, попробуйте позже"
	txtLoginPasswordCreated   = "Логин и пароль успешно сохранены\n" + txtDefault
	txtIncorrectLoginPassword = "Ой! Кажется, Вы ввели логин и пароль с ошибкой! Пожалуйста, введите через пробел сначала логин, потом пароль"
	txtFullNameUpdated        = "ФИО успешно обновлено\n" + txtDefault
	txtIncorrectButton        = "Ой! Такого предмета нет! Пожалуйста, нажмите на кнопку из доступных"
)

func (h *Handler) stateHandler(state string, msg tgmodel.Message) error {
	switch state {
	case stateUserCreate:
		return h.stateCreateUser(msg)
	case stateAddLoginPassword:
		return h.stateAddLoginPassword(msg)
	case stateUpdateFullName:
		return h.stateUpdateFullName(msg)
	case stateChooseSubject:
		return h.stateChooseSubject(msg)
	case stateActionAfterCreate:
		return h.stateActionAfterCreate(msg)
	case stateOtherStudentInput:
		return h.stateOtherStudentInput(msg)
	case stateOtherStudentAction:
		return h.stateOtherStudentAction(msg)
	}
	return nil
}

func (h *Handler) stateCreateUser(msg tgmodel.Message) error {
	err := h.services.CreateUser(h.ctx, msg.UserId, msg.Text)
	if err != nil {
		if !errors.Is(err, service.ErrUserAlreadyExists) {
			return h.bot.SendMessage(msg.UserId, txtUserCreateError)
		}
	}
	h.storage.setState(msg.UserId, stateActionAfterCreate)
	return h.bot.SendMessageWithInlineKeyboard(msg.UserId, txtUserCreated, tgmodel.YesOrNoButtons)
}

// Сюда пользователь попадает после нажатия кнопки да/нет после создания аккаунта. Текстом сообщения будет да/нет
func (h *Handler) stateActionAfterCreate(msg tgmodel.Message) error {
	if msg.Text == "да" {
		h.storage.setState(msg.UserId, stateAddLoginPassword)
		return h.bot.SendMessage(msg.UserId, txtAddLoginPassword)
	}
	return h.bot.SendMessage(msg.UserId, "Пользователь успешно создан\n"+txtDefault)
}

func (h *Handler) stateAddLoginPassword(msg tgmodel.Message) error {
	data := strings.Fields(msg.Text)
	if len(data) != 2 {
		return h.bot.SendMessage(msg.UserId, txtIncorrectLoginPassword)
	}
	defer h.storage.clear(msg.UserId)
	login, password := data[0], data[1]
	err := h.services.AddUserLoginPassword(h.ctx, msg.UserId, login, password)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return h.bot.SendMessage(msg.UserId, txtUserNotFound)
		}
		return h.bot.SendMessage(msg.UserId, txtUnknownError)
	}
	return h.bot.SendMessage(msg.UserId, txtLoginPasswordCreated)
}

func (h *Handler) stateUpdateFullName(msg tgmodel.Message) error {
	defer h.storage.clear(msg.UserId)
	err := h.services.UpdateUserFullName(h.ctx, msg.UserId, msg.Text)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return h.bot.SendMessage(msg.UserId, txtUserNotFound)
		}
		return h.bot.SendMessage(msg.UserId, txtUnknownError)
	}
	return h.bot.SendMessage(msg.UserId, txtFullNameUpdated)
}

// Реализация работы через брокер
//func (h *Handler) _stateChooseSubject(msg tgmodel.Message) error {
//	data := h.storage.getData(msg.UserId, "user_subjects")
//	num, err := strconv.Atoi(msg.Text)
//	if err != nil {
//		return h.bot.SendMessage(msg.UserId, txtUnknownError)
//	}
//	var subjects map[int]int
//	err = json.Unmarshal(data, &subjects)
//	if err != nil {
//		return h.bot.SendMessage(msg.UserId, txtUnknownError)
//	}
//	subject, ok := subjects[num]
//	if !ok {
//		return h.bot.SendMessage(msg.UserId, txtIncorrectButton)
//	}
//
//	if err = h.storage.updateData(msg.UserId, "subject", subject); err != nil {
//		return h.bot.SendMessage(msg.UserId, txtUnknownError)
//	}
//	//h.storage.clear(msg.UserId)
//	newMsg, _ := h.bot.SendMessageWithReturn(msg.UserId, txtParsingData)
//	return h.services.RequestSubjectInfo(h.ctx, msg.UserId, newMsg.MessageId)
//}

// Входящее сообщение - одна из цифр с кнопки
func (h *Handler) stateChooseSubject(msg tgmodel.Message) error {
	num, err := strconv.Atoi(msg.Text)
	if err != nil {
		log.Errorf("stateChooseSubject error convert string to int: %s", err)
		return h.bot.SendMessage(msg.UserId, txtUnknownError)
	}
	var subjects map[int]int // ключ - нажатая пользователем кнопка, значение - индекс предмета в модеусе
	if err = h.storage._getData(msg.UserId, "user_subjects", &subjects); err != nil {
		log.Errorf("stateChooseSubject error get user subjects: %s", err)
		return h.bot.SendMessage(msg.UserId, txtUnknownError)
	}
	subjectIndex, ok := subjects[num]
	if !ok {
		return h.bot.SendMessage(msg.UserId, txtIncorrectButton)
	}
	var amountSubjects int
	if err = h.storage._getData(msg.UserId, "amount_subjects", &amountSubjects); err != nil {
		log.Errorf("stateChooseSubject error get amount of subjects: %s", err)
		return h.bot.SendMessage(msg.UserId, txtUnknownError)
	}
	newMsg, _ := h.bot.SendMessageWithReturn(msg.UserId, txtParsingData)
	txt, err := h.services.SubjectGradesInfo(h.ctx, msg.UserId, subjectIndex)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return h.bot.SendMessage(msg.UserId, txtUserNotFound)
		}
		if errors.Is(err, service.ErrUserPermissionDenied) {
			return h.bot.SendMessage(msg.UserId, txtPermissionDenied)
		}
		return h.bot.SendMessage(msg.UserId, txtParserError)
	}
	_ = h.bot.DeleteMessage(msg.UserId, newMsg.MessageId)
	if err = h.bot.SendMessage(msg.UserId, txt); err != nil {
		return h.bot.SendMessage(msg.UserId, txtParserError)
	}
	return h.bot.SendMessageWithInlineKeyboard(msg.UserId, txtGetSubjectInfo, tgmodel.NumbersButtons(amountSubjects, 4))
}

func (h *Handler) stateOtherStudentInput(msg tgmodel.Message) error {
	if err := h.storage.updateData(msg.UserId, "student", msg.Text); err != nil {
		return h.bot.SendMessage(msg.UserId, txtUnknownError)
	}
	h.storage.setState(msg.UserId, stateOtherStudentAction)
	return h.bot.SendMessageWithInlineKeyboard(msg.UserId, "Выберите действие:", tgmodel.OtherStudentButtons)
}

func (h *Handler) stateOtherStudentAction(msg tgmodel.Message) error {
	data := h.storage.getData(msg.UserId, "student")
	var student string
	if err := json.Unmarshal(data, &student); err != nil {
		return h.bot.SendMessage(msg.UserId, txtUnknownError)
	}
	switch msg.Text {
	case "day_schedule":
		return otherStudentAction(h, msg, student, h.services.OtherStudentDaySchedule)
	case "week_schedule":
		return otherStudentAction(h, msg, student, h.services.OtherStudentWeekSchedule)
	}
	return nil
}

func otherStudentAction(h *Handler, msg tgmodel.Message, fullName string, p func(string) (string, error)) error {
	newMsg, _ := h.bot.SendMessageWithReturn(msg.UserId, txtParsingData)
	s, err := p(fullName)
	if err != nil {
		if errors.Is(err, service.ErrUserIncorrectFullName) {
			return h.bot.SendMessage(msg.UserId, txtIncorrectFullName)
		}
		if errors.Is(err, service.ErrUserNotFound) {
			return h.bot.SendMessage(msg.UserId, txtUserNotFound)
		}
		if errors.Is(err, service.ErrUserPermissionDenied) {
			return h.bot.SendMessage(msg.UserId, txtPermissionDenied)
		}
		return h.bot.SendMessage(msg.UserId, txtParserError)
	}
	_ = h.bot.DeleteMessage(msg.UserId, newMsg.MessageId)
	return h.bot.SendMessage(msg.UserId, s)
}
