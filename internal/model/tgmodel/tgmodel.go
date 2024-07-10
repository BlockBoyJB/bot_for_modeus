package tgmodel

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var SettingsButtons = [][]tgbotapi.InlineKeyboardButton{
	{tgbotapi.NewInlineKeyboardButtonData("Добавить логин и пароль", "/add_login_password"), tgbotapi.NewInlineKeyboardButtonData("Изменить ФИО", "/update_full_name")},
}

var YesOrNoButtons = [][]tgbotapi.InlineKeyboardButton{
	{tgbotapi.NewInlineKeyboardButtonData("Да", "да")},
	{tgbotapi.NewInlineKeyboardButtonData("Нет", "нет")},
}

var ChooseFriendsAction = [][]tgbotapi.InlineKeyboardButton{
	{tgbotapi.NewInlineKeyboardButtonData("Расписание на день", "day_schedule")},
	{tgbotapi.NewInlineKeyboardButtonData("Расписание на неделю", "week_schedule")},
	{tgbotapi.NewInlineKeyboardButtonData("Удалить друга", "delete_friend")},
}

var OtherStudentButtons = [][]tgbotapi.InlineKeyboardButton{
	{tgbotapi.NewInlineKeyboardButtonData("Расписание на день", "day_schedule")},
	{tgbotapi.NewInlineKeyboardButtonData("Расписание на неделю", "week_schedule")},
}

var MyProfileButtons = [][]tgbotapi.InlineKeyboardButton{
	{tgbotapi.NewInlineKeyboardButtonData("Обо мне", "/about_me")},
	{tgbotapi.NewInlineKeyboardButtonData("Рейтинги", "/ratings")},
}

func NumbersButtons(k, size int) [][]tgbotapi.InlineKeyboardButton {
	// Идею реализации подсмотрел у aiogram KeyboardBuilder.adjust
	// https://docs.aiogram.dev/en/latest/utils/keyboard.html#aiogram.utils.keyboard.ReplyKeyboardBuilder.adjust
	var buttons [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	var num string
	for i := 1; i <= k; i++ {
		if len(row) >= size {
			buttons = append(buttons, row)
			row = nil
		}
		num = fmt.Sprintf("%d", i)
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(num, num))
	}
	if len(row) != 0 {
		buttons = append(buttons, row)
	}
	return buttons
}

func FriendsButtons(friends map[string]string, size int) [][]tgbotapi.InlineKeyboardButton {
	// friends: ключ - personId (SubjectId), значение - ФИО
	buttons := [][]tgbotapi.InlineKeyboardButton{{tgbotapi.NewInlineKeyboardButtonData("Добавить друга", "/add_friend")}}
	var row []tgbotapi.InlineKeyboardButton
	for personId, fullName := range friends {
		if len(row) >= size {
			buttons = append(buttons, row)
			row = nil
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(fullName, personId))
	}
	if len(row) != 0 {
		buttons = append(buttons, row)
	}
	return buttons
}

var UICommands = []tgbotapi.BotCommand{
	{Command: "start", Description: "Перезапустить бота"},
	{Command: "help", Description: "Помощь"},
	{Command: "day_schedule", Description: "Расписание на день"},
	{Command: "week_schedule", Description: "Расписание на неделю"},
	{Command: "grades", Description: "Посмотреть баллы"},
	{Command: "friends", Description: "Посмотреть расписание друзей"},
	{Command: "me", Description: "Информация обо мне"},
	{Command: "settings", Description: "Настройки"},
	{Command: "other_student", Description: "Расписание другого студента"},
	{Command: "stop", Description: "Остановить бота"},
}
