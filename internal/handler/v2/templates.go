package v2

const (
	stateInputFullName     = "stateInputFullName"
	stateChooseStudent     = "stateChooseStudent"
	stateActionAfterCreate = "stateActionAfterCreate"
	stateAddLoginPassword  = "stateAddLoginPassword"
	stateConfirmDelete     = "stateConfirmDelete"

	stateAddFriend        = "stateAddFriend"
	stateChooseFindFriend = "stateChooseFindFriend"

	stateInputOtherStudent  = "stateInputOtherStudent"
	stateChooseOtherStudent = "stateChooseOtherStudent"
)

const (
	txtError             = "Ой! У нас произошла ошибка! Пожалуйста, воспользуйтесь сервисом позже или напишите в поддержку: /help -> Поддержка"
	txtModeusUnavailable = "Ой! Кажется, какие-то проблемы с модеусом!\nВы можете <b>посмотреть на сайте</b>, либо воспользоваться сервисом позже.\nЕсли ошибка уже давно, <b>обратитесь в поддержку</b> /help"

	txtStart              = "👋 Привет!\nЯ умею получать расписание и оценки из модеуса!\nНапишите Ваше <b>ФИО без ошибок, как указано в модеусе</b>, чтобы мы смогли найти Вас!"
	txtStudentNotFound    = "Ой! Никого не могу найти с ФИО \"%s\".\nПожалуйста, <b>введите ФИО точно как указано в модеусе</b> (возможно ошибка с буквами е и ё)"
	txtUserCreated        = "<b><i>Пользователь успешно создан</i></b>!\n\n<b>Добавить логин и пароль от модеуса</b>? Это откроет доступ к <b>разделу оценок</b>\nЕсли нет - <i>сможете смотреть только расписание</i>"
	txtAddLoginPassword   = "Пожалуйста, укажите через пробел сначала логин, потом пароль от учетной записи модеуса"
	txtRequiredLoginPass  = "<b>Требуется логин и пароль</b> от модеуса для входа в систему\n\n/settings -> \"Добавить логин и пароль\""
	txtIncorrectLoginPass = "Ой! Кажется, <b>Вы ввели логин или пароль с ошибкой</b>!\nПожалуйста измените его в настройках! (/settings)"
	txtUserNotFound       = "Ой! Мы не можем найти информацию от Вас 👀!\nПожалуйста, <b>перезапустите бота</b>, нажав команду /start"

	txtSettings = "⚙️ <b>Настройки</b>.\n\n" +
		"- <b>Добавить логин и пароль</b>: открывает доступ к оценкам и рейтингам\n\n" +
		"- <b>Изменить ФИО</b>: обновляем ФИО, если указали его с ошибкой"
	txtIncorrectLoginPassInput = "Ой! Кажется, Вы ввели логин и пароль с ошибкой! Пожалуйста, введите через пробел сначала логин, потом пароль"

	txtConfirmDelete = "<b><i>Вы уверены, что хотите остановить бота</i></b>?\nВся информация <b>будет удалена</b>.\nДля повторного использования нужно будет нажать /start и ввести данные заново."
	txtUserDeleted   = "Бот остановлен, данные удалены.\nДля повторного использования нажмите /start."

	txtFriends            = "👨‍🎓👩‍🎓 <b>Друзья</b>.\n\nВыберите друга, расписание которого хотите получить"
	txtChooseFriendAction = formatFullName + "Выберите действие с другом:"

	txtInputOtherStudent        = "Введите ФИО студента, расписание которого хотите узнать"
	txtChooseOtherStudentAction = "Вы выбрали: <b>%s</b>\nВыберите расписание, которое хотите получить:"

	txtMyProfile = "Вы находитесь в <i>своем профиле</i>.\n\n" +
		"- <b>Обо мне</b>: профиль подготовки и поток обучения\n\n" +
		"- <b>Рейтинги</b>: CGPA, а также GPA и посещаемость по семестрам"

	// для мамкиных хацкеров =)
	txtWarn = "Прекращай баловаться!"
)

// Шаблоны для форматирования данных в текст
const (
	formatStudent        = "👤 <b>%s</b>\n%s | %s\n%s\n"
	formatFullName       = "👤 <b>%s</b>\n"
	formatLesson         = "⏰ <b>%s</b>\n📚 <b>%s</b> | %s\n%s\n🏫 %s, %s\n👨‍🏫 %s"
	formatSemesterGrades = "%s <b>%s</b>\nТекущий результат: %s\nИтог модуля: %s\nПосещение: %s\nПропуск: %s\nНе отмечено: %s"
	formatDayGrades      = "⏰ <b>%s</b>\n📕 <b>%s</b> | %s\n%s\n📍 Отметка посещения: %s\n📊 Оценки: %s\n"
	formatLessonGrades   = "<b>%s</b>\n%s\n⏰ %s\n📍 Отметка посещения: %s\n📊 Оценки: %s\n"
	formatSemester       = "%s\nGPA: %s\nПосещение: %s\nПропуск: %s\nНе отмечено: %s"
	formatBuilding       = "%s: <a href=\"%s\">%s</a>\n"
)
