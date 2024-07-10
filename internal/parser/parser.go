package parser

import (
	"bot_for_modeus/internal/repo"
	"bot_for_modeus/pkg/modeus"
	"bot_for_modeus/pkg/redis"
	"context"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	serviceParserPrefixLog = "/parser"
	defaultParserTimeout   = time.Minute
	tokenPrefix            = "token:"
	defaultTTL             = time.Hour*23 + time.Minute*50 // Можно конечно парсить jwt токен, но чет не охота
)

var lessonTypes = map[string]string{
	"LECT":        "Лекционное занятие",
	"LAB":         "Лабораторное занятие",
	"SEMI":        "Практическое занятие",
	"CUR_CHECK":   "Текущий контроль",
	"CONS":        "Консультация",
	"EVENT_OTHER": "Прочее",
	"SELF":        "Самостоятельная работа",
	"FINAL_CHECK": "Итоговая аттестация",
	"MID_CHECK":   "Аттестация",
}

type Parser interface {
	FindAllUsers(ctx context.Context, fullName string) ([]ModeusUser, error)
	FindUserById(ctx context.Context, scheduleId string) (ModeusUser, error)
	DaySchedule(ctx context.Context, scheduleId string) ([]Lesson, error)
	WeekSchedule(ctx context.Context, scheduleId string) (map[int][]Lesson, error)
	CourseTotalGrades(ctx context.Context, input GradesInput) ([]SubjectGrades, error)
	UserRatings(ctx context.Context, input GradesInput) (string, map[int]*AcademicPeriod, error)
}

type Service struct {
	repo      repo.User
	modeus    modeus.Parser
	redis     redis.Pool
	rootLogin string
	rootPass  string
}

func NewParserService(repo repo.User, modeus modeus.Parser, redis redis.Pool, rootLogin, rootPass string) *Service {
	return &Service{
		repo:      repo,
		modeus:    modeus,
		redis:     redis,
		rootLogin: rootLogin,
		rootPass:  rootPass,
	}
}

// TODO добавить автоматическое обновление root token

func (s *Service) rootToken(ctx context.Context) (string, error) {
	return s.parseToken(ctx, s.rootLogin, s.rootPass)
}

func (s *Service) userToken(ctx context.Context, login, password string) (string, error) {
	return s.parseToken(ctx, login, password)
}

func (s *Service) parseToken(ctx context.Context, login, password string) (string, error) {
	token, err := s.redis.Get(ctx, tokenPrefix+login).Result()
	if err == nil && token != "" {
		return token, nil
	}
	token, err = s.modeus.ExtractToken(login, password, defaultParserTimeout)
	if err != nil {
		log.Errorf("%s/parseToken error parse token from selenium client: %s", serviceParserPrefixLog, err)
		return "", err
	}
	if err = s.redis.Set(ctx, tokenPrefix+login, token, defaultTTL).Err(); err != nil {
		log.Errorf("%s/parseToken error save token into redis: %s", serviceParserPrefixLog, err)
		return "", err
	}
	return token, nil
}
