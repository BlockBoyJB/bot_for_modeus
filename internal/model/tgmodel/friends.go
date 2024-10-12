package tgmodel

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

func ChooseFriendAction(scheduleId string) [][]tgbotapi.InlineKeyboardButton {
	now := time.Now()
	return [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("Расписание на день", formatScheduleButtonsData(now, "day", scheduleId, "friends")),
			tgbotapi.NewInlineKeyboardButtonData("Расписание на неделю", formatScheduleButtonsData(now, "week", scheduleId, "friends")),
		},
		{tgbotapi.NewInlineKeyboardButtonData("Удалить друга", "/friends/delete/"+scheduleId)},
		{tgbotapi.NewInlineKeyboardButtonData(txtBackButton, "/choose_friend_back")},
	}
}

func FriendsButtons(friends map[string]string) [][]tgbotapi.InlineKeyboardButton {
	// friends: ключ - personId (SubjectId), значение - ФИО
	buttons := [][]tgbotapi.InlineKeyboardButton{{tgbotapi.NewInlineKeyboardButtonData("Добавить друга", "/add_friend")}}
	return append(buttons, InlineRowButtons(friends, 1)...)
}
