package parser

import (
	"bytes"
	"errors"
	"fmt"
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
		if errors.Is(err, errRetriesExceeded) {
			return nil, ErrParserUnavailable
		}
		log.Err(err).Str("method", method).Str("uri", uri).Msg("parser/makeRequest error make http request")
		return nil, err
	}
	if resp.StatusCode > 400 {
		return nil, handleParserErr(resp)
	}
	return resp, nil
}

func handleParserErr(r *http.Response) error {
	switch r.StatusCode {
	case http.StatusForbidden:
		return ErrIncorrectLoginPassword

	case http.StatusInternalServerError:
		return ErrParserUnavailable

	case http.StatusServiceUnavailable:
		return ErrModeusUnavailable

	default:
		return fmt.Errorf("unexpected status code: %d", r.StatusCode)
	}
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

var errRetriesExceeded = errors.New("max retries exceeded")

type retry struct {
	retries int
	delay   time.Duration
}

func (rt *retry) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	for i := 0; i < rt.retries; i++ {
		resp, err = http.DefaultTransport.RoundTrip(r)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusInternalServerError {
			return
		}
		_ = resp.Body.Close()

		time.Sleep(rt.delay)
	}
	return nil, errRetriesExceeded
}
