package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/pkg/bot"
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
	"time"
)

func formatStudents(students []parser.Student) (string, [][]tgbotapi.InlineKeyboardButton) {
	text := "Вот все студенты, которых мне удалось найти:"
	for k, s := range students {
		text += fmt.Sprintf("\n\n<b>%d</b> ", k+1) + fmt.Sprintf(formatStudent, s.FullName, s.SpecialtyName, s.SpecialtyProfile, s.FlowCode)
	}
	return text, tgmodel.NumbersButtons(len(students), 3)
}

func findStudent(c bot.Context) (parser.Student, error) {
	var students []parser.Student
	if err := c.GetData("students", &students); err != nil {
		return parser.Student{}, err
	}
	cb := c.Update().CallbackQuery
	if cb == nil {
		return parser.Student{}, ErrIncorrectInput
	}
	num, err := strconv.Atoi(cb.Data)
	if err != nil || num > len(students) {
		return parser.Student{}, ErrIncorrectInput
	}

	return students[num-1], nil
}

// Функция парсит данные из коллбэка с расписанием и оценками.
// Паттерн ввода в формате тип/дата, например day/2006-01-02,
// который в дальнейшем используется для получения расписания на 2 января 2006.
// Ввод week/2006-01-02 будет использован для получения расписания на неделю, начинающуюся с этой даты.
// Ввод grades/2006-01-02 будет использован для получения оценок на эту дату
func parseCallbackDate(c bot.Context) (string, time.Time, error) {
	cb := c.Update().CallbackQuery
	if cb == nil {
		return "", time.Time{}, ErrIncorrectInput
	}
	data := strings.Split(cb.Data, "/")
	if len(data) != 2 {
		return "", time.Time{}, ErrIncorrectInput
	}
	t, err := time.Parse(time.DateOnly, data[1])
	return data[0], t, err
}

func studentDaySchedule(ctx context.Context, parser parser.Parser, now time.Time, scheduleId string) (string, [][]tgbotapi.InlineKeyboardButton, error) {
	schedule, err := parser.DaySchedule(ctx, scheduleId, now)
	if err != nil {
		return "", nil, err
	}

	text := fmt.Sprintf("Расписание на %s:", now.Format("02.01"))
	for _, lesson := range schedule {
		text += "\n" + fmt.Sprintf(formatLesson, lesson.Time, lesson.Subject, lesson.Name, lesson.Type, lesson.AuditoriumNum, lesson.BuildingAddr, lesson.Lector) + "\n"
	}
	if len(schedule) == 0 {
		text = fmt.Sprintf("На %s занятий нет!", now.Format("02.01"))
	}
	return text, tgmodel.DayScheduleButtons(now), nil
}

func studentWeekSchedule(ctx context.Context, parser parser.Parser, now time.Time, scheduleId string) (string, [][]tgbotapi.InlineKeyboardButton, error) {
	schedule, err := parser.WeekSchedule(ctx, scheduleId, now)
	if err != nil {
		return "", nil, err
	}

	start := now.Day() - int(now.Weekday()) + 1
	weekStart := time.Date(now.Year(), now.Month(), start, 0, 0, 0, 0, now.Location())
	weekEnd := time.Date(now.Year(), now.Month(), start+6, 0, 0, 0, 0, now.Location())

	text := fmt.Sprintf("Расписание на неделю %s - %s:\n", weekStart.Format("02.01"), weekEnd.Format("02.01"))

	for d := 1; d <= 6; d++ {
		text += fmt.Sprintf("\n<b><i>%s</i></b>:", dates[d])
		if len(schedule[d]) == 0 {
			text += "\nЗанятий нет\n"
			continue
		}
		for _, lesson := range schedule[d] {
			text += "\n" + fmt.Sprintf(formatLesson, lesson.Time, lesson.Subject, lesson.Name, lesson.Type, lesson.AuditoriumNum, lesson.BuildingAddr, lesson.Lector) + "\n"
		}
	}
	return text, tgmodel.WeekScheduleButtons(now), nil
}

func studentSchedule(c bot.Context, p parser.Parser, backKB [][]tgbotapi.InlineKeyboardButton) error {
	t, day, err := parseCallbackDate(c)
	if err != nil {
		if errors.Is(err, ErrIncorrectInput) {
			return c.SendMessage(txtWarn)
		}
		return err
	}
	var scheduleId string
	if err = c.GetData("schedule_id", &scheduleId); err != nil {
		return err
	}
	var fullName string
	if err = c.GetData("full_name", &fullName); err != nil {
		return err
	}

	var (
		text string
		kb   [][]tgbotapi.InlineKeyboardButton
	)
	switch t {
	case "day":
		text, kb, err = studentDaySchedule(c.Context(), p, day, scheduleId)
		if err != nil {
			return err
		}
	case "week":
		text, kb, err = studentWeekSchedule(c.Context(), p, day, scheduleId)
		if err != nil {
			return err
		}
	default:
		return errors.New("studentSchedule unexpected error")
	}
	text = fmt.Sprintf(formatFullName, fullName) + text
	if backKB != nil {
		kb = append(kb, backKB...)
	}
	return c.EditMessageWithInlineKB(text, kb)
}

func studentCurrentSchedule(c bot.Context, p parser.Parser, scheduleId string) (string, [][]tgbotapi.InlineKeyboardButton, error) {
	switch c.Text() {
	case "day_schedule":
		return studentDaySchedule(c.Context(), p, time.Now(), scheduleId)
	case "week_schedule":
		return studentWeekSchedule(c.Context(), p, time.Now(), scheduleId)
	default:
		return "Ой! Произошла ошибка при сборе данных...", nil, nil
	}
}

// От модеуса даты приходят в неудобном для чтения виде, поэтому мы приводим их в нормальный вариант
func parseSemesterDate(d string) string {
	t, err := time.Parse("2006-01-02T05:04:15", d)
	if err != nil {
		return d
	}
	return t.Format("02.01.2006")
}
