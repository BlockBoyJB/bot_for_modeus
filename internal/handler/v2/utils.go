package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
	"time"
)

// Для данных, которые в кэше используются в качестве ускорения работы, вешаем таймауты хранения.
// Нам не нужно хранить данные пользователей в кэше, которые единожды воспользовались ботом и больше не используют
const (
	defaultCacheTimeout     = time.Hour
	textCacheTimeout        = time.Minute * 12
	gradesInputCacheTimeout = time.Hour * 24 * 14
	fullNameCacheTimeout    = time.Hour * 24 * 7
	friendsCacheTimeout     = time.Hour * 24 * 7
	semesterCacheTimeout    = time.Hour * 12
)

var (
	defaultLocation = time.FixedZone("Tyumen", 5*60*60) // По умолчанию GMT+5 (время в Тюмени)

	dates = map[int]string{
		1: "Понедельник",
		2: "Вторник",
		3: "Среда",
		4: "Четверг",
		5: "Пятница",
		6: "Суббота",
		7: "Воскресенье",
	}

	months = map[time.Month]string{
		time.January:   "января",
		time.February:  "февраля",
		time.March:     "марта",
		time.April:     "апреля",
		time.May:       "мая",
		time.June:      "июня",
		time.July:      "июля",
		time.August:    "августа",
		time.September: "сентября",
		time.October:   "октября",
		time.November:  "ноября",
		time.December:  "декабря",
	}
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

	text := fmt.Sprintf("Расписание на <b>%d %s</b>:\n", now.Day(), months[now.Month()])
	for _, lesson := range schedule {
		text += "\n" + fmt.Sprintf(formatLesson, lesson.Time, lesson.Subject, lesson.Name, lesson.Type, lesson.AuditoriumNum, lesson.BuildingAddr, lesson.Lector) + "\n"
	}
	if len(schedule) == 0 {
		text = fmt.Sprintf("На <b>%d %s</b> занятий нет!", now.Day(), months[now.Month()])
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

	text := fmt.Sprintf("Расписание на <b>%d %s - %d %s</b>:\n", weekStart.Day(), months[weekStart.Month()], weekEnd.Day(), months[weekEnd.Month()])

	for d := 1; d <= 6; d++ {
		text += fmt.Sprintf("\n<b><i>%s %s</i></b>:", dates[d], weekStart.AddDate(0, 0, d-1).Format("02.01"))
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

// Вынес клавиатуру с друзьями сюда, чтобы сразу работать с FriendOutput, а не tgmodel.Button
func friendsButtons(friends []service.FriendOutput) [][]tgbotapi.InlineKeyboardButton {
	buttons := make([][]tgbotapi.InlineKeyboardButton, 0, len(friends)+1)

	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Добавить друга", "/add_friend")})

	for _, f := range friends {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(f.FullName, "/friends/choose/"+f.ScheduleId)})
	}
	return buttons
}

// Отдельно сохраняем все когда-либо использованные пользователем ФИО.
// К сожалению, телеграм имеет ограничение на размер callback data (64 байта) (сделали хотя бы 1kb!!!!).
// Поэтому идея сделать коллбэк на расписание в формате "тип/дата/scheduleId/ФИО" обернулась крахом.
// Было принято решение scheduleId оставить, а ФИО вынести в общие данные
func getFullName(c bot.Context, p parser.Parser, scheduleId string) (fullName string, err error) {
	if err = c.GetCommonData("full_name:"+scheduleId, &fullName); err == nil {
		return
	}

	// Есть необходимость использовать паттерн singleflight, потому что ФИО студента - общая информация.
	// Если не найдем ее в кэше - будет много запросов в модеус с одними и теми же данными.
	// P.S. кого я обманываю, никто не будет пользоваться этим одновременно, ведь у меня 10 человек онлайна

	ctx, cancel := context.WithTimeout(c.Context(), time.Second*7)
	defer cancel()

	result, err := c.DoOnce(ctx, "full_name:"+scheduleId, func() (any, error) {
		s, err := p.FindStudentById(scheduleId)
		if err != nil {
			return nil, err
		}
		_ = c.SetCommonData("full_name:"+s.ScheduleId, s.FullName, fullNameCacheTimeout)
		return s.FullName, nil
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
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
	_ = c.SetTempData("grades_input", gi, gradesInputCacheTimeout)

	if gi.Password != "" && decrypt {
		gi.Password, err = u.Decrypt(gi.Password)
		if err != nil {
			return parser.GradesInput{}, err
		}
	}
	return
}

func lookupFriends(c bot.Context, u service.User) (friends []service.FriendOutput, err error) {
	if err = c.GetData("friends", &friends); err == nil {
		return
	}
	user, err := u.Find(c.Context(), c.UserId())
	if err != nil {
		return nil, err
	}
	_ = c.SetTempData("friends", user.Friends, friendsCacheTimeout)
	return user.Friends, nil
}

func lookupSemesters(c bot.Context, p parser.Parser, gi parser.GradesInput) ([]parser.Semester, error) {
	var semesters []parser.Semester
	if err := c.GetData("semesters", &semesters); err == nil {
		return semesters, nil
	}
	semesters, err := p.FindAllSemesters(gi)
	if err != nil {
		return nil, err
	}
	if len(semesters) == 0 {
		return nil, errors.New("lookupSemesters: cannot find semesters from modeus")
	}
	_ = c.SetTempData("semesters", semesters, semesterCacheTimeout)
	return semesters, nil
}

// Функция ищет семестр по semesterId. Сначала кэш, потом запрос в модеус. Если semesterId не указан (пустая строка), то возвращаем последний (текущий)
func lookupSemester(c bot.Context, p parser.Parser, gi parser.GradesInput, semesterId string) (semester parser.Semester, err error) {
	if err = c.GetData("semester:"+semesterId, &semester); err == nil { // Ничего не найдет, если семестр не указан. Важно следить, чтобы Set всегда был с semesterId, иначе фатально
		return
	}
	semesters, err := lookupSemesters(c, p, gi)
	if err != nil {
		return parser.Semester{}, err
	}
	if semesterId == "" {
		semester = semesters[len(semesters)-1]
		_ = c.SetTempData("semester:"+semester.Id, semester, semesterCacheTimeout)
		return
	}
	for _, s := range semesters {
		if s.Id == semesterId {
			_ = c.SetTempData("semester:"+s.Id, s, semesterCacheTimeout)
			return s, nil
		}
	}
	return parser.Semester{}, fmt.Errorf("lookupSemester: semester with id %s not found", semesterId)
}

// От модеуса даты приходят в неудобном для чтения виде, поэтому мы приводим их в нормальный вариант
func parseSemesterDate(d string) string {
	t, err := time.Parse("2006-01-02T05:04:15", d)
	if err != nil {
		return d
	}
	return t.Format("02.01.2006")
}

// Функция добавляет/обновляет логин и пароль пользователя.
// Отправляет сообщения об ошибках ввода.
// Удаляет сообщение от пользователя с введенным логином и паролем
func addLoginPassword(c bot.Context, u service.User) error {
	data := strings.Fields(c.Text())
	if len(data) != 2 {
		return c.SendMessage(txtIncorrectLoginPassInput)
	}

	if err := c.DelData("grades_input"); err != nil { // сначала важно удалить старые данные из кэша
		return err
	}
	err := u.UpdateLoginPassword(c.Context(), service.UserLoginPasswordInput{
		UserId:   c.UserId(),
		Login:    data[0],
		Password: data[1],
	})
	if err != nil {
		if errors.Is(err, service.ErrUserIncorrectLogin) {
			return c.SendMessage(txtIncorrectLoginPassInput)
		}
		return err
	}
	return c.DeleteLastMessage()
}
