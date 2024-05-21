package handler

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/service"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
)

const (
	txtStart = "Привет! Я умею получать расписание из модеуса и отправлять вам! Чтобы начать пользоваться ботом, введите без ошибок ваше ФИО (обязательно как указано в модеусе!) или ФИО того студента, чье расписание Вы хотите получать"
	txtHelp  = "Бот создан для получения расписания из модеуса. При нажатии команды /start, Вы вводите свое ФИО, которое будет использовать бот для поиска вас внутри системы modeus\n" +
		"Вам также предложено добавить свой логин и пароль от учетной записи, чтобы открыть возможность узнавать свои оценки /grades\n" +
		"После того как Вы ввели свой логин и пароль, у Вас не будет возможности просматривать расписание других студентов\n" +
		"Существуют следующие команды:\n\n" +
		"/day_schedule - расписание на сегодня\n" +
		"/week_schedule - расписание на всю текущую неделю\n" +
		"/grades - баллы по предметам. Требует наличие логина и пароля от пользователя\n" +
		"/settings - настройки. Здесь вы можете добавить/удалить логин и пароль от модеуса, а также изменить ФИО, если оно было введено с ошибкой\n" +
		"/other_student - посмотреть расписание (на сегодня и неделю) другого студента. Вводите ФИО студента, расписание которого хотите получить. " +
		"Информация о нем никак не сохраняется и при использовании этой команды придется вводить ФИО студента заново\n" +
		"/menu - просто главное меню =)\n" +
		"/stop - остановка бота. Бот полностью удаляет все Ваши данные из нашей системы. Чтобы снова начать пользоваться ботом, нужно будет нажать /start и заново вводить все данные"
	txtMenu = "Вы находитесь в главном меню. Вы можете:\n" +
		"Получить расписание на день: /day_schedule\n" +
		"Получить расписание на неделю: /week_schedule\n" +
		"Узнать текущие баллы: /grades\n" +
		"Перейти в настройки: /settings\n" +
		"Обратиться за помощью: /help"
	txtUserNotFound = "Ой! Мы не можем найти информацию от Вас! Пожалуйста, перезапустите бота, нажав команду /start"
	txtSettings     = "Вы находитесь в настройках.\n" +
		"Существуют следующие функции:\n" +
		"Добавить логин и пароль: открывает возможность получать текущие баллы по предметам. <b>Внимание!</b> после использования этой функции Вы не сможете просматривать расписание других студентов (вместо этого воспользуйтесь /other_student)\n" +
		"Изменить ФИО: обновляем Ваше ФИО, если Вы указали его с ошибкой. Также можно указать ФИО другого студента, чтобы всегда получать его расписание\n" +
		"Удалить логин и пароль: удаляем Ваш логин и пароль. У Вас больше не будет возможности просматривать оценки. Однако теперь вы сможете смотреть расписание других студентов (введенное ФИО при старте)"
	txtOtherStudentInput = "Введите ФИО студента, информацию о котором хотите узнать:"
	txtStop              = "Бот остановлен! Информация о пользователе удалена! Чтобы снова воспользоваться ботом, нажмите /start"
	txtParsingData       = "Пожалуйста, подождите. Данные с сайта собираются и обрабатываются..."
)

func (h *Handler) commandHandler(msg tgmodel.Message) error {
	switch msg.Text {
	case "/start":
		return h.cmdStart(msg)
	case "/help":
		return h.cmdHelp(msg)
	case "/menu":
		return h.cmdMenu(msg)
	case "/day_schedule":
		return h.cmdDaySchedule(msg)
	case "/week_schedule":
		return h.cmdWeekSchedule(msg)
	case "/grades":
		return h.cmdGrades(msg)
	case "/other_student":
		return h.cmdOtherStudent(msg)
	case "/settings":
		return h.cmdSettings(msg)
	case "/stop":
		return h.cmdStop(msg)
	}
	return nil
}

func (h *Handler) cmdStart(msg tgmodel.Message) error {
	h.storage.setState(msg.UserId, stateUserCreate)
	return h.bot.SendMessage(msg.UserId, txtStart)
}

func (h *Handler) cmdHelp(msg tgmodel.Message) error {
	h.storage.clear(msg.UserId)
	return h.bot.SendMessage(msg.UserId, txtHelp)
}

func (h *Handler) cmdMenu(msg tgmodel.Message) error {
	h.storage.clear(msg.UserId)
	return h.bot.SendMessage(msg.UserId, txtMenu)
}

func userSchedule(h *Handler, msg tgmodel.Message, p func(context.Context, int64) (string, error)) error {
	newMsg, _ := h.bot.SendMessageWithReturn(msg.UserId, txtParsingData)
	s, err := p(h.ctx, msg.UserId)
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

func (h *Handler) cmdDaySchedule(msg tgmodel.Message) error {
	return userSchedule(h, msg, h.services.DaySchedule)
}

func (h *Handler) cmdWeekSchedule(msg tgmodel.Message) error {
	return userSchedule(h, msg, h.services.WeekSchedule)
}

func (h *Handler) cmdGrades(msg tgmodel.Message) error {
	newMsg, _ := h.bot.SendMessageWithReturn(msg.UserId, txtParsingData)
	subjects, txt, err := h.services.UserGrades(h.ctx, msg.UserId)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return h.bot.SendMessage(msg.UserId, txtUserNotFound)
		}
		if errors.Is(err, service.ErrUserPermissionDenied) {
			return h.bot.SendMessage(msg.UserId, txtPermissionDenied)
		}
		return h.bot.SendMessage(msg.UserId, txtParserError)
	}
	if err = h.storage.updateData(msg.UserId, "user_subjects", subjects); err != nil {
		log.Errorf("SendUserGrades encode subject map error: %s", err)
		return h.bot.SendMessage(msg.UserId, txtUnknownError)
	}
	if err = h.storage.updateData(msg.UserId, "amount_subjects", len(subjects)); err != nil {
		log.Errorf("SendUserGrades encode subject map error: %s", err)
		return h.bot.SendMessage(msg.UserId, txtUnknownError)
	}
	_ = h.bot.DeleteMessage(msg.UserId, newMsg.MessageId)
	h.storage.setState(msg.UserId, stateChooseSubject)
	if err = h.bot.SendMessage(msg.UserId, txt); err != nil {
		return h.bot.SendMessage(msg.UserId, txtParserError)
	}
	return h.bot.SendMessageWithInlineKeyboard(msg.UserId, txtGetSubjectInfo, tgmodel.NumbersButtons(len(subjects), 4))
}

func (h *Handler) cmdSettings(msg tgmodel.Message) error {
	return h.bot.SendMessageWithInlineKeyboard(msg.UserId, txtSettings, tgmodel.SettingsButtons)
}

func (h *Handler) cmdOtherStudent(msg tgmodel.Message) error {
	h.storage.setState(msg.UserId, stateOtherStudentInput)
	return h.bot.SendMessage(msg.UserId, txtOtherStudentInput)
}

func (h *Handler) cmdStop(msg tgmodel.Message) error {
	if err := h.services.DeleteUser(h.ctx, msg.UserId); err != nil {
		return err
	}
	return h.bot.SendMessage(msg.UserId, txtStop)
}

// Альтернативная реализация работы парсера через брокер. Отказался от этого решения, тк посчитал что усложнил работу всей системы, не прибавив скорости работы
//
//func (h *Handler) _cmdDaySchedule(msg tgmodel.Message) error {
//	newMsg, _ := h.bot.SendMessageWithReturn(msg.UserId, txtParsingData)
//	return h.services.RequestDaySchedule(h.ctx, msg.UserId, newMsg.MessageId)
//}
//
//func (h *Handler) _cmdWeekSchedule(msg tgmodel.Message) error {
//	newMsg, _ := h.bot.SendMessageWithReturn(msg.UserId, txtParsingData)
//	return h.services.RequestWeekSchedule(h.ctx, msg.UserId, newMsg.MessageId)
//}
//
//func (h *Handler) _cmdGrades(msg tgmodel.Message) error {
//	newMsg, _ := h.bot.SendMessageWithReturn(msg.UserId, txtParsingData)
//	return h.services.RequestUserGrades(h.ctx, msg.UserId, newMsg.MessageId)
//}
