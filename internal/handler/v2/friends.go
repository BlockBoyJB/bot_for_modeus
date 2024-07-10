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

type friendsRouter struct {
	user   service.User
	parser parser.Parser
}

func newFriendsRouter(cmd, cb, state *bot.Group, user service.User, parser parser.Parser) {
	r := friendsRouter{
		user:   user,
		parser: parser,
	}

	cmd.AddRoute("/friends", r.cmdFriends)
	cb.AddRoute("/add_friend", r.callbackAddFriend)
	state.AddRoute(stateAddFriend, r.stateAddFriend)
	state.AddRoute(stateChooseFriend, r.stateChooseFriend)
	state.AddRoute(stateChooseFindFriend, r.stateChooseFindFriend)
	state.AddRoute(stateChooseFriendAction, r.stateChooseFriendAction)
}

func (r *friendsRouter) cmdFriends(c bot.Context) error {
	user, err := r.user.FindUser(c.Context(), c.UserId())
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
	if err = c.SendMessageWithInlineKB(text, tgmodel.FriendsButtons(user.Friends, 1)); err != nil {
		return err
	}
	return c.SetState(stateChooseFriend)
}

func (r *friendsRouter) stateChooseFriend(c bot.Context) error {
	if err := c.UpdateData("friend", c.Text()); err != nil {
		return err
	}
	if err := c.SendMessageWithInlineKB(txtChooseSchedule, tgmodel.ChooseFriendsAction); err != nil {
		return err
	}
	return c.SetState(stateChooseFriendAction)
}

func (r *friendsRouter) callbackAddFriend(c bot.Context) error {
	if err := c.SendMessage("Введите ФИО друга, которого хотите добавить"); err != nil {
		return err
	}
	return c.SetState(stateAddFriend)
}

func (r *friendsRouter) stateAddFriend(c bot.Context) error {
	friends, err := r.parser.FindAllUsers(c.Context(), c.Text())
	if err != nil {
		if errors.Is(err, parser.ErrStudentsNotFound) {
			return c.SendMessage(fmt.Sprintf(txtStudentNotFound, c.Text()))
		}
		return err
	}
	if err = c.UpdateData("friends", friends); err != nil {
		return err
	}
	text, kb := formatStudents(friends)
	if err = c.SendMessageWithInlineKB(text, kb); err != nil {
		return err
	}
	return c.SetState(stateChooseFindFriend)
}

func (r *friendsRouter) stateChooseFindFriend(c bot.Context) error {
	var friends []parser.ModeusUser
	if err := c.GetData("friends", &friends); err != nil {
		return err
	}
	num, err := strconv.Atoi(c.Text())
	if err != nil {
		return err
	}
	friend := friends[num-1]
	if err = r.user.AddFriend(c.Context(), c.UserId(), friend.FullName, friend.ScheduleId); err != nil {
		return err
	}
	if err = c.UpdateData("friend", friend.ScheduleId); err != nil {
		return err
	}
	if err = c.SendMessageWithInlineKB("Друг успешно сохранен!\n"+txtChooseSchedule, tgmodel.ChooseFriendsAction); err != nil {
		return err
	}
	return c.SetState(stateChooseFriendAction)
}

func (r *friendsRouter) stateChooseFriendAction(c bot.Context) error {
	defer func() { _ = c.Clear() }()
	var friendId string
	err := c.GetData("friend", &friendId)
	if err != nil {
		return err
	}
	var text string
	if c.Text() == "delete_friend" {
		if err = r.user.DeleteFriend(c.Context(), c.UserId(), friendId); err != nil {
			return err
		}
		text = "Друг успешно удален\n" + txtDefault
	} else {
		text, err = studentSchedule(c.Context(), r.parser, c.Text(), friendId)
		if err != nil {
			return err
		}
	}
	return c.SendMessage(text)
}
