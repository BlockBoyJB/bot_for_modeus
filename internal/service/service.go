package service

import (
	"bot_for_modeus/internal/broker"
	"bot_for_modeus/internal/repo"
	"bot_for_modeus/pkg/parser"
	"context"
)

type User interface {
	CreateUser(ctx context.Context, userId int64, fullName string) error
	UpdateUserFullName(ctx context.Context, userId int64, fullName string) error
	AddUserLoginPassword(ctx context.Context, userId int64, login, password string) error
	DeleteLoginPassword(ctx context.Context, userId int64) error
	DeleteUser(ctx context.Context, userId int64) error
}

type Parser interface {
	DaySchedule(ctx context.Context, userId int64) (string, error)
	WeekSchedule(ctx context.Context, userId int64) (string, error)
	UserGrades(ctx context.Context, userId int64) (map[int]int, string, error)
	SubjectGradesInfo(ctx context.Context, userId int64, index int) (string, error)
	OtherStudentDaySchedule(fullName string) (string, error)
	OtherStudentWeekSchedule(fullName string) (string, error)
}

// Broker реализует работу с парсером через очередь. В текущем варианте отказался ввиду необоснованности усложнения
type Broker interface {
	RequestDaySchedule(ctx context.Context, userId int64, messageId int) error
	RequestWeekSchedule(ctx context.Context, userId int64, messageId int) error
	RequestUserGrades(ctx context.Context, userId int64, messageId int) error
	RequestSubjectInfo(ctx context.Context, userId int64, messageId int) error
}

type (
	Services struct {
		User
		Parser
		Broker
	}
	ServicesDependencies struct {
		Repos     *repo.Repositories
		Parser    parser.Parser
		Rabbit    *broker.Broker
		RootLogin string
		RootPass  string
	}
)

func NewServices(d ServicesDependencies) *Services {
	return &Services{
		User:   newUserService(d.Repos),
		Parser: newParserService(d.Repos, d.Parser, d.RootLogin, d.RootPass),
		Broker: newBrokerService(d.Rabbit),
	}
}
