package tgmodel

import "fmt"

type Message struct {
	UserId     int64
	MessageId  int
	Text       string
	IsCallback bool
	IsCommand  bool
}

type (
	InlineButton struct {
		Key   string
		Value string
	}
	RowInlineButtons []InlineButton
)

var SettingsButtons = []RowInlineButtons{
	{InlineButton{Key: "Добавить логин и пароль", Value: "/add_login_password"}, InlineButton{Key: "Изменить ФИО", Value: "/update_full_name"}},
	//{InlineButton{Key: "Добавить регулярное оповещение", Value: "/add_notification"}}, // может когда нибудь сделаю
	{InlineButton{Key: "Удалить логин и пароль", Value: "/delete_login_password"}},
}

var YesOrNoButtons = []RowInlineButtons{
	{InlineButton{Key: "Да", Value: "да"}, InlineButton{Key: "Нет", Value: "нет"}},
}

var OtherStudentButtons = []RowInlineButtons{
	{InlineButton{Key: "Расписание на сегодня", Value: "day_schedule"}, InlineButton{Key: "Расписание на неделю", Value: "week_schedule"}},
}

func NumbersButtons(k, size int) []RowInlineButtons {
	// Идею реализации подсмотрел у aiogram KeyboardBuilder.adjust
	// https://docs.aiogram.dev/en/latest/utils/keyboard.html#aiogram.utils.keyboard.ReplyKeyboardBuilder.adjust
	var buttons []RowInlineButtons
	var row RowInlineButtons
	var num string
	for i := 1; i <= k; i++ {
		if len(row) >= size {
			buttons = append(buttons, row)
			row = nil
		}
		num = fmt.Sprintf("%d", i)
		row = append(row, InlineButton{Key: num, Value: num})
	}
	if len(row) != 0 {
		buttons = append(buttons, row)
	}
	return buttons
}

type UIBotCommand struct {
	Command     string
	Description string
}

var UICommands = []UIBotCommand{
	{
		Command:     "start",
		Description: "Перезапустить бота",
	},
	{
		Command:     "help",
		Description: "Помощь",
	},
	{
		Command:     "menu",
		Description: "Главное меню",
	},
	{
		Command:     "settings",
		Description: "Настройки",
	},
	{
		Command:     "day_schedule",
		Description: "Расписание на день",
	},
	{
		Command:     "week_schedule",
		Description: "Расписание на неделю",
	},
	{
		Command:     "grades",
		Description: "Посмотреть баллы",
	},
	{
		Command:     "other_student",
		Description: "Расписание другого студента",
	},
	{
		Command:     "stop",
		Description: "Остановить бота",
	},
}
