package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"time"
)

var (
	defaultLocation, _ = time.LoadLocation("Asia/Yekaterinburg") // По умолчанию GMT+5 (время в Тюмени)
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

// Функция парсит данные из параметров пути
// Паттерн ввода в формате тип/дата/scheduleId, например day/2006-01-02/aaaaaaaa-0000-0000-0000-aaaaaaaaaaaa,
// который в дальнейшем используется для получения расписания на 2 января 2006 для пользователя с uuid "aaaaaaaa-0000-0000-0000-aaaaaaaaaaaa".
// Ввод week/2006-01-02/uuid будет использован для получения расписания на неделю, начинающуюся с этой даты.
// Ввод grades/2006-01-02/uuid будет использован для получения оценок на эту дату
func parseCallbackDate(c bot.Context) (t string, day time.Time, scheduleId string, err error) {
	cb := c.Update().CallbackQuery
	if cb == nil {
		return "", time.Time{}, "", ErrIncorrectInput
	}

	var d string
	t, d, scheduleId = c.Param("type"), c.Param("date"), c.Param("schedule_id")

	if t == "" || d == "" || scheduleId == "" {
		return "", time.Time{}, "", ErrIncorrectInput
	}

	day, err = time.Parse(time.DateOnly, d)
	if err != nil {
		return "", time.Time{}, "", ErrIncorrectInput
	}
	return
}

func studentDaySchedule(parser parser.Parser, now time.Time, scheduleId, prefix string) (string, [][]tgbotapi.InlineKeyboardButton, error) {
	now = now.In(defaultLocation)
	schedule, err := parser.DaySchedule(scheduleId, now)
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
	return text, tgmodel.DayScheduleButtons(now, scheduleId, prefix), nil
}

func studentWeekSchedule(parser parser.Parser, now time.Time, scheduleId, prefix string) (string, [][]tgbotapi.InlineKeyboardButton, error) {
	now = now.In(defaultLocation)
	schedule, err := parser.WeekSchedule(scheduleId, now)
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
	return text, tgmodel.WeekScheduleButtons(now, scheduleId, prefix), nil
}

func studentSchedule(c bot.Context, p parser.Parser, prefix string, backKB [][]tgbotapi.InlineKeyboardButton) error {
	t, day, scheduleId, err := parseCallbackDate(c)
	if err != nil {
		if errors.Is(err, ErrIncorrectInput) {
			return c.SendMessage(txtWarn)
		}
		return err
	}
	fullName, err := getFullName(c, p, scheduleId)
	if err != nil {
		return err
	}

	var (
		text string
		kb   [][]tgbotapi.InlineKeyboardButton
	)
	switch t {
	case "day":
		text, kb, err = studentDaySchedule(p, day, scheduleId, prefix)
		if err != nil {
			return err
		}
	case "week":
		text, kb, err = studentWeekSchedule(p, day, scheduleId, prefix)
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

// Отдельно сохраняем все когда-либо использованные пользователем ФИО.
// К сожалению, телеграм имеет ограничение на размер callback data (64 байта) (сделали хотя бы 1kb!!!!).
// Поэтому идея сделать коллбэк на расписание в формате "тип/дата/scheduleId/ФИО" обернулась крахом.
// Было принято решение scheduleId оставить, а ФИО вынести в общие данные
func getFullName(c bot.Context, p parser.Parser, scheduleId string) (fullName string, err error) {
	if err = c.GetCommonData("full_name:"+scheduleId, &fullName); err == nil {
		return
	}
	s, err := p.FindStudentById(scheduleId)
	if err != nil {
		return "", err
	}
	_ = c.SetCommonData("full_name:"+s.ScheduleId, s.FullName, 0)
	return s.FullName, nil
}

// Функция возвращает основную структуру для работы с пользователем - GradesInput (в которой scheduleId, gradesId, login, password).
// Флаг decrypt для явного указания необходимости дешифровать пароль (если есть)
// Без этого каждый вызов будет занимать на ~14.5 мс (см. бенчмарк pkg/crypter/crypter_test.go BenchmarkCrypter_Decrypt)
// больше из-за постоянного дешифрования пароля даже там, где он не нужен
func lookupGI(c bot.Context, u service.User, decrypt bool) (gi parser.GradesInput, err error) {
	err = c.GetData("grades_input", &gi)
	if err == nil {
		if gi.Password != "" && decrypt {
			gi.Password, err = u.Decrypt(gi.Password)
			if err != nil {
				return parser.GradesInput{}, err
			}
		}
		return
	}
	if !errors.Is(err, bot.ErrKeyNotExists) {
		return parser.GradesInput{}, err
	}

	user, err := u.Find(c.Context(), c.UserId())
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			_ = c.SendMessage(txtUserNotFound)
			return parser.GradesInput{}, err
		}
		return parser.GradesInput{}, err
	}

	gi = parser.GradesInput{
		Login:      user.Login,
		Password:   user.Password,
		ScheduleId: user.ScheduleId,
		GradesId:   user.GradesId,
	}
	// Тут необязательно возвращать ошибку, поскольку grades input у нас есть,
	// однако нагрузка на бд возрастет с количеством несохраненных gi в кэш
	_ = c.SetData("grades_input", gi)

	if gi.Password != "" && decrypt {
		gi.Password, err = u.Decrypt(gi.Password)
		if err != nil {
			return parser.GradesInput{}, err
		}
	}
	return
}

func lookupFriends(c bot.Context, u service.User) (friends map[string]string, err error) {
	if err = c.GetData("friends", &friends); err == nil {
		return
	}
	user, err := u.Find(c.Context(), c.UserId())
	if err != nil {
		return nil, err
	}
	_ = c.SetData("friends", user.Friends)
	return user.Friends, nil
}

// От модеуса даты приходят в неудобном для чтения виде, поэтому мы приводим их в нормальный вариант
func parseSemesterDate(d string) string {
	t, err := time.Parse("2006-01-02T05:04:15", d)
	if err != nil {
		return d
	}
	return t.Format("02.01.2006")
}
