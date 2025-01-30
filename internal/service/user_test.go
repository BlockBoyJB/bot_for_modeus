package service

import (
	"bot_for_modeus/internal/mocks/cryptermocks"
	"bot_for_modeus/internal/mocks/repomocks"
	"bot_for_modeus/internal/model/dbmodel"
	"bot_for_modeus/internal/repo/mongoerrs"
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func TestUserService_Create(t *testing.T) {
	type args struct {
		ctx   context.Context
		input UserInput
	}

	type mockBehaviour func(u *repomocks.MockUser, a args)

	testCases := []struct {
		testName      string
		args          args
		mockBehaviour mockBehaviour
		expectErr     error
	}{
		{
			testName: "correct test",
			args: args{
				ctx: context.Background(),
				input: UserInput{
					UserId:     1,
					FullName:   "vasya",
					ScheduleId: "foobar",
					GradesId:   "foobar",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().FindById(a.ctx, a.input.UserId).Return(dbmodel.User{}, mongoerrs.ErrNotFound)
				u.EXPECT().Create(a.ctx, dbmodel.User{
					UserId:     a.input.UserId,
					FullName:   a.input.FullName,
					ScheduleId: a.input.ScheduleId,
					GradesId:   a.input.GradesId,
					Friends:    []dbmodel.Friend{},
				}).Return(nil)
			},
			expectErr: nil,
		},
		{
			testName: "user already exist",
			args: args{
				ctx: context.Background(),
				input: UserInput{
					UserId:     1,
					FullName:   "vasya",
					ScheduleId: "foobar",
					GradesId:   "foobar",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().FindById(a.ctx, a.input.UserId).Return(dbmodel.User{UserId: 1}, nil)
			},
			expectErr: ErrUserAlreadyExists,
		},
		{
			testName: "unexpected user create error",
			args: args{
				ctx: context.Background(),
				input: UserInput{
					UserId:     1,
					FullName:   "vasya",
					ScheduleId: "foobar",
					GradesId:   "foobar",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().FindById(a.ctx, a.input.UserId).Return(dbmodel.User{}, mongoerrs.ErrNotFound)
				u.EXPECT().Create(a.ctx, dbmodel.User{
					UserId:     1,
					FullName:   a.input.FullName,
					ScheduleId: a.input.ScheduleId,
					GradesId:   a.input.GradesId,
					Friends:    []dbmodel.Friend{},
				}).Return(errors.New("unexpected error"))
			},
			expectErr: errors.New("unexpected error"),
		},
		{
			testName: "unexpected user find error",
			args: args{
				ctx: context.Background(),
				input: UserInput{
					UserId:     1,
					FullName:   "vasya",
					ScheduleId: "foobar",
					GradesId:   "foobar",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().FindById(a.ctx, a.input.UserId).Return(dbmodel.User{}, errors.New("unexpected error"))
			},
			expectErr: errors.New("unexpected error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			user := repomocks.NewMockUser(ctrl)
			tc.mockBehaviour(user, tc.args)

			s := newUserService(user, nil)

			err := s.Create(tc.args.ctx, tc.args.input)
			assert.Equal(t, tc.expectErr, err)
		})
	}
}

func TestUserService_Find(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}

	type mockBehaviour func(u *repomocks.MockUser, c *cryptermocks.MockCrypter, a args)

	testCases := []struct {
		testName      string
		args          args
		mockBehaviour mockBehaviour
		expectOutput  UserOutput
		expectErr     error
	}{
		{
			testName: "correct test",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehaviour: func(u *repomocks.MockUser, c *cryptermocks.MockCrypter, a args) {
				u.EXPECT().FindById(a.ctx, a.userId).Return(dbmodel.User{
					UserId:     a.userId,
					FullName:   "vasya",
					Login:      "foo",
					Password:   "crypt_password",
					ScheduleId: "foobar",
					GradesId:   "foobar",
					Friends: []dbmodel.Friend{
						{
							FullName:   "Иванов Иван Иванович",
							ScheduleId: "a07bd176-2cea-405a-8f69-baa82c28f089",
						},
						{
							FullName:   "Петров Петр Петрович",
							ScheduleId: "cecea70d-809b-4b5c-89eb-75de829352ea",
						},
					},
				}, nil)
			},
			expectOutput: UserOutput{
				FullName:   "vasya",
				Login:      "foo",
				Password:   "crypt_password",
				ScheduleId: "foobar",
				GradesId:   "foobar",
				Friends: []FriendOutput{
					{
						FullName:   "Иванов Иван Иванович",
						ScheduleId: "a07bd176-2cea-405a-8f69-baa82c28f089",
					},
					{
						FullName:   "Петров Петр Петрович",
						ScheduleId: "cecea70d-809b-4b5c-89eb-75de829352ea",
					},
				},
			},
			expectErr: nil,
		},
		{
			testName: "user not exist",
			args: args{
				ctx:    context.Background(),
				userId: 123,
			},
			mockBehaviour: func(u *repomocks.MockUser, c *cryptermocks.MockCrypter, a args) {
				u.EXPECT().FindById(a.ctx, a.userId).Return(dbmodel.User{}, mongoerrs.ErrNotFound)
			},
			expectErr: ErrUserNotFound,
		},
		{
			testName: "correct test user without password",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehaviour: func(u *repomocks.MockUser, c *cryptermocks.MockCrypter, a args) {
				u.EXPECT().FindById(a.ctx, a.userId).Return(dbmodel.User{
					UserId:     a.userId,
					FullName:   "vasya",
					ScheduleId: "foobar",
					GradesId:   "foobar",
					Friends:    []dbmodel.Friend{},
				}, nil)
			},
			expectOutput: UserOutput{
				FullName:   "vasya",
				ScheduleId: "foobar",
				GradesId:   "foobar",
				Friends:    []FriendOutput{},
			},
			expectErr: nil,
		},
		{
			testName: "unexpected user find error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehaviour: func(u *repomocks.MockUser, c *cryptermocks.MockCrypter, a args) {
				u.EXPECT().FindById(a.ctx, a.userId).Return(dbmodel.User{}, errors.New("unexpected error"))
			},
			expectErr: errors.New("unexpected error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			user := repomocks.NewMockUser(ctrl)
			crypt := cryptermocks.NewMockCrypter(ctrl)
			tc.mockBehaviour(user, crypt, tc.args)

			s := newUserService(user, crypt)

			output, err := s.Find(tc.args.ctx, tc.args.userId)
			assert.Equal(t, tc.expectOutput, output)
			assert.Equal(t, tc.expectErr, err)
		})
	}
}

func TestUserService_UpdateLoginPassword(t *testing.T) {
	type args struct {
		ctx   context.Context
		input UserLoginPasswordInput
	}

	type mockBehaviour func(u *repomocks.MockUser, c *cryptermocks.MockCrypter, a args)

	testCases := []struct {
		testName      string
		args          args
		mockBehaviour mockBehaviour
		expectErr     error
	}{
		{
			testName: "correct test",
			args: args{
				ctx: context.Background(),
				input: UserLoginPasswordInput{
					UserId:   1,
					Login:    "stud0000000000@study.utmn.ru",
					Password: "password",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, c *cryptermocks.MockCrypter, a args) {
				c.EXPECT().Encrypt(a.input.Password).Return("crypt_password", nil)
				u.EXPECT().Update(a.ctx, a.input.UserId,
					bson.D{{"$set", bson.D{{"login", a.input.Login}, {"password", "crypt_password"}}}},
				).Return(nil)
			},
			expectErr: nil,
		},
		{
			testName: "incorrect login input",
			args: args{
				ctx: context.Background(),
				input: UserLoginPasswordInput{
					UserId:   1,
					Login:    "foobar",
					Password: "password",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, c *cryptermocks.MockCrypter, a args) {},
			expectErr:     ErrUserIncorrectLogin,
		},
		{
			testName: "user not exist",
			args: args{
				ctx: context.Background(),
				input: UserLoginPasswordInput{
					UserId:   123123,
					Login:    "stud0000000001@study.utmn.ru",
					Password: "password",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, c *cryptermocks.MockCrypter, a args) {
				c.EXPECT().Encrypt(a.input.Password).Return("crypt_password", nil)
				u.EXPECT().Update(a.ctx, a.input.UserId,
					bson.D{{"$set", bson.D{{"login", a.input.Login}, {"password", "crypt_password"}}}},
				).Return(mongoerrs.ErrNotFound)
			},
			expectErr: ErrUserNotFound,
		},
		{
			testName: "unexpected user update error",
			args: args{
				ctx: context.Background(),
				input: UserLoginPasswordInput{
					UserId:   123123,
					Login:    "stud0000000000@study.utmn.ru",
					Password: "password",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, c *cryptermocks.MockCrypter, a args) {
				c.EXPECT().Encrypt(a.input.Password).Return("crypt_password", nil)
				u.EXPECT().Update(a.ctx, a.input.UserId,
					bson.D{{"$set", bson.D{{"login", a.input.Login}, {"password", "crypt_password"}}}},
				).Return(errors.New("unexpected error"))
			},
			expectErr: errors.New("unexpected error"),
		},
		{
			testName: "unexpected crypter encrypt error",
			args: args{
				ctx: context.Background(),
				input: UserLoginPasswordInput{
					UserId:   1,
					Login:    "stud0000000000@study.utmn.ru",
					Password: "password",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, c *cryptermocks.MockCrypter, a args) {
				c.EXPECT().Encrypt(a.input.Password).Return("", errors.New("unexpected error"))
			},
			expectErr: errors.New("unexpected error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			user := repomocks.NewMockUser(ctrl)
			crypt := cryptermocks.NewMockCrypter(ctrl)
			tc.mockBehaviour(user, crypt, tc.args)

			s := newUserService(user, crypt)

			err := s.UpdateLoginPassword(tc.args.ctx, tc.args.input)
			assert.Equal(t, tc.expectErr, err)
		})
	}
}

func TestUserService_UpdateInfo(t *testing.T) {
	type args struct {
		ctx   context.Context
		input UserInput
	}

	type mockBehaviour func(u *repomocks.MockUser, a args)

	testCases := []struct {
		testName      string
		args          args
		mockBehaviour mockBehaviour
		expectErr     error
	}{
		{
			testName: "correct test",
			args: args{
				ctx: context.Background(),
				input: UserInput{
					UserId:     1,
					FullName:   "vasya",
					ScheduleId: "foobar",
					GradesId:   "foobar",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().Update(a.ctx, a.input.UserId, bson.D{{"$set", bson.D{
					{"full_name", a.input.FullName},
					{"schedule_id", a.input.ScheduleId},
					{"grades_id", a.input.GradesId},
				}}}).Return(nil)
			},
			expectErr: nil,
		},
		{
			testName: "user not exist",
			args: args{
				ctx: context.Background(),
				input: UserInput{
					UserId:     1,
					FullName:   "vasya",
					ScheduleId: "foobar",
					GradesId:   "foobar",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().Update(a.ctx, a.input.UserId, bson.D{{"$set", bson.D{
					{"full_name", a.input.FullName},
					{"schedule_id", a.input.ScheduleId},
					{"grades_id", a.input.GradesId},
				}}}).Return(mongoerrs.ErrNotFound)
			},
			expectErr: ErrUserNotFound,
		},
		{
			testName: "unexpected error",
			args: args{
				ctx: context.Background(),
				input: UserInput{
					UserId:     1,
					FullName:   "vasya",
					ScheduleId: "foobar",
					GradesId:   "foobar",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().Update(a.ctx, a.input.UserId, bson.D{{"$set", bson.D{
					{"full_name", a.input.FullName},
					{"schedule_id", a.input.ScheduleId},
					{"grades_id", a.input.GradesId},
				}}}).Return(errors.New("unexpected error"))
			},
			expectErr: errors.New("unexpected error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			user := repomocks.NewMockUser(ctrl)
			tc.mockBehaviour(user, tc.args)

			s := newUserService(user, nil)

			err := s.UpdateInfo(tc.args.ctx, tc.args.input)
			assert.Equal(t, tc.expectErr, err)
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}

	type mockBehaviour func(u *repomocks.MockUser, a args)

	testCases := []struct {
		testName      string
		args          args
		mockBehaviour mockBehaviour
		expectErr     error
	}{
		{
			testName: "correct test",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().Delete(a.ctx, a.userId).Return(nil)
			},
			expectErr: nil,
		},
		{
			testName: "user not exist",
			args: args{
				ctx:    context.Background(),
				userId: 123132,
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().Delete(a.ctx, a.userId).Return(mongoerrs.ErrNotFound)
			},
			expectErr: ErrUserNotFound,
		},
		{
			testName: "unexpected error",
			args: args{
				ctx:    context.Background(),
				userId: 1,
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().Delete(a.ctx, a.userId).Return(errors.New("unexpected error"))
			},
			expectErr: errors.New("unexpected error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			user := repomocks.NewMockUser(ctrl)
			tc.mockBehaviour(user, tc.args)

			s := newUserService(user, nil)

			err := s.Delete(tc.args.ctx, tc.args.userId)
			assert.Equal(t, tc.expectErr, err)
		})
	}
}

func TestUserService_AddFriend(t *testing.T) {
	type args struct {
		ctx   context.Context
		input FriendInput
	}

	type mockBehaviour func(u *repomocks.MockUser, a args)

	testCases := []struct {
		testName      string
		args          args
		mockBehaviour mockBehaviour
		expectErr     error
	}{
		{
			testName: "correct test",
			args: args{
				ctx: context.Background(),
				input: FriendInput{
					UserId:     1,
					FullName:   "petya",
					ScheduleId: "foobar",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().Update(a.ctx, a.input.UserId,
					bson.D{{"$push", bson.M{"friends": dbmodel.Friend{
						FullName:   a.input.FullName,
						ScheduleId: a.input.ScheduleId,
					}}}}).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			testName: "user not exist",
			args: args{
				ctx: context.Background(),
				input: FriendInput{
					UserId:     123,
					FullName:   "petya",
					ScheduleId: "foobar",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().Update(a.ctx, a.input.UserId,
					bson.D{{"$push", bson.M{"friends": dbmodel.Friend{
						FullName:   a.input.FullName,
						ScheduleId: a.input.ScheduleId,
					}}}}).
					Return(mongoerrs.ErrNotFound)
			},
			expectErr: ErrUserNotFound,
		},
		{
			testName: "unexpected error",
			args: args{
				ctx: context.Background(),
				input: FriendInput{
					UserId:     1,
					FullName:   "petya",
					ScheduleId: "foobar",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().Update(a.ctx, a.input.UserId,
					bson.D{{"$push", bson.M{"friends": dbmodel.Friend{
						FullName:   a.input.FullName,
						ScheduleId: a.input.ScheduleId,
					}}}}).
					Return(errors.New("unexpected error"))
			},
			expectErr: errors.New("unexpected error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			user := repomocks.NewMockUser(ctrl)
			tc.mockBehaviour(user, tc.args)

			s := newUserService(user, nil)

			err := s.AddFriend(tc.args.ctx, tc.args.input)
			assert.Equal(t, tc.expectErr, err)
		})
	}
}

func TestUserService_DeleteFriend(t *testing.T) {
	type args struct {
		ctx   context.Context
		input FriendInput
	}

	type mockBehaviour func(u *repomocks.MockUser, a args)

	testCases := []struct {
		testName      string
		args          args
		mockBehaviour mockBehaviour
		expectErr     error
	}{
		{
			testName: "correct test",
			args: args{
				ctx: context.Background(),
				input: FriendInput{
					UserId:     1,
					ScheduleId: "foobar",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().Update(a.ctx, a.input.UserId,
					bson.D{{"$pull", bson.M{"friends": bson.M{"schedule_id": a.input.ScheduleId}}}}).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			testName: "user not exist",
			args: args{
				ctx: context.Background(),
				input: FriendInput{
					UserId:     123,
					ScheduleId: "foobar",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().Update(a.ctx, a.input.UserId,
					bson.D{{"$pull", bson.M{"friends": bson.M{"schedule_id": a.input.ScheduleId}}}}).
					Return(mongoerrs.ErrNotFound)
			},
			expectErr: ErrUserNotFound,
		},
		{
			testName: "unexpected error",
			args: args{
				ctx: context.Background(),
				input: FriendInput{
					UserId:     1,
					ScheduleId: "foobar",
				},
			},
			mockBehaviour: func(u *repomocks.MockUser, a args) {
				u.EXPECT().Update(a.ctx, a.input.UserId,
					bson.D{{"$pull", bson.M{"friends": bson.M{"schedule_id": a.input.ScheduleId}}}}).
					Return(errors.New("unexpected error"))
			},
			expectErr: errors.New("unexpected error"),
		},
	}

	for _, tc := range testCases {
		ctrl := gomock.NewController(t)

		user := repomocks.NewMockUser(ctrl)
		tc.mockBehaviour(user, tc.args)

		s := newUserService(user, nil)

		err := s.DeleteFriend(tc.args.ctx, tc.args.input)
		assert.Equal(t, tc.expectErr, err)
	}
}
