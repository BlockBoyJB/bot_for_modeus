package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewRouter(b *bot.Bot, services *service.Services) {
	b.Use(LoggingMiddleware)

	cmd, cb, state := b.NewGroup(bot.OnCommand), b.NewGroup(bot.OnCallback), b.NewGroup(bot.OnState)

	cmd.AddRoute("/test", test)

	newUserRouter(cmd, cb, state, services.User, services.Parser)
	newSettingsRouter(cmd, cb, state, services.User, services.Parser)
	newFriendsRouter(cmd, cb, state, services.User, services.Parser)
	newStudentsRouter(cmd, state, services.Parser)
}

func test(c bot.Context) error {
	return c.SendMessage("ok")
}

func formatStudents(students []parser.ModeusUser) (string, [][]tgbotapi.InlineKeyboardButton) {
	text := "Вот все студенты, которых мне удалось найти:\n"
	for k, s := range students {
		text += fmt.Sprintf("\n\n%d\n", k+1) + fmt.Sprintf(foundUser, s.FullName, s.FlowCode, s.SpecialtyName, s.SpecialtyProfile)
	}
	kb := tgmodel.NumbersButtons(len(students), 3)
	return text, kb
}

func studentSchedule(ctx context.Context, parser parser.Parser, state, studentId string) (string, error) {
	var text string
	switch state {
	case "day_schedule":
		daySchedule, err := parser.DaySchedule(ctx, studentId)
		if err != nil {
			return "", err
		}
		if len(daySchedule) == 0 {
			text = "На сегодня занятий нет!"
			break
		}
		text = "Расписание на сегодня:"
		for _, lesson := range daySchedule {
			text += "\n" + fmt.Sprintf(currentLesson, lesson.Subject, lesson.Name, lesson.Type, lesson.Time, lesson.AuditoriumNum, lesson.BuildingAddr, lesson.Lector) + "\n"
		}
	case "week_schedule":
		weekSchedule, err := parser.WeekSchedule(ctx, studentId)
		if err != nil {
			return "", err
		}
		text = "Расписание на неделю:"
		for d := 1; d <= 6; d++ {
			text += fmt.Sprintf("\nРасписание на %s:", dates[d])
			if len(weekSchedule[d]) == 0 {
				text += "\nЗанятий нет\n"
				continue
			}
			for _, lesson := range weekSchedule[d] {
				text += "\n" + fmt.Sprintf(currentLesson, lesson.Subject, lesson.Name, lesson.Type, lesson.Time, lesson.AuditoriumNum, lesson.BuildingAddr, lesson.Lector) + "\n"
			}
		}
	default:
		text = "Ой! Произошла ошибка при сборе данных"
	}
	return text, nil
}
