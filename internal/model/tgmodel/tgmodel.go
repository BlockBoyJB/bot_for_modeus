package tgmodel

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"time"
)

const (
	txtBackButton = "⬅️ Назад"
)

var nums = map[int]string{
	0: "0️⃣",
	1: "1️⃣",
	2: "2️⃣",
	3: "3️⃣",
	4: "4️⃣",
	5: "5️⃣",
	6: "6️⃣",
	7: "7️⃣",
	8: "8️⃣",
	9: "9️⃣",
}

var SettingsButtons = [][]tgbotapi.InlineKeyboardButton{
	{tgbotapi.NewInlineKeyboardButtonData("Добавить логин и пароль", "/add_login_password"), tgbotapi.NewInlineKeyboardButtonData("Изменить ФИО", "/update_full_name")},
}

var YesOrNoButtons = [][]tgbotapi.InlineKeyboardButton{
	{tgbotapi.NewInlineKeyboardButtonData("Да", "да"), tgbotapi.NewInlineKeyboardButtonData("Нет", "нет")},
}

func OtherStudentButtons(scheduleId string) [][]tgbotapi.InlineKeyboardButton {
	now := time.Now()
	return [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("Расписание на день", formatScheduleButtonsData(now, "day", scheduleId, "student")),
			tgbotapi.NewInlineKeyboardButtonData("Расписание на неделю", formatScheduleButtonsData(now, "week", scheduleId, "student")),
		},
		{tgbotapi.NewInlineKeyboardButtonData(txtBackButton, "/choose_other_student_back")},
	}
}

var MyProfileButtons = [][]tgbotapi.InlineKeyboardButton{
	{tgbotapi.NewInlineKeyboardButtonData("Обо мне", "/about_me"), tgbotapi.NewInlineKeyboardButtonData("Рейтинги", "/ratings")},
}

var HelpButtons = [][]tgbotapi.InlineKeyboardButton{
	{
		tgbotapi.NewInlineKeyboardButtonData("🗓 Расписание", "/help_schedule"),
		tgbotapi.NewInlineKeyboardButtonData("📊 Оценки", "/help_grades"),
	},
	{
		tgbotapi.NewInlineKeyboardButtonData("👨‍🎓👩‍🎓 Друзья", "/help_friends"),
		tgbotapi.NewInlineKeyboardButtonData("👥 Другие студенты", "/help_other_student"),
	},
	{
		tgbotapi.NewInlineKeyboardButtonData("⚙️ Настройки", "/help_settings"),
		tgbotapi.NewInlineKeyboardButtonData("\U0001FAF5 Обо мне", "/help_me"),
	},
	{
		tgbotapi.NewInlineKeyboardButtonData("🛡 Поддержка", "/help_support"),
		tgbotapi.NewInlineKeyboardButtonData("❓ FAQ", "/help_faq"),
	},
	{
		tgbotapi.NewInlineKeyboardButtonData("🏫 Адреса корпусов", "/help_buildings"),
	},
}

func NumbersButtons(k, size int) [][]tgbotapi.InlineKeyboardButton {
	// Идею реализации подсмотрел у aiogram KeyboardBuilder.adjust
	// https://docs.aiogram.dev/en/latest/utils/keyboard.html#aiogram.utils.keyboard.ReplyKeyboardBuilder.adjust
	var buttons [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	for i := 1; i <= k; i++ {
		if len(row) >= size {
			buttons = append(buttons, row)
			row = nil
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(formatNumber(i), fmt.Sprintf("%d", i)))
	}
	if len(row) != 0 {
		buttons = append(buttons, row)
	}
	return buttons
}

// BackButton Одна реализация кнопки "назад" под разные коллбэки
func BackButton(callback string) [][]tgbotapi.InlineKeyboardButton {
	return [][]tgbotapi.InlineKeyboardButton{{tgbotapi.NewInlineKeyboardButtonData(txtBackButton, callback)}}
}

func InlineRowButtons(data map[string]string, size int) [][]tgbotapi.InlineKeyboardButton {
	var buttons [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	for k, v := range data {
		if len(row) >= size {
			buttons = append(buttons, row)
			row = nil
		}
		// инвертируем значения, потому что ключ не должен быть виден пользователю
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(v, k))
	}
	if len(row) != 0 {
		buttons = append(buttons, row)
	}
	return buttons
}

func dayButtons(now time.Time, bType, scheduleId, prefix string) [][]tgbotapi.InlineKeyboardButton {
	y := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	t := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return [][]tgbotapi.InlineKeyboardButton{{
		tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("◀️ %s", y.Format("02.01")), formatScheduleButtonsData(y, bType, scheduleId, prefix)),
		tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s ▶️", t.Format("02.01")), formatScheduleButtonsData(t, bType, scheduleId, prefix)),
	}}
}

func formatScheduleButtonsData(t time.Time, bType, scheduleId, prefix string) string {
	return fmt.Sprintf("/%s/%s/%s/%s", prefix, bType, t.Format(time.DateOnly), scheduleId)
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

const (
	DayScheduleButton  = "📙 Расписание на день"
	WeekScheduleButton = "📚 Расписание на неделю"
	GradesButton       = "📊 Оценки"
	FriendsButton      = "👨‍🎓👩‍🎓 Друзья"
	OtherStudentButton = "👥 Другие студенты"
	MeButton           = "\U0001FAF5 Обо мне"
	SettingsButton     = "⚙️ Настройки"
	HelpButton         = "❓ Помощь"
)

var RowCommands = [][]tgbotapi.KeyboardButton{
	{
		tgbotapi.NewKeyboardButton(DayScheduleButton),
		tgbotapi.NewKeyboardButton(WeekScheduleButton),
	},
	{
		tgbotapi.NewKeyboardButton(GradesButton),
		tgbotapi.NewKeyboardButton(FriendsButton),
	},
	{
		tgbotapi.NewKeyboardButton(OtherStudentButton),
		tgbotapi.NewKeyboardButton(MeButton),
	},
	{
		tgbotapi.NewKeyboardButton(SettingsButton),
		tgbotapi.NewKeyboardButton(HelpButton),
	},
}

func formatNumber(k int) string {
	if k < 10 {
		return nums[k]
	}
	var result string
	for _, s := range strconv.Itoa(k) {
		n, _ := strconv.Atoi(string(s))
		result += nums[n]
	}
	return result
}
