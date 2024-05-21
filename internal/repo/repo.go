package repo

import (
	"bot_for_modeus/internal/model/pgmodel"
	"bot_for_modeus/internal/repo/pgdb"
	"bot_for_modeus/pkg/postgres"
	"context"
)

type User interface {
	CreateUser(ctx context.Context, u pgmodel.User) error
	GetUserById(ctx context.Context, userId int64) (pgmodel.User, error)
	AddLoginPassword(ctx context.Context, userId int64, login, password string) error
	DeleteLoginPassword(ctx context.Context, userId int64) error
	UpdateFullName(ctx context.Context, userId int64, fullName string) error
	DeleteUser(ctx context.Context, userId int64) error
}

type Repositories struct {
	User
}

func NewRepositories(pg *postgres.Postgres) *Repositories {
	return &Repositories{
		User: pgdb.NewUserRepo(pg),
	}
}
