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
	b.Command("/kb", r.cmdKB)
	b.Callback("/cmd_start_back", r.callbackStartBack)
	b.State(stateInputFullName, r.stateInputFullName)
	b.State(stateChooseStudent, r.stateChooseStudent)
	b.State(stateActionAfterCreate, r.stateActionAfterCreate)
	b.State(stateAddLoginPasswordAfterCreate, r.stateAddLoginPasswordAfterCreate)
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

func (r *userRouter) cmdKB(c bot.Context) error {
	return c.SendMessageWithReplyKB("👋", tgmodel.RowCommands)
}

func (r *userRouter) callbackStartBack(c bot.Context) error {
	if err := c.EditMessage("Пожалуйста, введите Ваше ФИО как указано в системе модеус"); err != nil {
		return err
	}
	return c.SetState(stateInputFullName)
}

func (r *userRouter) stateInputFullName(c bot.Context) error {
	if len(c.Text()) > 200 {
		return ErrIncorrectInput
	}
	students, err := r.parser.FindStudents(c.Text())
	if err != nil {
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
			if e := c.DelData("grades_input"); e != nil {
				return e
			}
			if e := r.user.UpdateInfo(c.Context(), input); e != nil {
				return e
			}
			_, _ = lookupGI(c, r.user, false) // перезаписываем grades_input в кэше
			return c.EditMessage("Информация о пользователе успешно обновлена!")
		}
		return err
	}

	if err = c.EditMessageWithInlineKB(txtUserCreated, tgmodel.YesOrNoButtons); err != nil {
		return err
	}
	return c.SetState(stateActionAfterCreate)
}

func (r *userRouter) stateActionAfterCreate(c bot.Context) error {
	if err := c.DelData("students", "state"); err != nil {
		return err
	}
	if c.Text() == "да" {
		if err := c.EditMessage(txtAddLoginPassword); err != nil {
			return err
		}
		return c.SetState(stateAddLoginPasswordAfterCreate)
	}
	return c.EditMessageWithInlineKB("Пользователь успешно создан!\n\n"+txtUserAfterCreate, tgmodel.GuideButtons)
}

func (r *userRouter) stateAddLoginPasswordAfterCreate(c bot.Context) error {
	if err := addLoginPassword(c, r.user); err != nil {
		return err
	}
	return c.SendMessageWithInlineKB("Логин и пароль успешно сохранены!\n\n"+txtUserAfterCreate, tgmodel.GuideButtons)
}

func (r *userRouter) cmdStop(c bot.Context) error {
	if err := c.SendMessageWithInlineKB(txtConfirmDelete, tgmodel.YesOrNoButtons); err != nil {
		return err
	}
	return c.SetState(stateConfirmDelete)
}

func (r *userRouter) stateConfirmDelete(c bot.Context) error {
	if c.Text() != "да" {
		return c.EditMessage("Пользователь не удален!")
	}
	u, err := r.user.Find(c.Context(), c.UserId())
	if err != nil {
		return err
	}
	if u.Login != "" {
		if err = r.parser.DeleteToken(u.Login); err != nil {
			return err
		}
	}
	if err = c.Clear(); err != nil {
		return err
	}
	if err = r.user.Delete(c.Context(), c.UserId()); err != nil {
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
	gi, err := lookupGI(c, r.user, false)
	if err != nil {
		return err
	}
	info, err := r.parser.FindStudentById(gi.ScheduleId)
	if err != nil {
		return err
	}
	text := "Вот информация о Вашем направлении обучения:\n\n"
	text += fmt.Sprintf(formatStudent, info.FullName, info.SpecialtyName, info.SpecialtyProfile, info.FlowCode)
	return c.EditMessageWithInlineKB(text, kbMeBack)
}

func (r *userRouter) callbackRatings(c bot.Context) error {
	gi, err := lookupGI(c, r.user, true)
	if err != nil {
		return err
	}
	if gi.Login == "" || gi.Password == "" {
		kb := append(tgmodel.GradesLink, kbMeBack...)
		return c.EditMessageWithInlineKB(txtRequiredLoginPass, kb)
	}

	cgpa, ratings, err := r.parser.Ratings(gi)
	if err != nil {
		return err
	}

	text := "Вот информация о Ваших рейтингах:\nТекущий CGPA: " + cgpa + "\n"
	for _, sem := range ratings {
		text += "\n" + fmt.Sprintf(formatSemester, sem.Name, sem.GPA, sem.PresentRate, sem.AbsentRate, sem.UndefinedRate) + "\n"
	}
	return c.EditMessageWithInlineKB(text, kbMeBack)
}
