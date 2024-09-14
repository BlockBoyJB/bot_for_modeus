package v2

import (
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/pkg/bot"
	"fmt"
	"os"
)

var kbHelpBack = tgmodel.BackButton("/help_back")

type helpRouter struct {
	//parser parser.Parser
}

// TODO информация о сетке расписания, адреса корпусов
func newHelpRouter(b *bot.Bot) {
	r := &helpRouter{}

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

const (
	txtHelp = "<b>Помощь</b>.\nЗдесь находится основная информация о функционале бота.\nДля получения информации достаточно нажать на одну из кнопок ниже\n" +
		"<b>Внимание</b>! Этот телеграм бот - стороннее приложение, которое разрабатывалось студентом-энтузиастом ТюмГУ.\nРазработчики системы модеус не несут ответственности за сохранность Ваших данных.\n" +
		"Указывая свой логин и пароль от учетной записи модеус Вы действуете на <b>свой страх и риск</b>!"
	txtHelpSchedule = "🗓 <b>Расписание</b>.\nБот имеет возможность получать расписание из системы модеус.\n" +
		"Можно смотреть расписание на день и неделю.\nДоступна возможность перемотки (смотреть расписание на любой из дней) - достаточно взаимодействовать с инлайн кнопками под расписанием\n" +
		"/day_schedule - расписание на день\n" +
		"/week_schedule - расписание сразу на неделю"
	txtHelpGrades = "📊 <b>Оценки</b>.\nБот имеет возможность получать оценки из системы модеус.\nОднако для этого требуется Ваш логин и пароль от модеуса, чтобы мы могли их получать.\n" +
		"Если в начале работы с нашим ботом Вы решили не указывать логин и пароль, то можно зайти в настройки (/settings) -> \"добавить логин и пароль\".\n" +
		"Доступна возможность смотреть оценки по семестрам (оценки за каждый семестр отдельно), а также детальный просмотр оценок и посещаемости по каждой встрече в рамках одного предмета.\n\n" +
		"<b>Внимание</b>!\nЭтот телеграм бот - стороннее приложение, которое разрабатывалось студентом-энтузиастом ТюмГУ.\nРазработчики системы модеус не несут ответственности за сохранность Ваших данных.\n" +
		"Указывая свой логин и пароль от учетной записи модеус Вы действуете на <b>свой страх и риск</b>!"
	txtHelpFriends = "👨‍🎓👩‍🎓 <b>Друзья</b>.\nБот имеет возможность получать расписание других студентов. Поэтому мы решили, что будет удобным добавить возможность сохранять друзей, чтобы быстро получать их расписание\n" +
		"Все очень просто: вводишь команду /friends, нажимаешь кнопку \"Добавить друга\", вводишь его ФИО и готово!\nТеперь достаточно нажать команду /friends, выбрать друга и получить его расписание на день/неделю (аналогично своему)\n" +
		"Если больше не хотите смотреть расписание друга, то есть возможность его удалить"
	txtHelpOtherStudent = "👥 <b>Другие студенты</b>.\nФункция похожа на функцию друзей, но в этом случае информация о студенте никак не сохраняется\n" +
		"Точно так же есть опция получения расписания на день/неделю\n" +
		"Удобно, если нужно посмотреть расписание случайного человека и никак больше с ним не взаимодействовать"
	txtHelpSettings = "⚙️ <b>Настройки</b>.\nДоступны следующие функции:\n" +
		"<b>Добавить логин и пароль</b>. Открывает возможность получать оценки и свои рейтинги в разделе /me -> Рейтинги\n" +
		"<b>Внимание</b>!\nЭтот телеграм бот - стороннее приложение, которое разрабатывалось студентом-энтузиастом ТюмГУ.\nРазработчики системы модеус не несут ответственности за сохранность Ваших данных.\n" +
		"Указывая свой логин и пароль от учетной записи модеус Вы действуете на <b>свой страх и риск</b>!\n\n" +
		"<b>Изменить ФИО</b>. Обновляем информацию о Вас, если вы указали не тот ФИО"
	txtHelpMe = "\U0001FAF5 <b>Обо мне</b>.\nЗдесь Вы можете узнать следующую информацию:\n" +
		"<b>Обо мне</b>. Профиль подготовки, поток обучения\n" +
		"<b>Рейтинги</b>. CGPA, а также по каждому из семестров получить GPA и посещаемость. Требует наличия логина и пароля"
	txtHelpSupport = "🛡 <b>Поддержка</b>.\nПо всем вопросам можно обращаться к создателю бота %s"
	txtHelpFAQ     = "❓ <b>FAQ</b>.\n<i>Обязательно ли указывать логин и пароль от учетной записи модеус?</i>\n\t- Нет, не обязательно, весь функционал бота будет доступен, но без него Вам не будет доступна опция просмотра оценок.\n\n" +
		"<i>Можно ли смотреть оценки другого студента?</i>\n\t- Нет, эта функция недоступна, даже если другой студент зарегистрирован в нашем боте и у нас есть возможность получать его оценки.\n" +
		"Модеус не дает возможность смотреть студентам чужие и мы согласны с этой позицией\n\n" +
		"<i>Это официальный телеграм бот модеуса?</i>\n\t- Нет, это стороннее приложение, сделанное студентом-энтузиастом.\nРазработчики системы модеус не несут ответственности за сохранность Ваших данных.\nУказывая свой логин и пароль от учетной записи модеус Вы действуете на <b>свой страх и риск</b>!"
)
