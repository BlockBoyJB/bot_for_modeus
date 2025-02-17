package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"time"
)

type scheduleRouter struct {
	user   service.User
	parser parser.Parser
}

func newScheduleRouter(b bot.Router, user service.User, parser parser.Parser) {
	r := &scheduleRouter{
		user:   user,
		parser: parser,
	}

	{
		g := b.Group(metricsMiddleware("schedule"))

		g.Command("/day_schedule", r.cmdDaySchedule)
		g.Message(tgmodel.DayScheduleButton, r.cmdDaySchedule)
		g.Command("/week_schedule", r.cmdWeekSchedule)
		g.Message(tgmodel.WeekScheduleButton, r.cmdWeekSchedule)
		g.AddTree(bot.OnCallback, "/user/:type/:date/:schedule_id", r.callbackUserSchedule)
	}
	{
		g := b.Group(metricsMiddleware("grades"))

		g.Command("/grades", r.cmdGrades)
		g.Message(tgmodel.GradesButton, r.cmdGrades)
		g.AddTree(bot.OnCallback, "/grades/semester/change/:semester_id", r.callbackChangeSemester)
		g.AddTree(bot.OnCallback, "/grades/semester/:semester_id/subjects", r.callbackChooseSemesterSubject)
		g.AddTree(bot.OnCallback, "/grades/semester/:semester_id", r.callbackSemesterGrades)
		g.AddTree(bot.OnCallback, "/grades/subjects/:subject_id/:index", r.callbackSubjectDetailedInfo)
	}
}

func (r *scheduleRouter) cmdDaySchedule(c bot.Context) error {
	gi, err := lookupGI(c, r.user, false)
	if err != nil {
		return err
	}

	text, kb, err := studentDaySchedule(r.parser, time.Now(), gi.ScheduleId, "user")
	if err != nil {
		return err
	}

	// Кнопка оценок на день доступна только для пользователей с логином и паролем
	if gi.Login != "" && gi.Password != "" {
		kb = append(kb, tgmodel.WatchDayGradesButton(time.Now())...)
	}

	return c.SendMessageWithInlineKB(text, kb)
}

func (r *scheduleRouter) cmdWeekSchedule(c bot.Context) error {
	gi, err := lookupGI(c, r.user, false)
	if err != nil {
		return err
	}

	text, kb, err := studentWeekSchedule(r.parser, time.Now(), gi.ScheduleId, "user")
	if err != nil {
		return err
	}
	return c.SendMessageWithInlineKB(text, kb)
}

func (r *scheduleRouter) callbackUserSchedule(c bot.Context) error {
	t, day, _, err := parseCallbackDate(c)
	if err != nil {
		return err
	}

	gi, err := lookupGI(c, r.user, false)
	if err != nil {
		return err
	}

	var (
		text string
		kb   [][]tgbotapi.InlineKeyboardButton
	)

	switch t {
	case "day":
		text, kb, err = studentDaySchedule(r.parser, day, gi.ScheduleId, "user")
		if err != nil {
			return err
		}
		// доступно только пользователем с логином и паролем
		if gi.Login != "" && gi.Password != "" {
			kb = append(kb, tgmodel.WatchDayGradesButton(day)...)
		}
	case "week":
		text, kb, err = studentWeekSchedule(r.parser, day, gi.ScheduleId, "user")
	case "grades":
		// на всякий случай, хотя фактически невозможно
		if gi.Login == "" || gi.Password == "" {
			kb = append(tgmodel.BackButton("/user/day/"+day.Format(time.DateOnly)+"/"+gi.ScheduleId), tgmodel.GradesLink...)
			return c.EditMessageWithInlineKB(txtRequiredLoginPass, kb)
		}
		gi.Password, err = r.user.Decrypt(gi.Password) // здесь явно дешифруем, потому что по умолчанию в зашифрованном виде
		if err != nil {
			return err
		}
		grades, e := r.parser.DayGrades(day, gi)
		if e != nil {
			return e
		}

		text = fmt.Sprintf("Оценки на <b>%d %s</b>:\n", day.Day(), months[day.Month()])
		for _, grade := range grades {
			text += "\n" + fmt.Sprintf(formatDayGrades, grade.Time, grade.Subject, grade.Name, grade.Type, grade.Attendance, grade.Grades)
		}
		if len(grades) == 0 {
			text += "\nОценок нет!"
		}
		kb = tgmodel.DayGradesButtons(day)
	}
	return c.EditMessageWithInlineKB(text, kb)
}

func (r *scheduleRouter) cmdGrades(c bot.Context) error {
	gi, err := lookupGI(c, r.user, true)
	if err != nil {
		return err
	}
	if gi.Login == "" || gi.Password == "" {
		return c.SendMessageWithInlineKB(txtRequiredLoginPass, tgmodel.GradesLink)
	}

	semester, err := lookupSemester(c, r.parser, gi, "")
	if err != nil {
		return err
	}

	// Были вопросы насчет кэширования текста тут, потому что логичным является
	// использовать команду /grades для "перезагрузки" текущего состояния оценок.
	// Но решил добавить кэширование во избежание злоупотребления командой и снижения нагрузки на модеус
	var text string
	if err = c.GetData("semester_grades:"+semester.Id, &text); err == nil {
		return c.SendMessageWithInlineKB(text, tgmodel.GradesButtons(semester.Id))
	}

	grades, err := r.parser.SemesterTotalGrades(gi, semester)
	if err != nil {
		return err
	}
	text = "Вот все оценки по изучаемым дисциплинам в текущем семестре:"
	for _, subjectGrades := range grades {
		text += "\n" + fmt.Sprintf(formatSemesterGrades, subjectGrades.Status, subjectGrades.Name, subjectGrades.CurrentResult, subjectGrades.SemesterResult, subjectGrades.PresentRate, subjectGrades.AbsentRate, subjectGrades.UndefinedRate) + "\n"
	}
	_ = c.SetTempData("semester_grades:"+semester.Id, text, textCacheTimeout)
	return c.SendMessageWithInlineKB(text, tgmodel.GradesButtons(semester.Id))
}

func (r *scheduleRouter) callbackChangeSemester(c bot.Context) error {
	gi, err := lookupGI(c, r.user, true)
	if err != nil {
		return err
	}
	semesters, err := lookupSemesters(c, r.parser, gi)
	if err != nil {
		return err
	}

	buttons := make([]tgmodel.Button, 0, len(semesters))
	for _, s := range semesters {
		buttons = append(buttons, tgmodel.Button{
			Text: fmt.Sprintf("%dй семестр. (%s - %s)", s.Number, parseSemesterDate(s.StartDate), parseSemesterDate(s.EndDate)),
			Data: "/grades/semester/" + s.Id,
		})
	}
	// TODO можно подсветить текущий semester

	kb := append(tgmodel.CustomInlineRowButtons(buttons, 1), tgmodel.BackButton("/grades/semester/"+c.Param("semester_id"))...)
	return c.EditMessageWithInlineKB("Выберите семестр:", kb)
}

func (r *scheduleRouter) callbackSemesterGrades(c bot.Context) error {
	var text string
	if err := c.GetData("semester_grades:"+c.Param("semester_id"), &text); err == nil {
		return c.EditMessageWithInlineKB(text, tgmodel.GradesButtons(c.Param("semester_id")))
	}

	gi, err := lookupGI(c, r.user, true)
	if err != nil {
		return err
	}

	semester, err := lookupSemester(c, r.parser, gi, c.Param("semester_id"))
	if err != nil {
		return err
	}

	grades, err := r.parser.SemesterTotalGrades(gi, semester)
	if err != nil {
		return err
	}

	text = fmt.Sprintf("Вот все оценки по изучаемым дисциплинам в %dм семестре (%s - %s):", semester.Number, parseSemesterDate(semester.StartDate), parseSemesterDate(semester.EndDate))
	for _, g := range grades {
		text += "\n" + fmt.Sprintf(formatSemesterGrades, g.Status, g.Name, g.CurrentResult, g.SemesterResult, g.PresentRate, g.AbsentRate, g.UndefinedRate) + "\n"
	}
	_ = c.SetTempData("semester_grades:"+semester.Id, text, textCacheTimeout)
	return c.EditMessageWithInlineKB(text, tgmodel.GradesButtons(semester.Id))
}

func (r *scheduleRouter) callbackChooseSemesterSubject(c bot.Context) error {
	semesterId := c.Param("semester_id")

	gi, err := lookupGI(c, r.user, true)
	if err != nil {
		return err
	}
	semester, err := lookupSemester(c, r.parser, gi, semesterId)
	if err != nil {
		return err
	}
	var subjects map[string]string

	// `subjects` сначала смотрим в кэше. Если не нашли/ошибка, то придется спрашивать у модеуса. Не забываем кэшировать
	if err = c.GetData("semester_subjects:"+semesterId, &subjects); err != nil {
		subjects, err = r.parser.FindSemesterSubjects(gi, semester)
		if err != nil {
			return err
		}
		_ = c.SetTempData("semester_subjects:"+semesterId, subjects, defaultCacheTimeout)
	}

	// коллбэк на просмотр посещений по предмету в формате /grades/subjects/:subject_id/:index
	// можно было бы вместо index (номер семестра) запихнуть semesterId, но увы и ах: размер CallbackData <= 64 байта
	buttons := make(map[string]string, len(subjects))
	for k, v := range subjects {
		key := fmt.Sprintf("/grades/subjects/%s/%d", k, semester.Number)
		buttons[key] = v
	}
	kb := append(tgmodel.InlineRowButtons(buttons, 1), tgmodel.BackButton("/grades/semester/"+semester.Id)...)
	return c.EditMessageWithInlineKB("Выберите предмет:", kb)
}

func (r *scheduleRouter) callbackSubjectDetailedInfo(c bot.Context) error {
	// TODO кэшировать?
	gi, err := lookupGI(c, r.user, true)
	if err != nil {
		return err
	}
	semesters, err := lookupSemesters(c, r.parser, gi)
	if err != nil {
		return err
	}
	s := semesters[len(semesters)-1] // делаем по умолчанию последний, чтобы не возвращать ошибку

	// Из коллбэка находим номер семестра, идем циклом в попытке найти совпадение...
	// P.S. цикл в обратную сторону, потому что чаще смотрят информацию о парах последнего (текущего) семестра
	if index, err := strconv.Atoi(c.Param("index")); err == nil {
		for i := len(semesters) - 1; i >= 0; i-- {
			if semesters[i].Number == index {
				s = semesters[i]
				break
			}
		}
	}

	subjectLessons, err := r.parser.SubjectDetailedInfo(gi, s, c.Param("subject_id"))
	if err != nil {
		return err
	}
	var messages []string
	text := "Вот все оценки за проведенные пары по выбранному предмету:"
	textLength := len(text)
	for _, lesson := range subjectLessons {
		n := "\n" + fmt.Sprintf(formatLessonGrades, lesson.Name, lesson.Type, lesson.Time, lesson.Attendance, lesson.Grades)
		textLength += len(n)
		if textLength > 4096 {
			messages = append(messages, text)
			textLength = 0
			text = ""
		}
		text += n
	}
	if textLength != 0 {
		messages = append(messages, text)
	}
	if err = c.EditMessage(messages[0]); err != nil {
		return err
	}
	for i := 1; i < len(messages)-1; i++ {
		if err = c.SendMessage(messages[i]); err != nil {
			return err
		}
	}
	return c.SendMessageWithInlineKB(messages[len(messages)-1], tgmodel.BackButton(fmt.Sprintf("/grades/semester/%s/subjects", s.Id)))
}
