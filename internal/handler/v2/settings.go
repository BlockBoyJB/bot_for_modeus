package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"strings"
)

type settingsRouter struct {
	userService service.User
	parser      parser.Parser
}

func newSettingsRouter(cmd, cb, state *bot.Group, userService service.User, parser parser.Parser) {
	r := settingsRouter{
		userService: userService,
		parser:      parser,
	}
	cmd.AddRoute("/settings", r.cmdSettings)
	state.AddRoute(stateAddLoginPassword, r.stateAddLoginPassword)
	cb.AddRoute("/add_login_password", r.callbackAddLoginPassword)
	cb.AddRoute("/update_full_name", r.callbackUpdateFullName)
}

func (r *settingsRouter) cmdSettings(c bot.Context) error {
	return c.SendMessageWithInlineKB(txtSettings, tgmodel.SettingsButtons)
}

func (r *settingsRouter) callbackAddLoginPassword(c bot.Context) error {
	if err := c.SendMessage(txtAddLoginPassword); err != nil {
		return err
	}
	return c.SetState(stateAddLoginPassword)
}

func (r *settingsRouter) stateAddLoginPassword(c bot.Context) error {
	defer func() { _ = c.Clear() }()
	data := strings.Fields(c.Text())
	if len(data) != 2 {
		return c.SendMessage(txtIncorrectLoginPassword)
	}
	login, password := data[0], data[1]
	if err := r.userService.UpdateLoginPassword(c.Context(), c.UserId(), login, password); err != nil {
		return err
	}
	return c.SendMessage("Логин и пароль успешно добавлены!\n" + txtDefault)
}

func (r *settingsRouter) callbackUpdateFullName(c bot.Context) error {
	if err := c.SendMessage("Введите новое ФИО без ошибок"); err != nil {
		return err
	}
	return c.SetState(stateInputFullName)
}
