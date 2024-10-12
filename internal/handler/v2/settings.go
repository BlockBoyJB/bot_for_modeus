package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"errors"
	"strings"
)

var (
	kbSettingsBack = tgmodel.BackButton("/cmd_settings_callback")
)

type settingsRouter struct {
	user   service.User
	parser parser.Parser
}

func newSettingsRouter(b *bot.Bot, user service.User, parser parser.Parser) {
	r := &settingsRouter{
		user:   user,
		parser: parser,
	}

	b.Command("/settings", r.cmdSettings)
	b.Message(tgmodel.SettingsButton, r.cmdSettings)
	b.Callback("/cmd_settings_callback", r.callbackSettingsBack)
	b.Callback("/add_login_password", r.callbackAddLoginPassword)
	b.State(stateAddLoginPassword, r.stateAddLoginPassword)
	b.Callback("/update_full_name", r.callbackUpdateFullName)
}

func (r *settingsRouter) cmdSettings(c bot.Context) error {
	return c.SendMessageWithInlineKB(txtSettings, tgmodel.SettingsButtons)
}

func (r *settingsRouter) callbackSettingsBack(c bot.Context) error {
	_ = c.DelData("state")
	return c.EditMessageWithInlineKB(txtSettings, tgmodel.SettingsButtons)
}

func (r *settingsRouter) callbackAddLoginPassword(c bot.Context) error {
	if err := c.EditMessageWithInlineKB(txtAddLoginPassword, kbSettingsBack); err != nil {
		return err
	}
	return c.SetState(stateAddLoginPassword)
}

func (r *settingsRouter) stateAddLoginPassword(c bot.Context) error {
	data := strings.Fields(c.Text())
	if len(data) != 2 {
		return c.SendMessage(txtIncorrectLoginPassInput)
	}

	err := r.user.UpdateLoginPassword(c.Context(), service.UserLoginPasswordInput{
		UserId:   c.UserId(),
		Login:    data[0],
		Password: data[1],
	})
	if err != nil {
		if errors.Is(err, service.ErrUserIncorrectLogin) {
			return c.SendMessage(txtIncorrectLoginPassInput)
		}
		if errors.Is(err, service.ErrUserNotFound) {
			return c.SendMessage(txtUserNotFound)
		}
		return err
	}
	_ = c.DelData("state")
	if err = c.DelData("grades_input"); err != nil {
		return err
	}
	_, _ = lookupGI(c, r.user) // перезаписываем grades_input в кэше
	return c.SendMessageWithReplyKB("Логин и пароль успешно добавлены!", tgmodel.RowCommands)
}

func (r *settingsRouter) callbackUpdateFullName(c bot.Context) error {
	if err := c.EditMessageWithInlineKB("Введите новое ФИО без ошибок", kbSettingsBack); err != nil {
		return err
	}
	return c.SetState(stateInputFullName)
}
