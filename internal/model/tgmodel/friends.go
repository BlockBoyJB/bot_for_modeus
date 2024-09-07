package tgmodel

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var ChooseFriendAction = [][]tgbotapi.InlineKeyboardButton{
	{tgbotapi.NewInlineKeyboardButtonData("Расписание на день", "day_schedule"), tgbotapi.NewInlineKeyboardButtonData("Расписание на неделю", "week_schedule")},
	{tgbotapi.NewInlineKeyboardButtonData("Удалить друга", "delete_friend")},
	{tgbotapi.NewInlineKeyboardButtonData(txtBackButton, "/choose_friend_back")},
}

func FriendsButtons(friends map[string]string) [][]tgbotapi.InlineKeyboardButton {
	// friends: ключ - personId (SubjectId), значение - ФИО
	buttons := [][]tgbotapi.InlineKeyboardButton{{tgbotapi.NewInlineKeyboardButtonData("Добавить друга", "/add_friend")}}
	return append(buttons, InlineRowButtons(friends, 1)...)
}
