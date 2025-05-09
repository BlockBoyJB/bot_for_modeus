package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type friendsRouter struct {
	user   service.User
	parser parser.Parser
}

func newFriendsRouter(b bot.Router, user service.User, parser parser.Parser) {
	r := &friendsRouter{
		user:   user,
		parser: parser,
	}

	b = b.Group(metricsMiddleware("friends"))

	b.Command("/friends", r.cmdFriends)
	b.Message(tgmodel.FriendsButton, r.cmdFriends)
	b.Callback("/choose_friend_back", r.callbackChooseFriendBack)
	b.AddTree(bot.OnCallback, "/friends/choose/:schedule_id", r.callbackChooseFriend)
	b.AddTree(bot.OnCallback, "/friends/delete/:schedule_id", r.callbackDeleteFriend)
	b.AddTree(bot.OnCallback, "/friends/:type/:date/:schedule_id", r.callbackFriendsSchedule)

	b.Callback("/add_friend", r.callbackAddFriend)
	b.State(stateAddFriend, r.stateAddFriend)
	b.State(stateChooseFindFriend, r.stateChooseFindFriend)
}

func (r *friendsRouter) cmdFriends(c bot.Context) error {
	friends, err := lookupFriends(c, r.user)
	if err != nil {
		return err
	}
	text := txtFriends
	if len(friends) == 0 {
		text = "Ой! Кажется, у Вас ни одного сохраненного друга!"
	}

	return c.SendMessageWithInlineKB(text, friendsButtons(friends))
}

func (r *friendsRouter) callbackChooseFriendBack(c bot.Context) error {
	friends, err := lookupFriends(c, r.user)
	if err != nil {
		return err
	}
	text := txtFriends
	if len(friends) == 0 {
		text = "Ой! Кажется, у Вас ни одного сохраненного друга!"
	}
	return c.EditMessageWithInlineKB(text, friendsButtons(friends))
}

func (r *friendsRouter) callbackChooseFriend(c bot.Context) error {
	scheduleId := c.Param("schedule_id")
	fullName, err := getFullName(c, r.parser, scheduleId)
	if err != nil {
		return err
	}
	return c.EditMessageWithInlineKB(fmt.Sprintf(txtChooseFriendAction, fullName), tgmodel.ChooseFriendAction(scheduleId))
}

func (r *friendsRouter) callbackChooseFriendActionBack(c bot.Context) error {
	scheduleId := c.Param("schedule_id")
	fullName, err := getFullName(c, r.parser, scheduleId)
	if err != nil {
		return err
	}
	return c.EditMessageWithInlineKB(fmt.Sprintf(txtChooseFriendAction, fullName), tgmodel.ChooseFriendAction(scheduleId))
}

func (r *friendsRouter) callbackDeleteFriend(c bot.Context) error {
	scheduleId := c.Param("schedule_id")

	// Сначала надо удалить всех друзей из кэша, а потом из бд, чтобы гарантировать консистентность данных.
	// Есть риск того, что сломаться может где-то "посередине" и данные в кэше и бд будут расходиться.
	// Например, удалили из бд, но при изменении в кэше произошла ошибка и старое значение осталось - печальный исход.
	// Поэтому удаляем сразу всех

	if err := c.DelData("friends"); err != nil {
		return err
	}

	err := r.user.DeleteFriend(c.Context(), service.FriendInput{
		UserId:     c.UserId(),
		ScheduleId: scheduleId,
	})
	if err != nil {
		return err
	}

	fullName, err := getFullName(c, r.parser, scheduleId)
	if err != nil {
		return err
	}

	return c.EditMessageWithInlineKB(fmt.Sprintf("<b>%s</b> удален из друзей!", fullName), tgmodel.BackButton("/choose_friend_back"))
}

func (r *friendsRouter) callbackFriendsSchedule(c bot.Context) error {
	return studentSchedule(c, r.parser, "friends", tgmodel.BackButton("/friends/choose/"+c.Param("schedule_id")))
}

func (r *friendsRouter) callbackAddFriend(c bot.Context) error {
	if err := c.EditMessageWithInlineKB("Введите ФИО друга, которого хотите добавить", tgmodel.BackButton("/choose_friend_back")); err != nil {
		return err
	}
	return c.SetState(stateAddFriend)
}

func (r *friendsRouter) stateAddFriend(c bot.Context) error {
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
	kb = append(kb, tgmodel.BackButton("/add_friend")...)
	if err = c.SendMessageWithInlineKB(text, kb); err != nil {
		return err
	}
	return c.SetState(stateChooseFindFriend)
}

func (r *friendsRouter) stateChooseFindFriend(c bot.Context) error {
	s, err := findStudent(c)
	if err != nil {
		return err
	}

	// Удаляем друзей из кэша, чтобы гарантировать консистентность (более подробно в функции callbackDeleteFriend)
	if err = c.DelData("students", "friends"); err != nil {
		return err
	}

	if err = r.user.AddFriend(c.Context(), service.FriendInput{
		UserId:     c.UserId(),
		FullName:   s.FullName,
		ScheduleId: s.ScheduleId,
	}); err != nil {
		return err
	}

	kb := [][]tgbotapi.InlineKeyboardButton{tgmodel.ChooseFriendAction(s.ScheduleId)[0][:2]}
	return c.EditMessageWithInlineKB(fmt.Sprintf("<b>%s</b> добавлен в друзья!\nВыберите действие", s.FullName), kb)
}
