package parser

import (
	"bytes"
	"fmt"
	"github.com/bytedance/sonic"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
	"unsafe"
)

const (
	parserServicePrefixLog = "/parser"
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
	FindAllSemesters(gi GradesInput) (map[string]Semester, error)

	FindSemesterSubjects(gi GradesInput, semester Semester) (map[string]string, error)
	SubjectDetailedInfo(gi GradesInput, semester Semester, subjectId string) ([]LessonGrades, error)

	FindBuildings() ([]Building, error)

	DeleteToken(login string) error
}

type parser struct {
	host   string
	client *http.Client
}

func NewParserService(host string) Parser {
	return &parser{
		host:   host,
		client: &http.Client{},
	}
}

func (p *parser) makeRequest(method, uri string, v any) (*http.Response, error) {
	body, err := sonic.Marshal(v)
	if err != nil {
		log.Errorf("%s/makeRequest parse body input error: %s", parserServicePrefixLog, err)
		return nil, err
	}

	r, err := http.NewRequest(method, p.host+uri, bytes.NewBuffer(body))
	if err != nil {
		log.Errorf("%s/makeRequest init request error: %s", parserServicePrefixLog, err)
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(r)
	if err != nil {
		log.Errorf("%s/makeRequest request error: %s", parserServicePrefixLog, err)
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
		log.Errorf("%s/parseBody error read response body: %s", parserServicePrefixLog, err)
		return err
	}
	_ = r.Body.Close()
	if err = sonic.UnmarshalString(b2s(body), v); err != nil {
		log.Errorf("%s/parseBody error unmarhal body to struct: %s", parserServicePrefixLog, err)
		return err
	}
	return nil
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
