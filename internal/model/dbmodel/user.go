package dbmodel

type User struct {
	UserId     int64    `bson:"user_id"`
	FullName   string   `bson:"full_name"`
	Login      string   `bson:"login"`
	Password   string   `bson:"password"`
	ScheduleId string   `bson:"schedule_id"` // Id пользователя для поиска расписания
	GradesId   string   `bson:"grades_id"`   // Id пользователя для поиска оценок
	Friends    []Friend `bson:"friends"`     // Слайс, а не мапа, чтобы гарантировать порядок
}

type Friend struct {
	FullName   string `bson:"full_name"`
	ScheduleId string `bson:"schedule_id"`
}
