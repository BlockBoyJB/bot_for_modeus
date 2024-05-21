package pgmodel

type User struct {
	Id       int    `db:"id"`
	UserId   int64  `db:"user_id"`
	FullName string `db:"full_name"`
	Login    string `db:"login"`
	Password string `db:"password"`
}
