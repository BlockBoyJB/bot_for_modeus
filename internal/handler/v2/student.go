package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/pkg/bot"
	"errors"
	"fmt"
)

type studentRouter struct {
	parser parser.Parser
}

func newStudentRouter(b *bot.Bot, parser parser.Parser) {
	r := &studentRouter{
		parser: parser,
	}

	b.Command("/other_student", r.cmdOtherStudent)
	b.Message(tgmodel.OtherStudentButton, r.cmdOtherStudent)
	b.Callback("/other_student_back", r.callbackOtherStudentBack)
	b.State(stateInputOtherStudent, r.stateInputOtherStudent)
	b.Callback("/choose_other_student_back", r.callbackChooseOtherStudentBack)
	b.State(stateChooseOtherStudent, r.stateChooseOtherStudent)
	b.Callback("/choose_other_student_action_back", r.callbackChooseOtherStudentActionBack)
	b.State(stateChooseOtherStudentAction, r.stateChooseOtherStudentAction)
	b.State(stateOtherStudentSchedule, r.stateOtherStudentSchedule)
}

func (r *studentRouter) cmdOtherStudent(c bot.Context) error {
	_ = c.Clear()
	if err := c.SendMessage(txtInputOtherStudent); err != nil {
		return err
	}
	return c.SetState(stateInputOtherStudent)
}

func (r *studentRouter) callbackOtherStudentBack(c bot.Context) error {
	if err := c.EditMessage(txtInputOtherStudent); err != nil {
		return err
	}
	return c.SetState(stateInputOtherStudent)
}

func (r *studentRouter) stateInputOtherStudent(c bot.Context) error {
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
	kb = append(kb, tgmodel.BackButton("/other_student_back")...)
	if err = c.SendMessageWithInlineKB(text, kb); err != nil {
		return err
	}
	return c.SetState(stateChooseOtherStudent)
}

func (r *studentRouter) callbackChooseOtherStudentBack(c bot.Context) error {
	var students []parser.Student
	if err := c.GetData("students", &students); err != nil {
		return err
	}
	text, kb := formatStudents(students)
	kb = append(kb, tgmodel.BackButton("/other_student_back")...)
	if err := c.EditMessageWithInlineKB(text, kb); err != nil {
		return err
	}
	return c.SetState(stateChooseOtherStudent)
}

func (r *studentRouter) stateChooseOtherStudent(c bot.Context) error {
	s, err := findStudent(c)
	if err != nil {
		if errors.Is(err, ErrIncorrectInput) {
			return c.SendMessage(txtWarn)
		}
		return err
	}

	if err = c.SetData("schedule_id", s.ScheduleId); err != nil {
		return err
	}
	if err = c.SetData("full_name", s.FullName); err != nil {
		return err
	}

	if err = c.EditMessageWithInlineKB(fmt.Sprintf(txtChooseOtherStudentAction, s.FullName), tgmodel.OtherStudentButtons); err != nil {
		return err
	}
	return c.SetState(stateChooseOtherStudentAction)
}

func (r *studentRouter) callbackChooseOtherStudentActionBack(c bot.Context) error {
	var fullName string
	if err := c.GetData("full_name", &fullName); err != nil {
		return err
	}
	if err := c.EditMessageWithInlineKB(fmt.Sprintf(txtChooseOtherStudentAction, fullName), tgmodel.OtherStudentButtons); err != nil {
		return err
	}
	return c.SetState(stateChooseOtherStudentAction)
}

func (r *studentRouter) stateChooseOtherStudentAction(c bot.Context) error {
	var scheduleId string
	if err := c.GetData("schedule_id", &scheduleId); err != nil {
		return err
	}
	var fullName string
	if err := c.GetData("full_name", &fullName); err != nil {
		return err
	}

	text, kb, err := studentCurrentSchedule(c, r.parser, scheduleId)
	if err != nil {
		return err
	}
	text = fmt.Sprintf(formatFullName, fullName) + text
	kb = append(kb, tgmodel.BackButton("/choose_other_student_action_back")...)
	if err = c.EditMessageWithInlineKB(text, kb); err != nil {
		return err
	}
	return c.SetState(stateOtherStudentSchedule)
}

func (r *studentRouter) stateOtherStudentSchedule(c bot.Context) error {
	return studentSchedule(c, r.parser, tgmodel.BackButton("/choose_other_student_action_back"))
}
