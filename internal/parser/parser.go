package parser

import (
	"bot_for_modeus/pkg/modeus"
	"bot_for_modeus/pkg/redis"
	"context"
	"time"
)

const (
	parserServicePrefixLog = "/parser"
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

	DeleteToken(login string) error
}

type parser struct {
	modeus    modeus.Parser
	redis     redis.Redis
	rootLogin string
	rootPass  string
}

func NewParserService(modeus modeus.Parser, redis redis.Redis, rootLogin, rootPass string) Parser {
	return &parser{
		modeus:    modeus,
		redis:     redis,
		rootLogin: rootLogin,
		rootPass:  rootPass,
	}
}
