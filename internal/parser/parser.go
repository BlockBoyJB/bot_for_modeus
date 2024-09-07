package parser

import (
	"bot_for_modeus/internal/repo"
	"bot_for_modeus/pkg/modeus"
	"bot_for_modeus/pkg/redis"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	parserServicePrefixLog = "/parser"
	defaultParserTimeout   = time.Minute
	tokenPrefix            = "token:"
	defaultTTL             = time.Hour*23 + time.Minute*50
)

var lessonTypes = map[string]string{
	"LECT":        "✍️ Лекционное занятие",
	"LAB":         "🔬 Лабораторное занятие",
	"SEMI":        "🧪 Практическое занятие",
	"CUR_CHECK":   "🔫 Текущий контроль",
	"CONS":        "🔮 Консультация",
	"EVENT_OTHER": "📌 Прочее",
	"SELF":        "🕯 Самостоятельная работа",
	"FINAL_CHECK": "🪓 Итоговая аттестация",
	"MID_CHECK":   "🔪 Аттестация",
}

type Parser interface {
	FindStudents(ctx context.Context, fullName string) ([]Student, error)
	FindStudentById(ctx context.Context, scheduleId string) (Student, error)

	DaySchedule(ctx context.Context, scheduleId string, now time.Time) ([]Lesson, error)
	WeekSchedule(ctx context.Context, scheduleId string, now time.Time) (map[int][]Lesson, error)
	DayGrades(ctx context.Context, day time.Time, input GradesInput) ([]SubjectDayGrades, error)

	SemesterTotalGrades(ctx context.Context, input GradesInput, semester Semester) ([]SubjectGrades, error)
	Ratings(ctx context.Context, input GradesInput) (string, map[int]SemesterResult, error)

	FindCurrentSemester(ctx context.Context, input GradesInput) (Semester, error)
	FindAllSemesters(ctx context.Context, input GradesInput) (map[string]Semester, error)

	FindSemesterSubjects(ctx context.Context, input GradesInput, semester Semester) (map[string]string, error)
	SubjectDetailedInfo(ctx context.Context, input GradesInput, semester Semester, subjectId string) (map[int]LessonGrades, error)
}

type parser struct {
	user      repo.User
	modeus    modeus.Parser
	redis     redis.Redis
	rootLogin string
	rootPass  string
}

func NewParserService(user repo.User, modeus modeus.Parser, redis redis.Redis, rootLogin, rootPass string) Parser {
	return &parser{
		user:      user,
		modeus:    modeus,
		redis:     redis,
		rootLogin: rootLogin,
		rootPass:  rootPass,
	}
}

// Такой немного костыльный метод, чтобы обновлять токен до его истечения
// Если этого не сделать, то в промежуток до его обновления селениум просто умрет от количества запросов
func (p *parser) loadToken(login, password string) {
	for {
		time.Sleep(defaultTTL - time.Minute)
		token, err := p.modeus.ExtractToken(login, password, defaultParserTimeout)
		if err != nil {
			log.Errorf("%s/loadToken error parse token from selenium client: %p", parserServicePrefixLog, err)
			return // завершаем работу, потому что parseToken в случае ошибки запустит еще одну такую же
		}
		if err = p.redis.Set(context.Background(), tokenPrefix+login, token, defaultTTL).Err(); err != nil {
			log.Errorf("%s/loadToken error save token into redis: %p", parserServicePrefixLog, err)
			return // завершаем работу, потому что parseToken в случае ошибки запустит еще одну такую же
		}
	}
}

func (p *parser) rootToken(ctx context.Context) (string, error) {
	return p.parseToken(ctx, p.rootLogin, p.rootPass)
}

func (p *parser) userToken(ctx context.Context, login, password string) (string, error) {
	return p.parseToken(ctx, login, password)
}

func (p *parser) parseToken(ctx context.Context, login, password string) (string, error) {
	token, err := p.redis.Get(ctx, tokenPrefix+login).Result()
	if err == nil && token != "" {
		return token, nil
	}
	token, err = p.modeus.ExtractToken(login, password, defaultParserTimeout)
	if err != nil {
		if errors.Is(err, modeus.ErrIncorrectInputData) {
			return "", ErrIncorrectLoginPassword
		}
		log.Errorf("%s/parseToken error parse token from selenium client: %s", parserServicePrefixLog, err)
		return "", err
	}
	if err = p.redis.Set(ctx, tokenPrefix+login, token, defaultTTL).Err(); err != nil {
		log.Errorf("%s/parseToken error save token into redis: %s", parserServicePrefixLog, err)
		return "", err
	}
	go p.loadToken(login, password)
	return token, nil
}
