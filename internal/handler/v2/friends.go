package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type friendsRouter struct {
	user   service.User
	parser parser.Parser
}

func newFriendsRouter(b *bot.Bot, user service.User, parser parser.Parser) {
	r := &friendsRouter{
		user:   user,
		parser: parser,
	}

	b.Command("/friends", r.cmdFriends)
	b.Callback("/choose_friend_back", r.callbackChooseFriendBack)
	b.State(stateChooseFriend, r.stateChooseFriend)
	b.Callback("/choose_friend_action_back", r.callbackChooseFriendActionBack)
	b.State(stateChooseFriendAction, r.stateChooseFriendAction)
	b.State(stateFriendSchedule, r.stateFriendSchedule)
	b.Callback("/add_friend", r.callbackAddFriend)
	b.State(stateAddFriend, r.stateAddFriend)
	b.State(stateChooseFindFriend, r.stateChooseFindFriend)
}

func (r *friendsRouter) cmdFriends(c bot.Context) error {
	_ = c.Clear()
	user, err := r.user.Find(c.Context(), c.UserId())
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.SendMessage(txtUserNotFound)
		}
		return err
	}
	text := txtChooseFriend
	if len(user.Friends) == 0 {
		text = "Ой! Кажется у Вас ни одного сохраненного друга!\nЧтобы добавить друга нажмите \"Добавить друга\""
	}

	if err = c.SetData("friends", user.Friends); err != nil {
		return err
	}

	if err = c.SendMessageWithInlineKB(text, tgmodel.FriendsButtons(user.Friends)); err != nil {
		return err
	}
	return c.SetState(stateChooseFriend)
}

func (r *friendsRouter) callbackChooseFriendBack(c bot.Context) error {
	var friends map[string]string
	if err := c.GetData("friends", &friends); err != nil {
		return err
	}
	text := txtChooseFriend
	if len(friends) == 0 {
		text = "Ой! Кажется у Вас ни одного сохраненного друга!\nЧтобы добавить друга нажмите \"Добавить друга\""
	}
	if err := c.EditMessageWithInlineKB(text, tgmodel.FriendsButtons(friends)); err != nil {
		return err
	}
	return c.SetState(stateChooseFriend)
}

func (r *friendsRouter) stateChooseFriend(c bot.Context) error {
	if err := c.SetData("schedule_id", c.Text()); err != nil {
		return err
	}
	if err := c.EditMessageWithInlineKB(txtChooseFriendAction, tgmodel.ChooseFriendAction); err != nil {
		return err
	}
	return c.SetState(stateChooseFriendAction)
}

func (r *friendsRouter) callbackChooseFriendActionBack(c bot.Context) error {
	if err := c.EditMessageWithInlineKB(txtChooseFriendAction, tgmodel.ChooseFriendAction); err != nil {
		return err
	}
	return c.SetState(stateChooseFriendAction)
}

func (r *friendsRouter) stateChooseFriendAction(c bot.Context) error {
	var scheduleId string
	err := c.GetData("schedule_id", &scheduleId)
	if err != nil {
		return err
	}
	if c.Text() == "delete_friend" {
		if err = r.user.DeleteFriend(c.Context(), service.FriendInput{
			UserId:     c.UserId(),
			ScheduleId: scheduleId,
		}); err != nil {
			return err
		}
		if err = c.EditMessage("Друг успешно удален\n" + txtDefault); err != nil {
			return err
		}
		return c.Clear()
	}

	text, kb, err := studentCurrentSchedule(c, r.parser, scheduleId)
	if err != nil {
		return err
	}
	kb = append(kb, tgmodel.BackButton("/choose_friend_action_back")...)
	if err = c.EditMessageWithInlineKB(text, kb); err != nil {
		return err
	}
	return c.SetState(stateFriendSchedule)
}

func (r *friendsRouter) stateFriendSchedule(c bot.Context) error {
	return studentSchedule(c, r.parser, tgmodel.BackButton("/choose_friend_action_back"))
}

func (r *friendsRouter) callbackAddFriend(c bot.Context) error {
	if err := c.EditMessageWithInlineKB("Введите ФИО друга, которого хотите добавить", tgmodel.BackButton("/choose_friend_back")); err != nil {
		return err
	}
	return c.SetState(stateAddFriend)
}

func (r *friendsRouter) stateAddFriend(c bot.Context) error {
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
	kb = append(kb, tgmodel.BackButton("/add_friend")...)
	if err = c.SendMessageWithInlineKB(text, kb); err != nil {
		return err
	}
	return c.SetState(stateChooseFindFriend)
}

func (r *friendsRouter) stateChooseFindFriend(c bot.Context) error {
	s, err := findStudent(c)
	if err != nil {
		if errors.Is(err, ErrIncorrectInput) {
			return c.SendMessage(txtWarn)
		}
		return err
	}

	if err = r.user.AddFriend(c.Context(), service.FriendInput{
		UserId:     c.UserId(),
		FullName:   s.FullName,
		ScheduleId: s.ScheduleId,
	}); err != nil {
		return err
	}

	var friends map[string]string
	if err = c.GetData("friends", &friends); err != nil {
		return err
	}
	_ = c.Clear()
	friends[s.ScheduleId] = s.FullName
	if err = c.SetData("friends", friends); err != nil {
		return err
	}
	if err = c.SetData("schedule_id", s.ScheduleId); err != nil {
		return err
	}
	kb := [][]tgbotapi.InlineKeyboardButton{tgmodel.ChooseFriendAction[0][:2]}
	if err = c.EditMessageWithInlineKB("Друг успешно сохранен!\n"+txtChooseFriendAction, kb); err != nil {
		return err
	}
	return c.SetState(stateChooseFriendAction)
}
