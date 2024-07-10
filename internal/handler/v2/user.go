package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"errors"
	"fmt"
	"strconv"
)

const (
	currentLesson        = "Предмет: %s\nПара: %s\nТип занятия: %s\nВремя: %s\nАудитория: %s\nАдрес: %s\nПреподаватель: %s"
	currentSubjectGrades = "Предмет: %s\nТекущий результат: %s\nИтог модуля: %s\nПроцент Посещения: %s\nПроцент Пропуска: %s\nПроцент Не отмечено: %s"
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

type userRouter struct {
	user   service.User
	parser parser.Parser
}

func newUserRouter(cmd, cb, state *bot.Group, user service.User, parser parser.Parser) {
	r := userRouter{
		user:   user,
		parser: parser,
	}

	cmd.AddRoute("/start", r.cmdStart)
	cmd.AddRoute("/stop", r.cmdStop)
	cmd.AddRoute("/day_schedule", r.cmdDaySchedule)
	cmd.AddRoute("/week_schedule", r.cmdWeekSchedule)
	cmd.AddRoute("/grades", r.cmdGrades)
	cmd.AddRoute("/help", r.cmdHelp)
	state.AddRoute(stateActionAfterCreate, r.stateActionAfterCreate)
	state.AddRoute(stateInputFullName, r.stateInputFullName)
	state.AddRoute(stateChooseUser, r.stateChooseUser)
	state.AddRoute(stateDeleteUser, r.stateDeleteUser)
	cmd.AddRoute("/me", r.cmdMe)
	cb.AddRoute("/about_me", r.callbackAboutMe)
	cb.AddRoute("/ratings", r.callbackRatings)
}

func (r *userRouter) cmdStart(c bot.Context) error {
	if err := c.SetState(stateInputFullName); err != nil {
		return err
	}
	return c.SendMessage(txtStart)
}

func (r *userRouter) stateInputFullName(c bot.Context) error {
	users, err := r.parser.FindAllUsers(c.Context(), c.Text())
	if err != nil {
		if errors.Is(err, parser.ErrStudentsNotFound) {
			return c.SendMessage(fmt.Sprintf(txtStudentNotFound, c.Text()))
		}
		return err
	}
	if err = c.UpdateData("users", users); err != nil {
		return err
	}
	text, kb := formatStudents(users)
	if err = c.SendMessageWithInlineKB(text, kb); err != nil {
		return err
	}
	return c.SetState(stateChooseUser)
}

func (r *userRouter) stateChooseUser(c bot.Context) error {
	var users []parser.ModeusUser
	if err := c.GetData("users", &users); err != nil {
		return err
	}
	num, err := strconv.Atoi(c.Text())
	if err != nil {
		return err
	}
	user := users[num-1]
	input := service.UserInput{
		UserId:     c.UserId(),
		FullName:   user.FullName,
		ScheduleId: user.ScheduleId,
		GradesId:   user.GradesId,
	}
	if err = r.user.CreateUser(c.Context(), input); err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			if err = r.user.UpdateUserInformation(c.Context(), input); err != nil {
				if errors.Is(err, service.ErrUserNotFound) { // это вообще реально???
					return c.SendMessage(txtUserNotFound)
				}
			}
			return c.SendMessage("Информация о пользователе успешно обновлена!\n" + txtDefault)
		}
		return err
	}
	if err = c.SendMessageWithInlineKB(txtUserCreated, tgmodel.YesOrNoButtons); err != nil {
		return err
	}
	return c.SetState(stateActionAfterCreate)
}

func (r *userRouter) stateActionAfterCreate(c bot.Context) error {
	if c.Text() == "да" {
		if err := c.SendMessage(txtAddLoginPassword); err != nil {
			return err
		}
		return c.SetState(stateAddLoginPassword)
	}
	_ = c.Clear()
	return c.SendMessage("Пользователь успешно создан!\n" + txtDefault)
}

func (r *userRouter) cmdDaySchedule(c bot.Context) error {
	user, err := r.user.FindUser(c.Context(), c.UserId())
	if err != nil {
		return err
	}
	schedule, err := r.parser.DaySchedule(c.Context(), user.ScheduleId)
	if err != nil {
		return err
	}
	if len(schedule) == 0 {
		return c.SendMessage("На сегодня занятий нет!")
	}
	text := "Расписание на сегодня:"
	for _, lesson := range schedule {
		text += "\n" + fmt.Sprintf(currentLesson, lesson.Subject, lesson.Name, lesson.Type, lesson.Time, lesson.AuditoriumNum, lesson.BuildingAddr, lesson.Lector) + "\n"
	}
	return c.SendMessage(text)
}

func (r *userRouter) cmdWeekSchedule(c bot.Context) error {
	user, err := r.user.FindUser(c.Context(), c.UserId())
	if err != nil {
		return err
	}
	weekSchedule, err := r.parser.WeekSchedule(c.Context(), user.ScheduleId)
	if err != nil {
		return err
	}
	text := "Расписание на неделю:"
	for d := 1; d <= 6; d++ {
		text += fmt.Sprintf("\nРасписание на %s:", dates[d])
		if len(weekSchedule[d]) == 0 {
			text += "\nЗанятий нет\n"
			continue
		}
		for _, lesson := range weekSchedule[d] {
			text += "\n" + fmt.Sprintf(currentLesson, lesson.Subject, lesson.Name, lesson.Type, lesson.Time, lesson.AuditoriumNum, lesson.BuildingAddr, lesson.Lector) + "\n"
		}
	}
	return c.SendMessage(text)
}

func (r *userRouter) cmdGrades(c bot.Context) error {
	user, err := r.user.FindUser(c.Context(), c.UserId())
	if err != nil {
		return err
	}
	if user.Login == "" || user.Password == "" {
		return c.SendMessage(txtRequiredLoginPass)
	}
	grades, err := r.parser.CourseTotalGrades(c.Context(), parser.GradesInput{
		GradesId:   user.GradesId,
		ScheduleId: user.ScheduleId,
		Login:      user.Login,
		Password:   user.Password,
	})
	if err != nil {
		return err
	}
	text := "Вот все оценки по изучаемым дисциплинам в семестре:"
	for _, grade := range grades {
		text += "\n" + fmt.Sprintf(currentSubjectGrades, grade.Subject, grade.CurrentResult, grade.CourseUnitResult, grade.PresentRate, grade.AbsentRate, grade.UndefinedRate) + "\n"
	}
	return c.SendMessage(text)
}

func (r *userRouter) cmdStop(c bot.Context) error {
	if err := c.SendMessageWithInlineKB(txtConfirmDelete, tgmodel.YesOrNoButtons); err != nil {
		return err
	}
	return c.SetState(stateDeleteUser)
}

func (r *userRouter) stateDeleteUser(c bot.Context) error {
	defer func() { _ = c.Clear() }()
	if c.Text() == "да" {
		if err := r.user.DeleteUser(c.Context(), c.UserId()); err != nil {
			if errors.Is(err, service.ErrUserNotFound) {
				return c.SendMessage(txtUserNotFound)
			}
			return err
		}
		return c.SendMessage(txtUserDeleted)
	}
	return c.SendMessage("Пользователь не удален!\n" + txtDefault)
}

func (r *userRouter) cmdMe(c bot.Context) error {
	return c.SendMessageWithInlineKB(txtMyProfile, tgmodel.MyProfileButtons)
}

func (r *userRouter) callbackAboutMe(c bot.Context) error {
	u, err := r.user.FindUser(c.Context(), c.UserId())
	if err != nil {
		return err
	}
	user, err := r.parser.FindUserById(c.Context(), u.ScheduleId)
	if err != nil {
		return err
	}
	text := "Вот информация о Вашем направлении обучения:\n"
	text += fmt.Sprintf(foundUser, user.FullName, user.FlowCode, user.SpecialtyName, user.SpecialtyProfile)
	return c.SendMessage(text)
}

func (r *userRouter) callbackRatings(c bot.Context) error {
	user, err := r.user.FindUser(c.Context(), c.UserId())
	if err != nil {
		return err
	}
	if user.Login == "" || user.Password == "" {
		return c.SendMessage(txtRequiredLoginPass)
	}
	cgpa, APRs, err := r.parser.UserRatings(c.Context(), parser.GradesInput{
		GradesId:   user.GradesId,
		ScheduleId: user.ScheduleId,
		Login:      user.Login,
		Password:   user.Password,
	})
	if err != nil {
		return err
	}
	const academicPeriod = "Период: %s\nGPA: %s\nПроцент посещений: %s\nПроцент пропусков: %s\nПроцент без отметки: %s"
	text := "Вот информация о Ваших рейтингах:\nТекущий CGPA: " + cgpa + "\n"
	for i := 1; i <= len(APRs); i++ {
		text += "\n" + fmt.Sprintf(academicPeriod, APRs[i].Name, APRs[i].GPA, APRs[i].PresentRate, APRs[i].AbsentRate, APRs[i].UndefinedRate) + "\n"
	}
	return c.SendMessage(text)
}

func (r *userRouter) cmdHelp(c bot.Context) error {
	_ = c.Clear()
	return c.SendMessage(txtHelp)
}
