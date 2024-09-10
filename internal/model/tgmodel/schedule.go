package tgmodel

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

var ScheduleLink = [][]tgbotapi.InlineKeyboardButton{{tgbotapi.NewInlineKeyboardButtonURL("Посмотреть на сайте", "https://utmn.modeus.org")}}

func DayScheduleButtons(now time.Time) [][]tgbotapi.InlineKeyboardButton {
	return dayButtons(now, "day")
}

func WeekScheduleButtons(now time.Time) [][]tgbotapi.InlineKeyboardButton {
	start := now.Day() - int(now.Weekday()) + 1

	prevWeekStart := time.Date(now.Year(), now.Month(), start-7, 0, 0, 0, 0, now.Location())
	prevWeekEnd := time.Date(now.Year(), now.Month(), start-1, 0, 0, 0, 0, now.Location())

	nextWeekStart := time.Date(now.Year(), now.Month(), start+7, 0, 0, 0, 0, now.Location())
	nextWeekEnd := time.Date(now.Year(), now.Month(), start+13, 0, 0, 0, 0, now.Location())

	return [][]tgbotapi.InlineKeyboardButton{{
		tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("◀️ %s - %s", prevWeekStart.Format("02.01"), prevWeekEnd.Format("02.01")), "week/"+prevWeekStart.Format(time.DateOnly)),
		tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s - %s ▶️", nextWeekStart.Format("02.01"), nextWeekEnd.Format("02.01")), "week/"+nextWeekStart.Format(time.DateOnly)),
	}}
}
