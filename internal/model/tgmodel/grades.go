package tgmodel

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

var GradesLink = [][]tgbotapi.InlineKeyboardButton{{tgbotapi.NewInlineKeyboardButtonURL("Посмотреть на сайте", "https://utmn.modeus.org/students-app/my-results")}}

var GradesButtons = [][]tgbotapi.InlineKeyboardButton{
	{
		tgbotapi.NewInlineKeyboardButtonData("Оценки по встречам", "/subject_lessons_grades"),
		tgbotapi.NewInlineKeyboardButtonData("Другой семестр", "/change_semester"),
	},
}

func WatchDayGradesButton(now time.Time) [][]tgbotapi.InlineKeyboardButton {
	return [][]tgbotapi.InlineKeyboardButton{{
		tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Оценки на %s", now.Format("02.01")), "grades/"+now.Format(time.DateOnly)),
	}}
}

func DayGradesButtons(now time.Time) [][]tgbotapi.InlineKeyboardButton {
	return append(dayButtons(now, "grades"), BackButton("day/"+now.Format(time.DateOnly))...)
}
