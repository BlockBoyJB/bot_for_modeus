package parser

import (
	"bot_for_modeus/pkg/modeus"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
)

type Student struct {
	FullName         string // ФИО
	FlowCode         string // Код направления
	SpecialtyName    string // Название направления
	SpecialtyProfile string // Специальность
	ScheduleId       string // Id для поиска расписания
	GradesId         string // Id для поиска оценок
}

func (p *parser) FindStudents(ctx context.Context, fullName string) ([]Student, error) {
	token, err := p.rootToken(ctx)
	if err != nil {
		return nil, err
	}

	response, err := p.modeus.FindStudents(token, fullName)
	if err != nil {
		if errors.Is(err, modeus.ErrModeusUnavailable) {
			return nil, ErrModeusUnavailable
		}
		if errors.Is(err, modeus.ErrStudentsNotFound) {
			return nil, ErrStudentsNotFound
		}
		log.Errorf("%s/FindStudents error find students from modeus: %s", parserServicePrefixLog, err)
		return nil, err
	}

	var result []Student
	for _, s := range response.Embedded.Students {
		student := Student{
			FullName:         "Ошибка", // Ставим ошибку по дефолту, потому что ФИО содержится в response.Embedded.Persons
			FlowCode:         s.FlowCode,
			SpecialtyName:    s.SpecialtyName,
			SpecialtyProfile: s.SpecialtyProfile,
			ScheduleId:       s.PersonId,
			GradesId:         s.Id,
		}
		// Идем доп циклом потому что порядок в persons и students иногда не совпадает.
		// Да, получается O(n^2), но у нас лимит на поиск пользователей маленький (максимум 10)
		for _, person := range response.Embedded.Persons {
			if person.Id == s.PersonId {
				student.FullName = person.FullName
				break
			}
		}
		result = append(result, student)
	}
	return result, nil
}

func (p *parser) FindStudentById(ctx context.Context, scheduleId string) (Student, error) {
	token, err := p.rootToken(ctx)
	if err != nil {
		return Student{}, err
	}
	student, err := p.modeus.FindStudentById(token, scheduleId)
	if err != nil {
		if errors.Is(err, modeus.ErrModeusUnavailable) {
			return Student{}, ErrModeusUnavailable
		}
		log.Errorf("%s/FindStudentById error find student: %s", parserServicePrefixLog, err)
		return Student{}, err
	}
	if student.Page.TotalElements != 1 {
		return Student{}, errors.New("students with specified id more than 1")
	}
	s := student.Embedded.Students[0]
	return Student{
		FullName:         student.Embedded.Persons[0].FullName,
		FlowCode:         s.FlowCode,
		SpecialtyName:    s.SpecialtyName,
		SpecialtyProfile: s.SpecialtyProfile,
	}, nil
}
