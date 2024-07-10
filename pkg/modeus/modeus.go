package modeus

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

const defaultModeusUrl = "https://utmn.modeus.org"

type Parser interface {
	modeusClient
	seleniumClient
}

type modeusClient interface {
	FindStudents(token, fullName string) (StudentResponse, error)
	FindStudentById(token, id string) (StudentResponse, error)

	Schedule(token string, input ScheduleRequest) (ScheduleResponse, error)

	FindAPR(token, gradesId string) ([]APRealization, error)
	FindCurrentAPR(token, gradesId string) (APRealization, error)
	StudentRatings(token string, input RatingsRequest) (RatingsResponse, error)
	CourseUnits(token string, input PrimaryGradesRequest) ([]CourseUnit, error)
	CoursesTotalResults(token string, input SecondaryGradesRequest) (SecondaryGradesResponse, error)
}

type seleniumClient interface {
	ExtractToken(login, password string, timeout time.Duration) (string, error)
}

type Modeus struct {
	*Selenium
	client *http.Client
}

func NewModeus(selenium *Selenium) *Modeus {
	return &Modeus{
		Selenium: selenium,
		client:   http.DefaultClient,
	}
}

func (s *Modeus) makeRequest(token, method, uri string, v any) (*http.Response, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequest(method, defaultModeusUrl+uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")

	return s.client.Do(r)
}
