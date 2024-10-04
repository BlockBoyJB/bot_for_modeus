package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

const (
	// Решил кэшировать некоторые данные (в частности текст) в разделе оценок, потому что их повторное обновление очень долгое
	// Время кэширования небольшое из расчета на то, что эти данные не успеют измениться за этот промежуток
	defaultCacheTimeout = time.Minute * 7
)

var dates = map[int]string{
	1: "Понедельник",
	2: "Вторник",
	3: "Среда",
	4: "Четверг",
	5: "Пятница",
	6: "Суббота",
	7: "Воскресенье",
}

type scheduleRouter struct {
	user   service.User
	parser parser.Parser
}

func newScheduleRouter(b *bot.Bot, user service.User, parser parser.Parser) {
	r := &scheduleRouter{
		user:   user,
		parser: parser,
	}

	b.Command("/day_schedule", r.cmdDaySchedule)
	b.Message(tgmodel.DayScheduleButton, r.cmdDaySchedule)
	b.Command("/week_schedule", r.cmdWeekSchedule)
	b.Message(tgmodel.WeekScheduleButton, r.cmdWeekSchedule)
	b.State(stateUserSchedule, r.stateUserSchedule)

	b.Command("/grades", r.cmdGrades)
	b.Message(tgmodel.GradesButton, r.cmdGrades)
	b.Callback("/semester_grades_back", r.callbackSemesterGradesBack)
	b.Callback("/change_semester", r.callbackChangeSemester)
	b.State(stateChooseSemester, r.stateChooseSemester)
	b.Callback("/subject_lessons_grades", r.callbackSubjectLessonsGrades)
	b.Callback("/choose_subject_back", r.callbackChooseSubjectBack)
	b.State(stateChooseSubject, r.stateChooseSubject)
}

func (r *scheduleRouter) cmdDaySchedule(c bot.Context) error {
	user, err := r.user.Find(c.Context(), c.UserId())
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.SendMessage(txtUserNotFound)
		}
	}

	gi := parser.GradesInput{
		Login:      user.Login,
		Password:   user.Password,
		ScheduleId: user.ScheduleId,
		GradesId:   user.GradesId,
	}
	if err = c.SetData("grades_input", gi); err != nil {
		return err
	}

	text, kb, err := studentDaySchedule(c.Context(), r.parser, time.Now(), user.ScheduleId)
	if err != nil {
		return err
	}

	// Кнопка оценок на день доступна только для пользователей с логином и паролем
	if user.Login != "" && user.Password != "" {
		kb = append(kb, tgmodel.WatchDayGradesButton(time.Now())...)
	}

	if err = c.SendMessageWithInlineKB(text, kb); err != nil {
		return err
	}
	return c.SetState(stateUserSchedule)
}

func (r *scheduleRouter) cmdWeekSchedule(c bot.Context) error {
	user, err := r.user.Find(c.Context(), c.UserId())
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.SendMessage(txtUserNotFound)
		}
	}

	gi := parser.GradesInput{
		Login:      user.Login,
		Password:   user.Password,
		ScheduleId: user.ScheduleId,
		GradesId:   user.GradesId,
	}
	if err = c.SetData("grades_input", gi); err != nil {
		return err
	}

	text, kb, err := studentWeekSchedule(c.Context(), r.parser, time.Now(), user.ScheduleId)
	if err != nil {
		return err
	}
	if err = c.SendMessageWithInlineKB(text, kb); err != nil {
		return err
	}
	return c.SetState(stateUserSchedule)
}

func (r *scheduleRouter) stateUserSchedule(c bot.Context) error {
	t, day, err := parseCallbackDate(c)
	if err != nil {
		if errors.Is(err, ErrIncorrectInput) {
			return c.SendMessage(txtWarn)
		}
		return err
	}

	var gi parser.GradesInput
	if err = c.GetData("grades_input", &gi); err != nil {
		return err
	}

	var (
		text string
		kb   [][]tgbotapi.InlineKeyboardButton
	)

	switch t {
	case "day":
		text, kb, err = studentDaySchedule(c.Context(), r.parser, day, gi.ScheduleId)
		if err != nil {
			return err
		}
		// доступно только пользователем с логином и паролем
		if gi.Login != "" && gi.Password != "" {
			kb = append(kb, tgmodel.WatchDayGradesButton(day)...)
		}
	case "week":
		text, kb, err = studentWeekSchedule(c.Context(), r.parser, day, gi.ScheduleId)
	case "grades":
		// на всякий случай, хотя фактически невозможно
		if gi.Login == "" || gi.Password == "" {
			kb = append(tgmodel.BackButton("day/"+day.Format(time.DateOnly)), tgmodel.GradesLink...)
			return c.EditMessageWithInlineKB(txtRequiredLoginPass, kb)
		}
		grades, e := r.parser.DayGrades(c.Context(), day, gi)
		if e != nil {
			return e
		}

		text = fmt.Sprintf("Вот все оценки за %s:", day.Format("02.01"))
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
	u, err := r.user.Find(c.Context(), c.UserId())
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.SendMessage(txtUserNotFound)
		}
		return err
	}
	if u.Login == "" || u.Password == "" {
		return c.SendMessageWithInlineKB(txtRequiredLoginPass, tgmodel.GradesLink)
	}

	gi := parser.GradesInput{
		Login:      u.Login,
		Password:   u.Password,
		ScheduleId: u.ScheduleId,
		GradesId:   u.GradesId,
	}

	semester, err := r.parser.FindCurrentSemester(c.Context(), gi)
	if err != nil {
		if errors.Is(err, parser.ErrIncorrectLoginPassword) {
			return c.SendMessage(txtIncorrectLoginPass)
		}
		return err
	}

	if err = c.SetData("grades_input", gi); err != nil {
		return err
	}
	if err = c.SetData("semester", semester); err != nil {
		return err
	}

	grades, err := r.parser.SemesterTotalGrades(c.Context(), gi, semester)
	if err != nil {
		return err
	}
	text := "Вот все оценки по изучаемым дисциплинам в текущем семестре:"
	for _, subjectGrades := range grades {
		text += "\n" + fmt.Sprintf(formatSemesterGrades, subjectGrades.Status, subjectGrades.Name, subjectGrades.CurrentResult, subjectGrades.SemesterResult, subjectGrades.PresentRate, subjectGrades.AbsentRate, subjectGrades.UndefinedRate) + "\n"
	}
	_ = c.SetTempData("semester_grades", text, defaultCacheTimeout)
	return c.SendMessageWithInlineKB(text, tgmodel.GradesButtons)
}

func (r *scheduleRouter) callbackSemesterGradesBack(c bot.Context) error {
	var text string
	if err := c.GetData("semester_grades", &text); err == nil {
		return c.EditMessageWithInlineKB(text, tgmodel.GradesButtons)
	}

	var gi parser.GradesInput
	if err := c.GetData("grades_input", &gi); err != nil {
		return err
	}
	var s parser.Semester
	if err := c.GetData("semester", &s); err != nil {
		return err
	}
	grades, err := r.parser.SemesterTotalGrades(c.Context(), gi, s)
	if err != nil {
		return err
	}
	text = fmt.Sprintf("Вот все оценки по изучаемым дисциплинам в %dм семестре (%s - %s):", s.Number, parseSemesterDate(s.StartDate), parseSemesterDate(s.EndDate))
	for _, subjectGrades := range grades {
		text += "\n" + fmt.Sprintf(formatSemesterGrades, subjectGrades.Status, subjectGrades.Name, subjectGrades.CurrentResult, subjectGrades.SemesterResult, subjectGrades.PresentRate, subjectGrades.AbsentRate, subjectGrades.UndefinedRate) + "\n"
	}
	return c.EditMessageWithInlineKB(text, tgmodel.GradesButtons)
}

func (r *scheduleRouter) callbackChangeSemester(c bot.Context) error {
	var gi parser.GradesInput
	if err := c.GetData("grades_input", &gi); err != nil {
		return err
	}
	semesters, err := r.parser.FindAllSemesters(c.Context(), gi)
	if err != nil {
		return err
	}
	if err = c.SetData("semesters", semesters); err != nil {
		return err
	}

	buttons := make(map[string]string)
	for id, s := range semesters {
		buttons[id] = fmt.Sprintf("%dй семестр. (%s - %s)", s.Number, parseSemesterDate(s.StartDate), parseSemesterDate(s.EndDate))
	}

	kb := append(tgmodel.InlineRowButtons(buttons, 1), tgmodel.BackButton("/semester_grades_back")...)
	if err = c.EditMessageWithInlineKB("Выберите семестр:", kb); err != nil {
		return err
	}
	return c.SetState(stateChooseSemester)
}

func (r *scheduleRouter) stateChooseSemester(c bot.Context) error {
	var semesters map[string]parser.Semester
	if err := c.GetData("semesters", &semesters); err != nil {
		return err
	}

	s, ok := semesters[c.Text()]
	if !ok {
		return errors.New("cannot find user semester")
	}
	if err := c.SetData("semester", s); err != nil {
		return err
	}

	var gi parser.GradesInput
	if err := c.GetData("grades_input", &gi); err != nil {
		return err
	}
	grades, err := r.parser.SemesterTotalGrades(c.Context(), gi, s)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("Вот все оценки по изучаемым дисциплинам в %dм семестре (%s - %s):", s.Number, parseSemesterDate(s.StartDate), parseSemesterDate(s.EndDate))
	for _, subjectGrades := range grades {
		text += "\n" + fmt.Sprintf(formatSemesterGrades, subjectGrades.Status, subjectGrades.Name, subjectGrades.CurrentResult, subjectGrades.SemesterResult, subjectGrades.PresentRate, subjectGrades.AbsentRate, subjectGrades.UndefinedRate) + "\n"
	}
	_ = c.SetTempData("semester_grades", text, defaultCacheTimeout)
	return c.EditMessageWithInlineKB(text, tgmodel.GradesButtons)
}

func (r *scheduleRouter) callbackSubjectLessonsGrades(c bot.Context) error {
	var gi parser.GradesInput
	if err := c.GetData("grades_input", &gi); err != nil {
		return err
	}
	var semester parser.Semester
	if err := c.GetData("semester", &semester); err != nil {
		return err
	}
	subjects, err := r.parser.FindSemesterSubjects(c.Context(), gi, semester)
	if err != nil {
		return err
	}
	_ = c.SetTempData("semester_subjects", subjects, defaultCacheTimeout)
	kb := append(tgmodel.InlineRowButtons(subjects, 1), tgmodel.BackButton("/semester_grades_back")...)
	if err = c.EditMessageWithInlineKB("Выберите предмет:", kb); err != nil {
		return err
	}
	return c.SetState(stateChooseSubject)
}

func (r *scheduleRouter) callbackChooseSubjectBack(c bot.Context) error {
	var subjects map[string]string
	if err := c.GetData("semester_subjects", &subjects); err == nil {
		_ = c.DeleteInlineKB()
		kb := append(tgmodel.InlineRowButtons(subjects, 1), tgmodel.BackButton("/semester_grades_back")...)
		if e := c.SendMessageWithInlineKB("Выберите предмет:", kb); e != nil {
			return e
		}
		return c.SetState(stateChooseSubject)
	}
	var gi parser.GradesInput
	if err := c.GetData("grades_input", &gi); err != nil {
		return err
	}
	var semester parser.Semester
	if err := c.GetData("semester", &semester); err != nil {
		return err
	}
	subjects, err := r.parser.FindSemesterSubjects(c.Context(), gi, semester)
	if err != nil {
		return err
	}
	_ = c.SetTempData("semester_subjects", subjects, defaultCacheTimeout)
	_ = c.DeleteInlineKB()
	if err = c.SendMessageWithInlineKB("Выберите предмет:", tgmodel.InlineRowButtons(subjects, 1)); err != nil {
		return err
	}
	return c.SetState(stateChooseSubject)
}

func (r *scheduleRouter) stateChooseSubject(c bot.Context) error {
	var gi parser.GradesInput
	if err := c.GetData("grades_input", &gi); err != nil {
		return err
	}
	var semester parser.Semester
	if err := c.GetData("semester", &semester); err != nil {
		return err
	}
	subjectLessons, err := r.parser.SubjectDetailedInfo(c.Context(), gi, semester, c.Text())
	if err != nil {
		return err
	}
	var messages []string
	text := "Вот все оценки за проведенные пары по выбранному предмету:"
	textLength := len(text)
	for i := 0; i < len(subjectLessons); i++ {
		n := "\n" + fmt.Sprintf(formatLessonGrades, i+1, subjectLessons[i].Name, subjectLessons[i].Type, subjectLessons[i].Time, subjectLessons[i].Attendance, subjectLessons[i].Grades)
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
	return c.SendMessageWithInlineKB(messages[len(messages)-1], tgmodel.BackButton("/choose_subject_back"))
}
