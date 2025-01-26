package tgmodel

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

func ChooseFriendAction(scheduleId string) [][]tgbotapi.InlineKeyboardButton {
	now := time.Now().In(defaultLocation)
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
	buttons := make([][]tgbotapi.InlineKeyboardButton, 0, len(friends)+1)

	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Добавить друга", "/add_friend")})

	for scheduleId, fullName := range friends {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(fullName, "/friends/choose/"+scheduleId)})
	}
	return buttons
}
