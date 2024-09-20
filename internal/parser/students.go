package parser

import (
	"bot_for_modeus/pkg/modeus"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"time"
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
	for _, person := range response.Embedded.Persons {
		student := Student{
			FullName: person.FullName,
		}
		var ok bool

		for _, s := range response.Embedded.Students {
			if s.PersonId == person.Id {
				student = Student{
					FullName:         student.FullName,
					FlowCode:         s.FlowCode,
					SpecialtyName:    s.SpecialtyName,
					SpecialtyProfile: s.SpecialtyProfile,
					ScheduleId:       s.PersonId,
					GradesId:         s.Id,
				}
				if s.LearningEndDate.Unix() > 0 && s.LearningStartDate.Before(time.Now()) {
					student.SpecialtyProfile = "<i>Не является студентом на данный момент</i>" // Костыль, но сойдет
				}
				ok = true
				break
			}
		}
		if !ok {
			// Есть такой вариант, что пользователь ввел ФИО преподавателя,
			// а его данные находятся в другом списке в ответе
			for _, e := range response.Embedded.Employees {
				// Проверка на DateOut == "" нужна, потому что в ответе приходят разные занимаемые должности в рамках одного преподавателя.
				// Т.е есть разные периоды работы, в которых разный GroupName. Нам нужен тот, который не закончился
				if e.PersonId == person.Id && e.DateOut == "" {
					student = Student{
						FullName:         student.FullName,
						SpecialtyName:    "Преподаватель", // Вообще-то, не совсем понятно, что это преподаватель (в ответе массив называется "сотрудники").
						SpecialtyProfile: e.GroupName,
						ScheduleId:       e.PersonId,
						GradesId:         e.Id,
					}
					ok = true
					break
				}
			}
		}
		if ok {
			result = append(result, student)
		}
	}
	if len(result) == 0 {
		return nil, ErrStudentsNotFound
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
