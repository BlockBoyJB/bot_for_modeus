package parser

import (
	"bytes"
	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"time"
	"unsafe"
)

type Parser interface {
	FindStudents(fullName string) ([]Student, error)
	FindStudentById(scheduleId string) (Student, error)

	DaySchedule(scheduleId string, now time.Time) ([]Lesson, error)
	WeekSchedule(scheduleId string, now time.Time) (map[int][]Lesson, error)
	DayGrades(day time.Time, gi GradesInput) ([]DayGrades, error)

	SemesterTotalGrades(gi GradesInput, semester Semester) ([]SubjectGrades, error)
	Ratings(gi GradesInput) (string, []SemesterRatings, error)

	FindCurrentSemester(gi GradesInput) (Semester, error)
	FindAllSemesters(gi GradesInput) ([]Semester, error)

	FindSemesterSubjects(gi GradesInput, semester Semester) (map[string]string, error)
	SubjectDetailedInfo(gi GradesInput, semester Semester, subjectId string) ([]LessonGrades, error)

	FindBuildings() ([]Building, error)

	DeleteToken(login string) error
}

const (
	defaultRetryCount = 3
	defaultRetryDelay = time.Second
)

type parser struct {
	host   string
	client *http.Client
}

func NewParserService(host string) Parser {
	return &parser{
		host: host,
		client: &http.Client{
			Transport: &retry{
				next:    http.DefaultTransport,
				retries: defaultRetryCount,
				delay:   defaultRetryDelay,
			},
		},
	}
}

func (p *parser) makeRequest(method, uri string, v any) (*http.Response, error) {
	body, err := sonic.Marshal(v)
	if err != nil {
		log.Err(err).Msg("parser/makeRequest error marshal input body") // тут тело запроса логировать небезопасно
		return nil, err
	}

	r, err := http.NewRequest(method, p.host+uri, bytes.NewBuffer(body))
	if err != nil {
		log.Err(err).Str("method", method).Str("uri", uri).Msg("parser/makeRequest error init request")
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(r)
	if err != nil {
		// логируем все ошибки, даже типовые
		log.Err(err).Str("method", method).Str("uri", uri).Msg("parser/makeRequest error make http request")
		return nil, err
	}
	return resp, nil
}

func parseBody(r *http.Response, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Err(err).Msg("parser/parseBody error read response body")
		return err
	}
	_ = r.Body.Close()
	if err = sonic.UnmarshalString(b2s(body), v); err != nil {
		log.Err(err).Msg("parser/parseBody error unmarshal body to struct")
		return err
	}
	return nil
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

type retry struct {
	next    http.RoundTripper
	retries int
	delay   time.Duration
}

func (rt *retry) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	for i := 0; i < rt.retries; i++ {
		resp, err = rt.next.RoundTrip(r)
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode {
		case http.StatusOK, http.StatusBadRequest:
			return

		case http.StatusForbidden:
			return nil, ErrIncorrectLoginPassword

		case http.StatusServiceUnavailable:
			return nil, ErrModeusUnavailable

		}
		_ = r.Body.Close()

		select {
		case <-r.Context().Done():
			return nil, r.Context().Err()

		case <-time.After(rt.delay):
		}
	}
	return nil, ErrParserUnavailable
}
