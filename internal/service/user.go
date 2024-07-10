package service

import (
	"bot_for_modeus/internal/model/dbmodel"
	"bot_for_modeus/internal/repo"
	"bot_for_modeus/internal/repo/mongoerrs"
	"bot_for_modeus/pkg/crypter"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

const serviceUserPrefixLog = "/service/user"

type userService struct {
	repo    repo.User
	crypter crypter.PasswordCrypter
}

func newUserService(repo repo.User, crypter crypter.PasswordCrypter) *userService {
	return &userService{repo: repo, crypter: crypter}
}

func (s *userService) CreateUser(ctx context.Context, input UserInput) error {
	u, err := s.repo.GetUserById(ctx, input.UserId)
	if u.UserId != 0 && err == nil {
		return ErrUserAlreadyExists
	}
	if !errors.Is(err, mongoerrs.ErrNotFound) {
		log.Errorf("%s/CreateUser error finding user: %s", serviceUserPrefixLog, err)
		return err
	}
	if err = s.repo.CreateUser(ctx, dbmodel.User{
		UserId:     input.UserId,
		FullName:   input.FullName,
		ScheduleId: input.ScheduleId,
		GradesId:   input.GradesId,
		Friends:    map[string]string{},
	}); err != nil {
		log.Errorf("%s/CreateUser error create user: %s", serviceUserPrefixLog, err)
		return err
	}
	return nil
}

func (s *userService) FindUser(ctx context.Context, userId int64) (dbmodel.User, error) {
	user, err := s.repo.GetUserById(ctx, userId)
	if err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return dbmodel.User{}, ErrUserNotFound
		}
		log.Errorf("%s/FindUser error find user: %s", serviceUserPrefixLog, err)
		return dbmodel.User{}, err
	}
	if user.Password != "" {
		user.Password, err = s.crypter.Decrypt(user.Password)
		if err != nil {
			log.Errorf("%s/FindUser error decrypt password: %s", serviceUserPrefixLog, err)
			return dbmodel.User{}, err
		}
	}
	return user, nil
}

func (s *userService) UpdateLoginPassword(ctx context.Context, userId int64, login, password string) error {
	password, err := s.crypter.Encrypt(password)
	if err != nil {
		log.Errorf("%s/UpdateLoginPassword error encrypt password: %s", serviceUserPrefixLog, err)
		return err
	}
	update := bson.D{{"$set", bson.D{{"login", login}, {"password", password}}}}
	if err = s.repo.UpdateData(ctx, userId, update); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Errorf("%s/UpdateLoginPassword error update user login and password: %s", serviceUserPrefixLog, err)
		return err
	}
	return nil
}

func (s *userService) UpdateUserInformation(ctx context.Context, input UserInput) error {
	update := bson.D{{"$set", bson.D{{"full_name", input.FullName}, {"schedule_id", input.ScheduleId}, {"grades_id", input.GradesId}}}}
	if err := s.repo.UpdateData(ctx, input.UserId, update); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Errorf("%s/UpdateUserInformation error update user data: %s", serviceUserPrefixLog, err)
		return err
	}
	return nil
}

func (s *userService) AddFriend(ctx context.Context, userId int64, fullName, scheduleId string) error {
	update := bson.D{{"$set", bson.D{{"friends." + scheduleId, fullName}}}}
	if err := s.repo.UpdateData(ctx, userId, update); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Errorf("%s/AddFriend error add friend to friends object: %s", serviceUserPrefixLog, err)
		return err
	}
	return nil
}

func (s *userService) DeleteFriend(ctx context.Context, userId int64, scheduleId string) error {
	update := bson.D{{"$unset", bson.D{{"friends." + scheduleId, ""}}}}
	if err := s.repo.UpdateData(ctx, userId, update); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Errorf("%s/DeleteFriend error remove friend from friends object: %s", serviceUserPrefixLog, err)
		return err
	}
	return nil
}

func (s *userService) DeleteUser(ctx context.Context, userId int64) error {
	if err := s.repo.DeleteUser(ctx, userId); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Errorf("%s/DeleteUser error delete user: %s", serviceUserPrefixLog, err)
		return err
	}
	return nil
}
