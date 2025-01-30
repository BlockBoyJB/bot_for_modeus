package service

import (
	"bot_for_modeus/internal/parser"
	"bot_for_modeus/internal/repo"
	"bot_for_modeus/pkg/crypter"
	"context"
)

type (
	UserInput struct {
		UserId     int64
		FullName   string
		ScheduleId string
		GradesId   string
	}
	UserOutput struct {
		FullName   string
		Login      string
		Password   string
		ScheduleId string
		GradesId   string
		Friends    []FriendOutput
	}
	UserLoginPasswordInput struct {
		UserId   int64
		Login    string
		Password string
	}
	FriendInput struct {
		UserId     int64
		FullName   string
		ScheduleId string
	}
	FriendOutput struct {
		FullName   string `json:"full_name"`
		ScheduleId string `json:"schedule_id"`
	}
)

type User interface {
	Create(ctx context.Context, input UserInput) error
	Find(ctx context.Context, userId int64) (UserOutput, error)
	UpdateLoginPassword(ctx context.Context, input UserLoginPasswordInput) error
	UpdateInfo(ctx context.Context, input UserInput) error
	Delete(ctx context.Context, userId int64) error
	AddFriend(ctx context.Context, input FriendInput) error
	DeleteFriend(ctx context.Context, input FriendInput) error

	Decrypt(input string) (string, error)
}

type (
	Services struct {
		User   User
		Parser parser.Parser
	}
	ServicesDependencies struct {
		Repos      *repo.Repositories
		Crypter    crypter.Crypter
		ParserHost string
	}
)

func NewServices(d *ServicesDependencies) *Services {
	return &Services{
		User:   newUserService(d.Repos.User, d.Crypter),
		Parser: parser.NewParserService(d.ParserHost),
	}
}
