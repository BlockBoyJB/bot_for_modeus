package service

import (
	"bot_for_modeus/internal/model/dbmodel"
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/repo"
	"bot_for_modeus/pkg/crypter"
	"bot_for_modeus/pkg/modeus"
	"bot_for_modeus/pkg/redis"
	"context"
)

type (
	UserInput struct {
		UserId     int64
		FullName   string
		ScheduleId string
		GradesId   string
	}
	User interface {
		CreateUser(ctx context.Context, input UserInput) error
		FindUser(ctx context.Context, userId int64) (dbmodel.User, error)
		UpdateLoginPassword(ctx context.Context, userId int64, login, password string) error
		UpdateUserInformation(ctx context.Context, input UserInput) error
		AddFriend(ctx context.Context, userId int64, fullName, scheduleId string) error
		DeleteFriend(ctx context.Context, userId int64, scheduleId string) error
		DeleteUser(ctx context.Context, userId int64) error
	}
)

type (
	Services struct {
		User   User
		Parser parser.Parser
	}
	ServicesDependencies struct {
		Repos     *repo.Repositories
		Parser    modeus.Parser
		Redis     redis.Pool
		Crypter   crypter.PasswordCrypter
		RootLogin string
		RootPass  string
	}
)

func NewServices(d ServicesDependencies) *Services {
	return &Services{
		User:   newUserService(d.Repos, d.Crypter),
		Parser: parser.NewParserService(d.Repos.User, d.Parser, d.Redis, d.RootLogin, d.RootPass),
	}
}
