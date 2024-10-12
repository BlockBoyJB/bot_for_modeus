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
	b.Message(tgmodel.FriendsButton, r.cmdFriends)
	b.Callback("/choose_friend_back", r.callbackChooseFriendBack)
	b.State(stateChooseFriend, r.stateChooseFriend)
	b.AddTree(bot.OnCallback, "/friends/action/:schedule_id", r.callbackChooseFriendActionBack)
	b.AddTree(bot.OnCallback, "/friends/delete/:schedule_id", r.callbackDeleteFriend)
	b.AddTree(bot.OnCallback, "/friends/:type/:date/:schedule_id", r.callbackFriendsSchedule)

	b.Callback("/add_friend", r.callbackAddFriend)
	b.State(stateAddFriend, r.stateAddFriend)
	b.State(stateChooseFindFriend, r.stateChooseFindFriend)
}

func (r *friendsRouter) cmdFriends(c bot.Context) error {
	user, err := r.user.Find(c.Context(), c.UserId())
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.SendMessage(txtUserNotFound)
		}
		return err
	}
	text := txtFriends
	if len(user.Friends) == 0 {
		text = "Ой! Кажется, у Вас ни одного сохраненного друга!"
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
	text := txtFriends
	if len(friends) == 0 {
		text = "Ой! Кажется, у Вас ни одного сохраненного друга!"
	}
	if err := c.EditMessageWithInlineKB(text, tgmodel.FriendsButtons(friends)); err != nil {
		return err
	}
	return c.SetState(stateChooseFriend)
}

func (r *friendsRouter) stateChooseFriend(c bot.Context) error {
	var friends map[string]string
	if err := c.GetData("friends", &friends); err != nil {
		return err
	}
	fullName, ok := friends[c.Text()]
	if !ok { // Париться с проверкой, что это коллбэк нет смысла. Если захочет ввести свой id сам, пусть делает. Главное чтобы гадости не писали)
		return c.SendMessage(txtWarn)
	}

	return c.EditMessageWithInlineKB(fmt.Sprintf(txtChooseFriendAction, fullName), tgmodel.ChooseFriendAction(c.Text()))
}

func (r *friendsRouter) callbackChooseFriendActionBack(c bot.Context) error {
	scheduleId := c.Param("schedule_id")
	fullName, err := getFullName(c, scheduleId)
	if err != nil {
		return err
	}
	return c.EditMessageWithInlineKB(fmt.Sprintf(txtChooseFriendAction, fullName), tgmodel.ChooseFriendAction(scheduleId))
}

func (r *friendsRouter) callbackDeleteFriend(c bot.Context) error {
	scheduleId := c.Param("schedule_id")

	err := r.user.DeleteFriend(c.Context(), service.FriendInput{
		UserId:     c.UserId(),
		ScheduleId: scheduleId,
	})
	if err != nil {
		return err
	}

	var friends map[string]string
	if err = c.GetData("friends", &friends); err != nil {
		return err
	}
	fullName := friends[scheduleId]
	delete(friends, scheduleId)
	if err = c.SetData("friends", friends); err != nil {
		return err
	}

	return c.EditMessage(fmt.Sprintf("<b>%s</b> удален из друзей!", fullName))
}

func (r *friendsRouter) callbackFriendsSchedule(c bot.Context) error {
	return studentSchedule(c, r.parser, "friends", tgmodel.BackButton("/friends/action/"+c.Param("schedule_id")))
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
	if err = c.DelData("students"); err != nil {
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
	friends[s.ScheduleId] = s.FullName
	if err = c.SetData("friends", friends); err != nil {
		return err
	}
	if err = addFullName(c, s.ScheduleId, s.FullName); err != nil {
		return err
	}

	kb := [][]tgbotapi.InlineKeyboardButton{tgmodel.ChooseFriendAction(s.ScheduleId)[0][:2]}
	return c.EditMessageWithInlineKB(fmt.Sprintf("<b>%s</b> добавлен в друзья!\nВыберите действие", s.FullName), kb)
}
