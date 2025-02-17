package service

import (
	"bot_for_modeus/internal/model/dbmodel"
	"bot_for_modeus/internal/repo"
	"bot_for_modeus/internal/repo/mongoerrs"
	"bot_for_modeus/pkg/crypter"
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"regexp"
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
			log.Err(err).Int64("user_id", input.UserId).Msg("user/Create error find user by id")
			return err
		}
	}

	err = s.user.Create(ctx, dbmodel.User{
		UserId:     input.UserId,
		FullName:   input.FullName,
		ScheduleId: input.ScheduleId,
		GradesId:   input.GradesId,
		Friends:    []dbmodel.Friend{},
	})
	if err != nil {
		log.Err(err).Interface("input", input).Msg("user/Create error create user in database")
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
		log.Err(err).Int64("user_id", userId).Msg("user/Find error find user by id")
		return UserOutput{}, err
	}
	o := UserOutput{
		FullName:   u.FullName,
		Login:      u.Login,
		Password:   u.Password,
		ScheduleId: u.ScheduleId,
		GradesId:   u.GradesId,
		Friends:    make([]FriendOutput, 0, len(u.Friends)),
	}
	for _, f := range u.Friends {
		o.Friends = append(o.Friends, FriendOutput{
			FullName:   f.FullName,
			ScheduleId: f.ScheduleId,
		})
	}
	return o, nil
}

func (s *userService) UpdateLoginPassword(ctx context.Context, input UserLoginPasswordInput) error {
	if !emailRegex.MatchString(input.Login) {
		return ErrUserIncorrectLogin
	}
	password, err := s.crypter.Encrypt(input.Password)
	if err != nil {
		log.Err(err).Int64("user_id", input.UserId).Msg("user/UpdateLoginPassword error encrypt password")
		return err
	}
	update := bson.D{{"$set", bson.D{{"login", input.Login}, {"password", password}}}}
	if err = s.user.Update(ctx, input.UserId, update); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Err(err).Int64("user_id", input.UserId).Msg("user/UpdateLoginPassword error update login and password in database")
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
		log.Err(err).Interface("input", input).Msg("user/UpdateInfo error update user in database")
		return err
	}
	return nil
}

func (s *userService) Delete(ctx context.Context, userId int64) error {
	if err := s.user.Delete(ctx, userId); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Err(err).Int64("user_id", userId).Msg("user/Delete error delete user in database")
		return err
	}
	return nil
}

func (s *userService) AddFriend(ctx context.Context, input FriendInput) error {
	update := bson.D{{"$push", bson.M{"friends": dbmodel.Friend{
		FullName:   input.FullName,
		ScheduleId: input.ScheduleId,
	}}}}
	if err := s.user.Update(ctx, input.UserId, update); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Err(err).Int64("user_id", input.UserId).Str("schedule_id", input.ScheduleId).Str("full_name", input.FullName).
			Msg("user/AddFriend error add friend to user in database")
		return err
	}
	return nil
}

func (s *userService) DeleteFriend(ctx context.Context, input FriendInput) error {
	// Нужно удалить объект из массива friends по совпадению schedule_id
	update := bson.D{{"$pull", bson.M{"friends": bson.M{"schedule_id": input.ScheduleId}}}}

	if err := s.user.Update(ctx, input.UserId, update); err != nil {
		if errors.Is(err, mongoerrs.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Err(err).Int64("user_id", input.UserId).Str("schedule_id", input.ScheduleId).
			Msg("user/DeleteFriend error delete user friend in database")
		return err
	}
	return nil
}

func (s *userService) Decrypt(input string) (string, error) {
	d, err := s.crypter.Decrypt(input)
	if err != nil {
		log.Err(err).Msg("user/Decrypt error decrypt data")
		return "", err
	}
	return d, nil
}
