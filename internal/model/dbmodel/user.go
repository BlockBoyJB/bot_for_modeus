package dbmodel

type User struct {
	UserId     int64             `bson:"user_id"`
	FullName   string            `bson:"full_name"`
	Login      string            `bson:"login"`
	Password   string            `bson:"password"`
	ScheduleId string            `bson:"schedule_id"` // Id пользователя для поиска расписания
	GradesId   string            `bson:"grades_id"`   // Id пользователя для поиска оценок
	Friends    map[string]string `bson:"friends"`     // Ключ - PersonId, значение - ФИО кореша
}
