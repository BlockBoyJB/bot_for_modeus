package modeus

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

const (
	defaultModeusUrl     = "https://utmn.modeus.org"
	defaultModeusTimeout = time.Second * 30
)

type Parser interface {
	modeusClient
	tokenClient
}

type modeusClient interface {
	FindStudents(token, fullName string) (StudentResponse, error)
	FindStudentById(token, id string) (StudentResponse, error)

	Schedule(token string, input ScheduleRequest) (ScheduleResponse, error)
	EventGrades(token, eventId string, input EventGradesRequest) (EventGradesResponse, error)

	FindAPR(token, gradesId string) ([]APRealization, error)
	FindCurrentAPR(token, gradesId string) (APRealization, error)
	StudentRatings(token string, input RatingsRequest) (RatingsResponse, error)
	CourseUnits(token string, input PrimaryGradesRequest) ([]CourseUnit, error)
	CoursesTotalResults(token string, input SecondaryGradesRequest) (SecondaryGradesResponse, error)
}

// Вынес работу с токенами в отдельный сервис. Теперь взаимодействие через http (может grpc сделаю).
type tokenClient interface {
	// GetToken получает токен из модеуса, сохраняет в кэш и ежедневно обновляет
	GetToken(login, password string) (string, error)
	// DeleteToken Удаляет токен из кэша, а также останавливает ежедневное обновление токена.
	// Если этого не сделать, то будут обновляться токены не существующих пользователей (если пользователь нажал /stop)
	DeleteToken(login string) error
}

type modeus struct {
	*TokenService
	client *http.Client
}

func NewModeus(ts *TokenService) Parser {
	return &modeus{
		TokenService: ts,
		client: &http.Client{
			Timeout: defaultModeusTimeout,
		},
	}
}

func (s *modeus) makeRequest(token, method, uri string, v any) (*http.Response, error) {
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

	resp, err := s.client.Do(r)
	if err != nil || resp.StatusCode > 300 {
		return nil, ErrModeusUnavailable
	}
	return resp, nil
}
