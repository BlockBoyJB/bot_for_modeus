package v2

import (
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
)

func NewHandler(b *bot.Bot, services *service.Services) {
	b.PreUse(errorMiddleware)
	b.Use(recoverMiddleware, loggingMiddleware())

	b.Command("/test", test)

	newHelpRouter(b, services.Parser)
	newStudentRouter(b, services.Parser)
	newUserRouter(b, services.User, services.Parser)
	newFriendsRouter(b, services.User, services.Parser)
	newScheduleRouter(b, services.User, services.Parser)
	newSettingsRouter(b, services.User, services.Parser)
}

func test(c bot.Context) error {
	return c.SendMessage("ok")
}
