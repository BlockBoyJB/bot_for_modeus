package v2

const (
	stateInputFullName     = "stateInputFullName"
	stateChooseStudent     = "stateChooseStudent"
	stateActionAfterCreate = "stateActionAfterCreate"
	stateAddLoginPassword  = "stateAddLoginPassword"
	stateConfirmDelete     = "stateConfirmDelete"

	stateChooseFriend       = "stateChooseFriend"
	stateChooseFriendAction = "stateChooseFriendAction"
	stateAddFriend          = "stateAddFriend"
	stateChooseFindFriend   = "stateChooseFindFriend"

	stateInputOtherStudent        = "stateInputOtherStudent"
	stateChooseOtherStudent       = "stateChooseOtherStudent"
	stateChooseOtherStudentAction = "stateChooseOtherStudentAction"

	stateUserSchedule         = "stateUserSchedule"
	stateFriendSchedule       = "stateFriendSchedule"
	stateOtherStudentSchedule = "stateOtherStudentSchedule"

	stateChooseSubject  = "stateChooseSubject"
	stateChooseSemester = "stateChooseSemester"
)

const (
	txtDefault           = "Получить расписание на день: /day_schedule\nПолучить расписание на неделю: /week_schedule\nУзнать текущие баллы: /grades\nНастройки: /settings\nПомощь: /help\nОбо мне: /me"
	txtError             = "Ой! У нас произошла ошибка! Пожалуйста, воспользуйтесь сервисом позже или напишите в поддержку: /help -> Поддержка"
	txtModeusUnavailable = "Ой! Кажется, какие-то проблемы с модеусом!\nВы можете попробовать <b>посмотреть на сайте</b>, либо воспользоваться сервисом позже.\nЕсли ошибка уже давно, <b>обратитесь в поддержку</b> /help"
	txtStart             = "👋 Привет!\nЯ умею получать расписание и оценки из модеуса и отправлять Вам!\nДавайте найдем Вас в системе модеус! Напишите Ваше <b>ФИО без ошибок, как указано в модеусе</b>"
	txtStudentNotFound   = "Ой! Никого не могу найти с ФИО \"%s\".\nПожалуйста, введите ФИО точно как указано в модеусе (возможно ошибка с буквами е и ё)"
	txtUserCreated       = "Пользователь успешно создан!\n\n" +
		"Хотите добавить логин и пароль от своей учетной записи модеус?\n" +
		"Вы получите возможность <b>просматривать раздел оценок</b>.\nВы можете нажать кнопку \"Нет\", если не хотите указывать\n" +
		"Вам так же будет доступна возможность смотреть расписание, однако раздел оценок будет недоступен."
	txtAddLoginPassword   = "Пожалуйста, укажите через пробел сначала логин, потом пароль от учетной записи модеуса"
	txtRequiredLoginPass  = "Требуется логин и пароль от модеуса для входа в систему"
	txtIncorrectLoginPass = "Ой! Кажется, Вы ввели логин или пароль с ошибкой!\nПожалуйста измените его!\nЗайдите в настройки /settings -> \"добавить логин и пароль\" и введите новый логин и пароль без ошибок"
	txtSettings           = "Вы находитесь в настройках.\n\n" +
		"Существуют следующие функции:\n" +
		"\t- <b>Добавить логин и пароль</b>: открывает возможность получать текущие баллы по предметам и свои рейтинги в разделе /me -> Рейтинги\n" +
		"\t- <b>Изменить ФИО</b>: обновляем Ваше ФИО, если Вы указали его с ошибкой.\n"
	txtIncorrectLoginPassInput = "Ой! Кажется, Вы ввели логин и пароль с ошибкой! Пожалуйста, введите через пробел сначала логин, потом пароль"
	txtUserNotFound            = "Ой! Мы не можем найти информацию от Вас 👀!\nПожалуйста, перезапустите бота, нажав команду /start"

	txtConfirmDelete = "Вы уверены что хотите остановить бота? Вся информация о Вас будет <b>удалена из нашей системы</b>!\nЧтобы заново воспользоваться ботом, будет необходимо нажать /start и заново вводить все данные\nВы уверены, что <b>хотите остановить бота</b>?"
	txtUserDeleted   = "Бот остановлен! Информация о пользователе удалена!\nЧтобы снова воспользоваться ботом, нажмите /start"

	txtChooseFriend       = "Пожалуйста выберите друга, расписание которого хотите получить"
	txtChooseFriendAction = "Вы выбрали: <b>%s</b>\nВыберите действие с другом:"

	txtInputOtherStudent        = "Введите ФИО студента, расписание которого хотите узнать"
	txtChooseOtherStudentAction = "Вы выбрали: <b>%s</b>\nВыберите расписание, которое хотите получить:"

	txtMyProfile = "Вы находитесь в своем профиле. На данный момент можно узнать:\n" +
		"\t- <b>Обо мне</b>: узнать профиль подготовки и поток обучения\n" +
		"\t- <b>Рейтинги</b>: узнать CGPA, а также GPA и посещаемость по учебным семестрам"

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
	formatLessonGrades   = "Номер пары: %d\n<b>%s</b>\n%s\n⏰ %s\n📍 Отметка посещения: %s\n📊 Оценки: %s\n"
	formatSemester       = "%s\nGPA: %s\nПосещение: %s\nПропуск: %s\nНе отмечено: %s"
)
