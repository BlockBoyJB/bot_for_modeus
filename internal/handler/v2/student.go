package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/pkg/bot"
	"fmt"
	"strconv"
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

	b.AddTree(bot.OnCallback, "/student/action/:schedule_id", r.callbackChooseOtherStudentActionBack)
	b.AddTree(bot.OnCallback, "/student/:type/:date/:schedule_id", r.callbackOtherStudentSchedule)
}

func (r *studentRouter) cmdOtherStudent(c bot.Context) error {
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
	if len(c.Text()) > 200 {
		return ErrIncorrectInput
	}
	students, err := r.parser.FindStudents(c.Context(), c.Text())
	if err != nil {
		return err
	}
	if err = c.SetData("other_students", students); err != nil {
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
	if err := c.GetData("other_students", &students); err != nil {
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
	var students []parser.Student
	if err := c.GetData("other_students", &students); err != nil {
		return err
	}
	cb := c.Update().CallbackQuery
	if cb == nil {
		return c.SendMessage(txtWarn)
	}
	num, err := strconv.Atoi(cb.Data)
	if err != nil || num > len(students) {
		return c.SendMessage(txtWarn)
	}
	s := students[num-1]

	return c.EditMessageWithInlineKB(fmt.Sprintf(txtChooseOtherStudentAction, s.FullName), tgmodel.OtherStudentButtons(s.ScheduleId))
}

func (r *studentRouter) callbackChooseOtherStudentActionBack(c bot.Context) error {
	scheduleId := c.Param("schedule_id")
	fullName, err := getFullName(c, r.parser, scheduleId)
	if err != nil {
		return err
	}
	return c.EditMessageWithInlineKB(fmt.Sprintf(txtChooseOtherStudentAction, fullName), tgmodel.OtherStudentButtons(scheduleId))
}

func (r *studentRouter) callbackOtherStudentSchedule(c bot.Context) error {
	return studentSchedule(c, r.parser, "student", tgmodel.BackButton("/student/action/"+c.Param("schedule_id")))
}
