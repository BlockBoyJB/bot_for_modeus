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
		var e *modeus.ErrModeusUnavailable
		if errors.As(err, &e) {
			log.Errorf("%s/FindStudents modeus error: %s", parserServicePrefixLog, e)
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
		// Минимальная информация о студенте/преподавателе, необходимая для работы
		student := Student{
			FullName:   person.FullName,
			ScheduleId: person.Id,
		}
		fillStudent(&response, &student)
		result = append(result, student)
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
	response, err := p.modeus.FindStudentById(token, scheduleId)
	if err != nil {
		var e *modeus.ErrModeusUnavailable
		if errors.As(err, &e) {
			log.Errorf("%s/FindStudentById modeus error: %s", parserServicePrefixLog, e)
			return Student{}, ErrModeusUnavailable
		}
		log.Errorf("%s/FindStudentById error find student: %s", parserServicePrefixLog, err)
		return Student{}, err
	}
	if response.Page.TotalElements != 1 {
		return Student{}, errors.New("students with specified id more than 1")
	}

	// Минимальная информация о студенте/преподавателе, необходимая для работы
	student := Student{
		FullName:   response.Embedded.Persons[0].FullName,
		ScheduleId: scheduleId,
	}

	fillStudent(&response, &student)
	return student, nil
}

// Функция, которая заполняет информацию о студенте из StudentResponse
// Предполагается, что ФИО и scheduleId (Person Id) уже имеются
func fillStudent(response *modeus.StudentResponse, student *Student) {
	for _, s := range response.Embedded.Students {
		if s.PersonId == student.ScheduleId {
			student.FlowCode = s.FlowCode
			student.SpecialtyName = s.SpecialtyName
			student.SpecialtyProfile = s.SpecialtyProfile
			student.ScheduleId = s.PersonId
			student.GradesId = s.Id

			if s.LearningEndDate.Unix() > 0 && s.LearningStartDate.Before(time.Now()) {
				student.SpecialtyProfile = "<i>Не является студентом на данный момент</i>" // Костыль, но сойдет
			}
			return
		}
	}
	for _, e := range response.Embedded.Employees {
		if e.PersonId == student.ScheduleId && e.DateOut == "" {
			// Проверка на DateOut == "" нужна, потому что в ответе приходят разные занимаемые должности в рамках одного преподавателя.
			// Т.е есть разные периоды работы, в которых разный GroupName. Нам нужен тот, который не закончился
			student.SpecialtyName = "Преподаватель" // Вообще-то, не совсем понятно, что это преподаватель (в ответе массив называется "сотрудники").
			student.SpecialtyProfile = e.GroupName
			student.ScheduleId = e.PersonId
			student.GradesId = e.Id
			return
		}
	}
}
