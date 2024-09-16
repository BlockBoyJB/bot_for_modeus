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
	"regexp"
)

const (
	userServicePrefixLog = "/service/user"
)

// Регулярка для первичной проверки корректности почты.
var (
	emailRegex = regexp.MustCompile(`^stud\d{10}@study\.utmn\.ru$`)
)

type userService struct {
	user    repo.User
	crypter crypter.Crypter
}

func newUserService(user repo.User, crypter crypter.Crypter) *userService {
	return &userService{
		user:    user,
		crypter: crypter,
	}
}

func (s *userService) Create(ctx context.Context, input UserInput) error {
	u, err := s.user.FindById(ctx, input.UserId)
	if u.UserId != 0 && err == nil {
		return ErrUserAlreadyExists
	}

	if err != nil {
		if !errors.Is(err, mongoerrs.ErrNotFound) {
			log.Errorf("%s/Create error find user: %s", userServicePrefixLog, err)
			return err
		}
	}

	err = s.user.Create(ctx, dbmodel.User{
		UserId:     input.UserId,
		FullName:   input.FullName,
		ScheduleId: input.ScheduleId,
		GradesId:   input.GradesId,
		Friends:    map[string]string{},
	})
	if err != nil {
		log.Errorf("%s/Create error create user: %s", userServicePrefixLog, err)
		return err
	}
	return nil
}

func (s *userService) Find(ctx context.Context, userId int64) (UserOutput, error) {
	u, err := s.user.FindById(ctx, userId)
	if err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return UserOutput{}, ErrUserNotFound
		}
		log.Errorf("%s/Find error find user: %s", userServicePrefixLog, err)
		return UserOutput{}, err
	}
	user := UserOutput{
		FullName:   u.FullName,
		Login:      u.Login,
		Password:   u.Password,
		ScheduleId: u.ScheduleId,
		GradesId:   u.GradesId,
		Friends:    u.Friends,
	}
	if u.Password != "" {
		user.Password, err = s.crypter.Decrypt(u.Password)
		if err != nil {
			log.Errorf("%s/Find user error decrypt password: %s", userServicePrefixLog, err)
			return UserOutput{}, err
		}
	}
	return user, nil
}

func (s *userService) UpdateLoginPassword(ctx context.Context, input UserLoginPasswordInput) error {
	if !emailRegex.MatchString(input.Login) {
		return ErrUserIncorrectLogin
	}
	password, err := s.crypter.Encrypt(input.Password)
	if err != nil {
		log.Errorf("%s/UpdateLoginPassword error encrypt password: %s", userServicePrefixLog, err)
		return err
	}
	update := bson.D{{"$set", bson.D{{"login", input.Login}, {"password", password}}}}
	if err = s.user.Update(ctx, input.UserId, update); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Errorf("%s/UpdateLoginPassword error update user login and password: %s", userServicePrefixLog, err)
		return err
	}
	return nil
}

func (s *userService) UpdateInfo(ctx context.Context, input UserInput) error {
	update := bson.D{{"$set", bson.D{
		{"full_name", input.FullName},
		{"schedule_id", input.ScheduleId},
		{"grades_id", input.GradesId},
	}}}
	if err := s.user.Update(ctx, input.UserId, update); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Errorf("%s/UpdateInfo error update user info: %s", userServicePrefixLog, err)
		return err
	}
	return nil
}

func (s *userService) Delete(ctx context.Context, userId int64) error {
	if err := s.user.Delete(ctx, userId); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Errorf("%s/Delete error delete user: %s", userServicePrefixLog, err)
		return err
	}
	return nil
}

func (s *userService) AddFriend(ctx context.Context, input FriendInput) error {
	update := bson.D{{"$set", bson.D{{"friends." + input.ScheduleId, input.FullName}}}}
	if err := s.user.Update(ctx, input.UserId, update); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Errorf("%s/AddFriend error add user friend: %s", userServicePrefixLog, err)
		return err
	}
	return nil
}

func (s *userService) DeleteFriend(ctx context.Context, input FriendInput) error {
	update := bson.D{{"$unset", bson.D{{"friends." + input.ScheduleId, ""}}}}
	if err := s.user.Update(ctx, input.UserId, update); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Errorf("%s/DeleteFriend error delete user friend: %s", userServicePrefixLog, err)
		return err
	}
	return nil
}
