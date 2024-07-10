package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/pkg/bot"
	"errors"
	"fmt"
	"strconv"
)

type studentsRouter struct {
	parser parser.Parser
}

func newStudentsRouter(cmd, state *bot.Group, parser parser.Parser) {
	r := studentsRouter{parser: parser}

	cmd.AddRoute("/other_student", r.cmdOtherStudent)
	state.AddRoute(stateInputOtherStudent, r.stateInputOtherStudent)
	state.AddRoute(stateChooseOtherStudent, r.stateChooseOtherStudent)
	state.AddRoute(stateChooseStudentAction, r.stateChooseStudentAction)
}

func (r *studentsRouter) cmdOtherStudent(c bot.Context) error {
	if err := c.SendMessage(txtInputOtherStudent); err != nil {
		return err
	}
	return c.SetState(stateInputOtherStudent)
}

func (r *studentsRouter) stateInputOtherStudent(c bot.Context) error {
	students, err := r.parser.FindAllUsers(c.Context(), c.Text())
	if err != nil {
		if errors.Is(err, parser.ErrStudentsNotFound) {
			return c.SendMessage(fmt.Sprintf(txtStudentNotFound, c.Text()))
		}
		return err
	}
	if err = c.UpdateData("students", students); err != nil {
		return err
	}
	text, kb := formatStudents(students)
	if err = c.SendMessageWithInlineKB(text, kb); err != nil {
		return err
	}
	return c.SetState(stateChooseOtherStudent)
}

func (r *studentsRouter) stateChooseOtherStudent(c bot.Context) error {
	var students []parser.ModeusUser
	if err := c.GetData("students", &students); err != nil {
		return err
	}
	num, err := strconv.Atoi(c.Text())
	if err != nil {
		return err
	}
	student := students[num-1]
	if err = c.UpdateData("student", student.ScheduleId); err != nil {
		return err
	}
	if err = c.SendMessageWithInlineKB(txtChooseStudentAction, tgmodel.OtherStudentButtons); err != nil {
		return err
	}
	return c.SetState(stateChooseStudentAction)
}

func (r *studentsRouter) stateChooseStudentAction(c bot.Context) error {
	defer func() { _ = c.Clear() }()
	var studentId string
	if err := c.GetData("student", &studentId); err != nil {
		return err
	}
	text, err := studentSchedule(c.Context(), r.parser, c.Text(), studentId)
	if err != nil {
		return err
	}
	return c.SendMessage(text)
}
