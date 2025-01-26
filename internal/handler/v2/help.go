package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/pkg/bot"
	"fmt"
	"os"
)

var kbHelpBack = tgmodel.BackButton("/help_back")

type helpRouter struct {
	parser parser.Parser
}

func newHelpRouter(b *bot.Bot, parser parser.Parser) {
	r := &helpRouter{
		parser: parser,
	}

	b.Command("/help", r.cmdHelp)
	b.Message(tgmodel.HelpButton, r.cmdHelp)
	b.Callback("/help_back", r.callbackHelpBack)
	b.Callback("/help_schedule", r.callbackSchedule)
	b.Callback("/help_grades", r.callbackGrades)
	b.Callback("/help_friends", r.callbackFriends)
	b.Callback("/help_other_student", r.callbackOtherStudent)
	b.Callback("/help_settings", r.callbackSettings)
	b.Callback("/help_me", r.callbackMe)
	b.Callback("/help_support", r.callbackSupport)
	b.Callback("/help_faq", r.callbackFAQ)
	b.Callback("/help_buildings", r.callbackBuildings)
}

func (r *helpRouter) cmdHelp(c bot.Context) error {
	return c.SendMessageWithInlineKB(txtHelp, tgmodel.HelpButtons)
}

func (r *helpRouter) callbackHelpBack(c bot.Context) error {
	return c.EditMessageWithInlineKB(txtHelp, tgmodel.HelpButtons)
}

func (r *helpRouter) callbackSchedule(c bot.Context) error {
	return c.EditMessageWithInlineKB(txtHelpSchedule, kbHelpBack)
}

func (r *helpRouter) callbackGrades(c bot.Context) error {
	return c.EditMessageWithInlineKB(txtHelpGrades, kbHelpBack)
}

func (r *helpRouter) callbackFriends(c bot.Context) error {
	return c.EditMessageWithInlineKB(txtHelpFriends, kbHelpBack)
}

func (r *helpRouter) callbackOtherStudent(c bot.Context) error {
	return c.EditMessageWithInlineKB(txtHelpOtherStudent, kbHelpBack)
}

func (r *helpRouter) callbackSettings(c bot.Context) error {
	return c.EditMessageWithInlineKB(txtHelpSettings, kbHelpBack)
}

func (r *helpRouter) callbackMe(c bot.Context) error {
	return c.EditMessageWithInlineKB(txtHelpMe, kbHelpBack)
}

func (r *helpRouter) callbackSupport(c bot.Context) error {
	support, _ := os.LookupEnv("MAIN_DEVELOPER")
	return c.EditMessageWithInlineKB(fmt.Sprintf(txtHelpSupport, support), kbHelpBack)
}

func (r *helpRouter) callbackFAQ(c bot.Context) error {
	return c.EditMessageWithInlineKB(txtHelpFAQ, kbHelpBack)
}

func (r *helpRouter) callbackBuildings(c bot.Context) error {
	buildings, err := r.parser.FindBuildings()
	if err != nil {
		return err
	}
	txt := "Вот все адреса корпусов:\n"
	for _, b := range buildings {
		txt += fmt.Sprintf(formatBuilding, b.Name, b.SearchUrl, b.Address)
	}
	return c.EditMessageWithInlineKB(txt, kbHelpBack)
}

const (
	txtHelp = "<b>Помощь</b>.\nЗдесь находится основная информация о функционале бота.\n\n" +
		"<b>Внимание</b>! Этот бот создан студентом-энтузиастом и <b>не связан с разработчиками модеус</b>.\n" +
		"Указывая свой логин и пароль, <b>Вы действуете на свой страх и риск</b>!"
	txtHelpSchedule = "🗓 <b>Расписание</b>.\nБот может получать Ваше расписание из модеуса.\nДоступен просмотр расписания как один день, так и на всю неделю\n\n<b><i>Команды</i></b>:\n" +
		"- /day_schedule - расписание на один день\n" +
		"- /week_schedule - расписание на всю неделю."
	txtHelpGrades = "📊 <b>Оценки</b>.\nБот может получать Ваши оценки из модеуса, но для этого <i>требуется логин и пароль</i>.\n" +
		"Если Вы не указали их при запуске бота, то можно сделать это разделе настроек (/settings)\n\n<b><i>Доступные функции</i></b>:\n" +
		"- <b>Просмотр оценок</b> по каждому семестру.\n" +
		"- <b>Детальный просмотр</b> баллов и посещаемости по каждой встрече в рамках предмета"
	txtHelpFriends = "👨‍🎓👩‍🎓 <b>Друзья</b>.\nБот может добавлять студентов/преподавателей в друзья, чтобы смотреть их расписание быстро и удобно!\n\n" +
		"Все очень просто:\n" +
		"1) Нажимаете на кнопку <code>👨‍🎓👩‍🎓 Друзья</code>  (/friends), выбираете <b>\"Добавить друга\"</b>\n" +
		"2) Вводите ФИО друга\n\n" +
		"Теперь можно смотреть расписание друзей аналогично своему. Если расписание больше не интересно, друга можно удалить"
	txtHelpOtherStudent = "👥 <b>Другие студенты</b>.\nФункция для расписания студентов/преподавателей\n\nИнформация о них никак <b>не сохраняется</b> (в отличие от функционала друзей)\n" +
		"Удобно, если нужно посмотреть расписание случайного человека и <i>никак не взаимодействовать</i>"
	txtHelpSettings = "⚙️ <b>Настройки</b>.\n<b><i>Доступные функции</i></b>:\n" +
		"- <b>Добавить логин и пароль</b>. Открывает доступ к оценкам и рейтингам\n" +
		"- <b>Изменить ФИО</b>. Обновляем ФИО, если указали его с ошибкой"
	txtHelpMe = "\U0001FAF5 <b>Обо мне</b>.\n<b><i>Доступные функции</i></b>:\n" +
		"- <b>Обо мне</b>. Профиль подготовки, поток обучения\n" +
		"- <b>Рейтинги</b>. CGPA, а также GPA и посещаемость по семестрам. Требуется логин и пароль"
	txtHelpSupport = "🛡 <b>Поддержка</b>.\nПо всем вопросам/предложениям можно обращаться к создателю бота %s"
	txtHelpFAQ     = "❓ <b>FAQ</b>.\n" +
		"<blockquote>Обязательно ли указывать логин и пароль от учетной записи модеус?</blockquote>\n- Нет, это не обязательно, <b>весь функционал доступен</b>, кроме просмотра оценок.\n\n" +
		"<blockquote>Можно ли смотреть оценки другого студента?</blockquote>\n- Нет, эта <b>функция недоступна, даже если другой студент зарегистрирован в нашем боте</b> и у нас есть возможность получать его оценки.\n" +
		"Модеус не дает возможность смотреть студентам чужие и <i>мы согласны с этой позицией</i>\n\n" +
		"<blockquote>Это официальный телеграм бот модеуса?</blockquote>\n- Нет, это стороннее приложение, <b>никак не связанное с разработчиками модеус</b>.\nУказывая свой логин и пароль, <b>Вы действуете на свой страх и риск</b>!"
)
