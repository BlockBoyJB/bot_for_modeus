package parser

import (
	"bot_for_modeus/pkg/modeus"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
)

type ModeusUser struct {
	FullName         string
	FlowCode         string
	SpecialtyName    string
	SpecialtyProfile string
	ScheduleId       string // save to db
	GradesId         string // save to db
}

// FindAllUsers чаще всего находит только одного пользователя, однако могут быть приколы у Ивановых Иванов (но на 06.24 таких повторений нет)
// Несколько пользователей может быть только в случае когда ФИО введено неточно (например только фамилия и имя)
func (s *Service) FindAllUsers(ctx context.Context, fullName string) ([]ModeusUser, error) {
	token, err := s.rootToken(ctx)
	if err != nil {
		return nil, err
	}
	var result []ModeusUser
	response, err := s.modeus.FindStudents(token, fullName)
	if err != nil {
		if errors.Is(err, modeus.ErrStudentsNotFound) {
			return nil, ErrStudentsNotFound
		}
		log.Errorf("%s/FindAllUsers error find students from modeus: %s", serviceParserPrefixLog, err)
		return nil, err
	}
	for _, student := range response.Embedded.Students {
		u := ModeusUser{
			FullName:         "Ошибка",
			FlowCode:         student.FlowCode,
			SpecialtyName:    student.SpecialtyName,
			SpecialtyProfile: student.SpecialtyProfile,
			ScheduleId:       student.PersonId,
			GradesId:         student.Id,
		}
		// Идем доп циклом потому что порядок в persons и students иногда не совпадает
		for _, person := range response.Embedded.Persons {
			if person.Id == student.PersonId {
				u.FullName = person.FullName
				break
			}
		}
		result = append(result, u)
	}
	return result, nil
}

func (s *Service) FindUserById(ctx context.Context, scheduleId string) (ModeusUser, error) {
	token, err := s.rootToken(ctx)
	if err != nil {
		return ModeusUser{}, err
	}
	student, err := s.modeus.FindStudentById(token, scheduleId)
	if err != nil {
		return ModeusUser{}, err
	}
	if student.Page.TotalElements != 1 {
		log.Errorf("%s/FindUserById error find user by id: more than 1 was found", serviceParserPrefixLog)
		return ModeusUser{}, errors.New("students with specified id more than 1")
	}
	user := student.Embedded.Students[0]
	return ModeusUser{
		FullName:         student.Embedded.Persons[0].FullName,
		FlowCode:         user.FlowCode,
		SpecialtyName:    user.SpecialtyName,
		SpecialtyProfile: user.SpecialtyProfile,
	}, nil
}
