package service

import (
	"bot_for_modeus/internal/repo"
	"bot_for_modeus/pkg/parser"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	serviceSchedulePrefixLog = "/service/schedule"

	// TODO возможно, нужно намного больше, потому что при большой нагрузке и низкой скорости интернета на сервере
	//  некоторые окна селениума не успевают загрузиться до таймаута
	defaultParserTimeout = time.Minute
)

type ParserService struct {
	repo      repo.User
	parser    parser.Parser
	rootLogin string
	rootPass  string
}

func newParserService(repo repo.User, parser parser.Parser, rootLogin, rootPass string) *ParserService {
	return &ParserService{
		repo:      repo,
		parser:    parser,
		rootLogin: rootLogin,
		rootPass:  rootPass,
	}
}

// TODO добавить кастомные ошибки для расписания (ну типо там время вышло туда сюда)
//  Возможно, есть смысл отслеживать timeout exception, потому что она вылетает при большом количестве запросов

func (s *ParserService) DaySchedule(ctx context.Context, userId int64) (string, error) {
	driver, err := s.parser.InitRemote()
	if err != nil {
		return "", err
	}
	defer func() { _ = driver.Quit() }()
	u, err := s.repo.GetUserById(ctx, userId)
	if err != nil {
		return "", ErrUserNotFound
	}
	if len(u.Login) != 0 && len(u.Password) != 0 {
		data, err := s.parser.DayScheduleWithLoginPass(driver, u.Login, u.Password, defaultParserTimeout)
		if err != nil {
			if errors.Is(err, parser.ErrIncorrectUserData) {
				return "", ErrUserPermissionDenied
			}
			log.Errorf("%s/DaySchedule error parse day schedule with login and password: %s", serviceSchedulePrefixLog, err)
			return "", err
		}
		return data, nil
	}
	data, err := s.parser.DaySchedule(driver, s.rootLogin, s.rootPass, u.FullName, defaultParserTimeout)
	if err != nil {
		if errors.Is(err, parser.ErrIncorrectFullName) {
			return "", ErrUserIncorrectFullName
		}
		log.Errorf("%s/DaySchedule error parse day schedule with root login and password: %s", serviceSchedulePrefixLog, err)
		return "", err
	}
	return data, nil
}

func (s *ParserService) WeekSchedule(ctx context.Context, userId int64) (string, error) {
	driver, err := s.parser.InitRemote()
	if err != nil {
		return "", err
	}
	defer func() { _ = driver.Quit() }()
	u, err := s.repo.GetUserById(ctx, userId)
	if err != nil {
		return "", ErrUserNotFound
	}
	if len(u.Login) != 0 && len(u.Password) != 0 {
		data, err := s.parser.WeekScheduleWithLoginPass(driver, u.Login, u.Password, defaultParserTimeout)
		if err != nil {
			if errors.Is(err, parser.ErrIncorrectUserData) {
				return "", ErrUserPermissionDenied
			}
			log.Errorf("%s/WeekSchedule error parse week schedule with login and password: %s", serviceSchedulePrefixLog, err)
			return "", err
		}
		return data, nil
	}
	data, err := s.parser.WeekSchedule(driver, s.rootLogin, s.rootPass, u.FullName, defaultParserTimeout)
	if err != nil {
		if errors.Is(err, parser.ErrIncorrectFullName) {
			return "", ErrUserIncorrectFullName
		}
		log.Errorf("%s/WeekSchedule error parse week schedule with root login and password: %s", serviceSchedulePrefixLog, err)
		return "", err
	}
	return data, nil
}

func (s *ParserService) UserGrades(ctx context.Context, userId int64) (map[int]int, string, error) {
	driver, err := s.parser.InitRemote()
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = driver.Quit() }()
	u, err := s.repo.GetUserById(ctx, userId)
	if err != nil {
		return nil, "", ErrUserNotFound
	}
	if len(u.Login) == 0 && len(u.Password) == 0 {
		return nil, "", ErrUserPermissionDenied
	}
	subjects, data, err := s.parser.UserGrades(driver, u.Login, u.Password, defaultParserTimeout)
	if err != nil {
		if errors.Is(err, parser.ErrIncorrectUserData) {
			return nil, "", ErrUserPermissionDenied
		}
		log.Errorf("%s/UserGrades error parse user grades: %s", serviceUserPrefixLog, err)
		return nil, "", err
	}
	return subjects, data, nil
}

func (s *ParserService) SubjectGradesInfo(ctx context.Context, userId int64, index int) (string, error) {
	driver, err := s.parser.InitRemote()
	if err != nil {
		return "", err
	}
	defer func() { _ = driver.Quit() }()
	u, err := s.repo.GetUserById(ctx, userId)
	if err != nil {
		return "", ErrUserNotFound
	}
	if len(u.Login) == 0 && len(u.Password) == 0 {
		return "", ErrUserPermissionDenied
	}
	data, err := s.parser.SubjectDetailedInfo(driver, u.Login, u.Password, index, defaultParserTimeout)
	if err != nil {
		if errors.Is(err, parser.ErrIncorrectUserData) {
			return "", ErrUserPermissionDenied
		}
		log.Errorf("%s/SubjectGradesInfo error parse subject grades info: %s", serviceUserPrefixLog, err)
		return "", err
	}
	return data, nil
}

func (s *ParserService) OtherStudentDaySchedule(fullName string) (string, error) {
	driver, err := s.parser.InitRemote()
	if err != nil {
		return "", err
	}
	defer func() { _ = driver.Quit() }()
	data, err := s.parser.DaySchedule(driver, s.rootLogin, s.rootPass, fullName, defaultParserTimeout)
	if err != nil {
		if errors.Is(err, parser.ErrIncorrectFullName) {
			return "", ErrUserIncorrectFullName
		}
		log.Errorf("%s/OtherStudentDaySchedule error parse day schedule: %s", serviceSchedulePrefixLog, err)
		return "", err
	}
	return data, nil
}

func (s *ParserService) OtherStudentWeekSchedule(fullName string) (string, error) {
	driver, err := s.parser.InitRemote()
	if err != nil {
		return "", err
	}
	defer func() { _ = driver.Quit() }()
	data, err := s.parser.WeekSchedule(driver, s.rootLogin, s.rootPass, fullName, defaultParserTimeout)
	if err != nil {
		if errors.Is(err, parser.ErrIncorrectFullName) {
			return "", ErrUserIncorrectFullName
		}
		log.Errorf("%s/OtherStudentDaySchedule error parse day schedule: %s", serviceSchedulePrefixLog, err)
		return "", err
	}
	return data, nil
}
