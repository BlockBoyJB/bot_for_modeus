package modeus

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const (
	findStudentsUri = "/schedule-calendar-v2/api/people/persons/search"
)

type StudentResponse struct {
	Embedded struct {
		Persons []struct {
			LastName   string `json:"lastName"`
			FirstName  string `json:"firstName"`
			MiddleName string `json:"middleName"`
			FullName   string `json:"fullName"`
			Links      struct {
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"_links"`
			Id string `json:"id"` // Совпадает с Students PersonId (только для поиска расписания)
		} `json:"persons"`
		Students []struct {
			Id                string      `json:"id"`                // Id для поиска оценок (тут должно быть совпадение с id внутри jwt токена)
			PersonId          string      `json:"personId"`          // Id для поиска расписания пользователя
			FlowId            string      `json:"flowId"`            // Id потока
			FlowCode          string      `json:"flowCode"`          // Название потока (например 2023, Бакалавриат, Специалитет, Очная форма)
			SpecialtyCode     string      `json:"specialtyCode"`     // Код специальности
			SpecialtyName     string      `json:"specialtyName"`     // Название специальности
			SpecialtyProfile  string      `json:"specialtyProfile"`  // Профиль специальности
			LearningStartDate time.Time   `json:"learningStartDate"` // Начало обучения
			LearningEndDate   interface{} `json:"learningEndDate"`   // Конец обучения, может приходить null
		} `json:"students"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
}

func (s *Modeus) FindStudents(token, fullName string) (StudentResponse, error) {
	type request struct {
		FullName string `json:"fullName"`
		Sort     string `json:"sort"`
		Size     int    `json:"size"`
		Page     int    `json:"page"`
	}
	req := request{
		FullName: fullName,
		Sort:     "+fullName",
		Size:     10,
		Page:     0,
	}
	return s.findStudentRequest(token, req)
}

func (s *Modeus) FindStudentById(token, id string) (StudentResponse, error) {
	type request struct {
		Id   []string `json:"id"`
		Sort string   `json:"sort"`
		Size int      `json:"size"`
	}
	req := request{
		Id:   []string{id},
		Sort: "+fullName",
		Size: 2147483647, // TODO size???
	}
	return s.findStudentRequest(token, req)
}

func (s *Modeus) findStudentRequest(token string, req interface{}) (StudentResponse, error) {
	resp, err := s.makeRequest(token, http.MethodPost, findStudentsUri, req)
	if err != nil {
		return StudentResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return StudentResponse{}, err
	}
	var response StudentResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return StudentResponse{}, err
	}
	if response.Page.TotalElements == 0 {
		return StudentResponse{}, ErrStudentsNotFound
	}
	return response, nil
}
