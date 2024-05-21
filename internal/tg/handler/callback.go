package handler

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/service"
	"errors"
)

const (
	txtAddLoginPassword    = "Пожалуйста, укажите через пробел сначала логин, потом пароль от учетной записи модеуса"
	txtUpdateFullName      = "Пожалуйста, введите без ошибок полностью новое значение для ФИО (как указано в модеусе)"
	txtDeleteLoginPassword = "Логин и пароль успешно удалены!\n" + txtDefault
)

func (h *Handler) callbackHandler(msg tgmodel.Message) error {
	switch msg.Text {
	case "/add_login_password":
		return h.callbackAddLoginPassword(msg)
	case "/update_full_name":
		return h.callbackUpdateFullName(msg)
	case "/delete_login_password":
		return h.callbackDeleteLoginPassword(msg)
	default:
		state := h.storage.getState(msg.UserId) // может быть какое-то нажатие на инлайн кнопку
		if len(state) != 0 {
			return h.stateHandler(state, msg)
		}
	}
	return nil
}

func (h *Handler) callbackAddLoginPassword(msg tgmodel.Message) error {
	h.storage.setState(msg.UserId, stateAddLoginPassword)
	return h.bot.SendMessage(msg.UserId, txtAddLoginPassword)
}

func (h *Handler) callbackUpdateFullName(msg tgmodel.Message) error {
	h.storage.setState(msg.UserId, stateUpdateFullName)
	return h.bot.SendMessage(msg.UserId, txtUpdateFullName)
}

func (h *Handler) callbackDeleteLoginPassword(msg tgmodel.Message) error {
	if err := h.services.DeleteLoginPassword(h.ctx, msg.UserId); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return h.bot.SendMessage(msg.UserId, txtUserNotFound)
		}
		return h.bot.SendMessage(msg.UserId, txtUnknownError)
	}
	return h.bot.SendMessage(msg.UserId, txtDeleteLoginPassword)
}
