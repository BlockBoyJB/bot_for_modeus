package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"errors"
	"fmt"
)

type userRouter struct {
	user   service.User
	parser parser.Parser
}

func newUserRouter(b *bot.Bot, user service.User, parser parser.Parser) {
	r := &userRouter{
		user:   user,
		parser: parser,
	}

	b.Command("/start", r.cmdStart)
	b.Callback("/cmd_start_back", r.callbackStartBack)
	b.State(stateInputFullName, r.stateInputFullName)
	b.State(stateChooseStudent, r.stateChooseStudent)
	b.State(stateActionAfterCreate, r.stateActionAfterCreate)
	b.Command("/stop", r.cmdStop)
	b.State(stateConfirmDelete, r.stateConfirmDelete)

	b.Command("/me", r.cmdMe)
	b.Message(tgmodel.MeButton, r.cmdMe)
	b.Callback("/me_back", r.callbackMeBack)
	b.Callback("/about_me", r.callbackAboutMe)
	b.Callback("/ratings", r.callbackRatings)
}

func (r *userRouter) cmdStart(c bot.Context) error {
	_ = c.Clear()
	if err := c.SetState(stateInputFullName); err != nil {
		return err
	}
	return c.SendMessageWithReplyKB(txtStart, tgmodel.RowCommands)
}

func (r *userRouter) callbackStartBack(c bot.Context) error {
	if err := c.EditMessage("Пожалуйста, введите Ваше ФИО как указано в системе модеус"); err != nil {
		return err
	}
	return c.SetState(stateInputFullName)
}

func (r *userRouter) stateInputFullName(c bot.Context) error {
	// тут находим всех пользователей с введенным фио
	students, err := r.parser.FindStudents(c.Context(), c.Text())
	if err != nil {
		if errors.Is(err, parser.ErrStudentsNotFound) {
			return c.SendMessage(fmt.Sprintf(txtStudentNotFound, c.Text()))
		}
		return err
	}
	if err = c.SetData("students", students); err != nil {
		return err
	}
	text, kb := formatStudents(students)
	// кнопка назад для того, чтобы заново ввести ФИО
	kb = append(kb, tgmodel.BackButton("/cmd_start_back")...)
	if err = c.SendMessageWithInlineKB(text, kb); err != nil {
		return err
	}
	return c.SetState(stateChooseStudent)
}

func (r *userRouter) stateChooseStudent(c bot.Context) error {
	s, err := findStudent(c)
	if err != nil {
		if errors.Is(err, ErrIncorrectInput) {
			return c.SendMessage(txtWarn)
		}
		return err
	}

	input := service.UserInput{
		UserId:     c.UserId(),
		FullName:   s.FullName,
		ScheduleId: s.ScheduleId,
		GradesId:   s.GradesId,
	}
	if err = r.user.Create(c.Context(), input); err != nil {
		// если пользователь уже существует, то просто обновляем информацию о нем
		if errors.Is(err, service.ErrUserAlreadyExists) {
			if e := r.user.UpdateInfo(c.Context(), input); e != nil {
				return e
			}
			return c.EditMessage("Информация о пользователе успешно обновлена!\n" + txtDefault)
		}
		return err
	}

	if err = c.EditMessageWithInlineKB(txtUserCreated, tgmodel.YesOrNoButtons); err != nil {
		return err
	}
	return c.SetState(stateActionAfterCreate)
}

func (r *userRouter) stateActionAfterCreate(c bot.Context) error {
	if c.Text() == "да" {
		if err := c.EditMessage(txtAddLoginPassword); err != nil {
			return err
		}
		return c.SetState(stateAddLoginPassword)
	}
	if err := c.DelData("students"); err != nil {
		return err
	}
	return c.EditMessage("Пользователь успешно создан!\n" + txtDefault)
}

func (r *userRouter) cmdStop(c bot.Context) error {
	if err := c.SendMessageWithInlineKB(txtConfirmDelete, tgmodel.YesOrNoButtons); err != nil {
		return err
	}
	return c.SetState(stateConfirmDelete)
}

func (r *userRouter) stateConfirmDelete(c bot.Context) error {
	if c.Text() != "да" {
		return c.EditMessage("Пользователь не удален!\n" + txtDefault)
	}
	u, err := r.user.Find(c.Context(), c.UserId())
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.EditMessage(txtUserNotFound)
		}
		return err
	}
	if err = r.user.Delete(c.Context(), c.UserId()); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.EditMessage(txtUserNotFound)
		}
		return err
	}
	if err = r.parser.DeleteToken(u.Login); err != nil {
		return err
	}
	if err = c.Clear(); err != nil {
		return err
	}
	return c.EditMessage(txtUserDeleted)
}

var (
	kbMeBack = tgmodel.BackButton("/me_back")
)

func (r *userRouter) cmdMe(c bot.Context) error {
	return c.SendMessageWithInlineKB(txtMyProfile, tgmodel.MyProfileButtons)
}

func (r *userRouter) callbackMeBack(c bot.Context) error {
	return c.EditMessageWithInlineKB(txtMyProfile, tgmodel.MyProfileButtons)
}

func (r *userRouter) callbackAboutMe(c bot.Context) error {
	u, err := r.user.Find(c.Context(), c.UserId())
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.EditMessage(txtUserNotFound)
		}
		return err
	}
	info, err := r.parser.FindStudentById(c.Context(), u.ScheduleId)
	if err != nil {
		return err
	}
	text := "Вот информация о Вашем направлении обучения:\n"
	text += fmt.Sprintf(formatStudent, info.FullName, info.SpecialtyName, info.SpecialtyProfile, info.FlowCode)
	return c.EditMessageWithInlineKB(text, kbMeBack)
}

func (r *userRouter) callbackRatings(c bot.Context) error {
	u, err := r.user.Find(c.Context(), c.UserId())
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.EditMessage(txtUserNotFound)
		}
		return err
	}
	if u.Login == "" || u.Password == "" {
		kb := append(tgmodel.GradesLink, kbMeBack...)
		return c.EditMessageWithInlineKB(txtRequiredLoginPass, kb)
	}

	cgpa, ratings, err := r.parser.Ratings(c.Context(), parser.GradesInput{
		Login:      u.Login,
		Password:   u.Password,
		ScheduleId: u.ScheduleId,
		GradesId:   u.GradesId,
	})
	if err != nil {
		return err
	}

	text := "Вот информация о Ваших рейтингах:\nТекущий CGPA: " + cgpa + "\n"
	for i := 1; i <= len(ratings); i++ {
		text += "\n" + fmt.Sprintf(formatSemester, ratings[i].Name, ratings[i].GPA, ratings[i].PresentRate, ratings[i].AbsentRate, ratings[i].UndefinedRate) + "\n"
	}
	return c.EditMessageWithInlineKB(text, kbMeBack)
}
