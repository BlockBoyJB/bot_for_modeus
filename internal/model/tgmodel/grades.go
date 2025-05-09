package tgmodel

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

var GradesLink = [][]tgbotapi.InlineKeyboardButton{{tgbotapi.NewInlineKeyboardButtonURL("Посмотреть на сайте", "https://utmn.modeus.org/students-app/my-results")}}

func GradesButtons(semesterId string) [][]tgbotapi.InlineKeyboardButton {
	return [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("Оценки по встречам", fmt.Sprintf("/grades/semester/%s/subjects", semesterId)),
			tgbotapi.NewInlineKeyboardButtonData("Другой семестр", fmt.Sprintf("/grades/semester/change/%s", semesterId)),
		},
	}
}

func WatchDayGradesButton(now time.Time) [][]tgbotapi.InlineKeyboardButton {
	now = now.In(defaultLocation)
	return [][]tgbotapi.InlineKeyboardButton{{
		tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Оценки на %s", now.Format("02.01")), formatScheduleButtonsData(now, "grades", "user", "user")),
	}}
}

func DayGradesButtons(now time.Time) [][]tgbotapi.InlineKeyboardButton {
	return append(
		dayButtons(now, "grades", "user", "user"),
		BackButton(formatScheduleButtonsData(now, "day", "user", "user"))...,
	)
}
