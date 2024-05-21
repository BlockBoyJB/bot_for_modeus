package service

import (
	"bot_for_modeus/internal/model/pgmodel"
	"bot_for_modeus/internal/repo"
	"bot_for_modeus/internal/repo/pgerrs"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"strings"
)

const serviceUserPrefixLog = "/service/user"

type UserService struct {
	repo repo.User
}

func newUserService(repo repo.User) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, userId int64, fullName string) error {
	n := strings.Fields(fullName)
	name := n[0]
	for i := 1; i < len(n); i++ {
		name += " " + n[i]
	}
	// операцию выше проводим, потому что поисковик по ФИО очень нежный и реагирует на каждый лишний пробел,
	// поэтому формируем нормальную строчку
	err := s.repo.CreateUser(ctx, pgmodel.User{
		UserId:   userId,
		FullName: name,
	})
	if err != nil {
		if errors.Is(err, pgerrs.ErrAlreadyExists) {
			return ErrUserAlreadyExists
		} else {
			log.Errorf("%s/CreateUser error create user: %s", serviceUserPrefixLog, err)
			return ErrUserCannotCreate
		}
	}
	return nil
}

func (s *UserService) UpdateUserFullName(ctx context.Context, userId int64, fullName string) error {
	if err := s.repo.UpdateFullName(ctx, userId, fullName); err != nil {
		if errors.Is(err, pgerrs.ErrNotFound) {
			return ErrUserNotFound
		} else {
			log.Errorf("%s/UpdateUserFullName error update full name: %s", serviceUserPrefixLog, err)
			return ErrUserCannotUpdate
		}
	}
	return nil
}

func (s *UserService) AddUserLoginPassword(ctx context.Context, userId int64, login, password string) error {
	if err := s.repo.AddLoginPassword(ctx, userId, login, password); err != nil {
		if errors.Is(err, pgerrs.ErrNotFound) {
			return ErrUserNotFound
		} else {
			log.Errorf("%s/AddUserLoginPassword error add login and password: %s", serviceUserPrefixLog, err)
			return ErrUserCannotUpdate
		}
	}
	return nil
}

func (s *UserService) DeleteLoginPassword(ctx context.Context, userId int64) error {
	if err := s.repo.DeleteLoginPassword(ctx, userId); err != nil {
		if errors.Is(err, pgerrs.ErrNotFound) {
			return ErrUserNotFound
		} else {
			log.Errorf("%s/DeleteLoginPassword error delete login and password: %s", serviceUserPrefixLog, err)
			return ErrUserCannotUpdate
		}
	}
	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, userId int64) error {
	if err := s.repo.DeleteUser(ctx, userId); err != nil {
		if errors.Is(err, pgerrs.ErrNotFound) {
			return ErrUserNotFound
		} else {
			log.Errorf("%s/DeleteUser error delete user: %s", serviceUserPrefixLog, err)
			return err
		}
	}
	return nil
}
