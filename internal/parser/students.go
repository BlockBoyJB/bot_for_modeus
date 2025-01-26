package parser

import (
	"fmt"
	"net/http"
)

const (
	findStudentsUri = "/students"
)

type Student struct {
	FullName         string `json:"full_name"`         // ФИО
	FlowCode         string `json:"flow_code"`         // Код направления
	SpecialtyName    string `json:"specialty_name"`    // Название направления
	SpecialtyProfile string `json:"specialty_profile"` // Специальность
	ScheduleId       string `json:"schedule_id"`       // Id для поиска расписания
	GradesId         string `json:"grades_id"`         // Id для поиска оценок
}

func (p *parser) FindStudents(fullName string) ([]Student, error) {
	uri := fmt.Sprintf("%s?full_name=%s", findStudentsUri, fullName)

	resp, err := p.makeRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusBadRequest {
		return nil, ErrStudentsNotFound
	}

	var result []Student
	if err = parseBody(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *parser) FindStudentById(scheduleId string) (Student, error) {
	uri := fmt.Sprintf("%s?schedule_id=%s", findStudentsUri, scheduleId)

	resp, err := p.makeRequest(http.MethodGet, uri, nil)
	if err != nil {
		return Student{}, err
	}

	var result Student
	if err = parseBody(resp, &result); err != nil {
		return Student{}, err
	}
	return result, nil
}
