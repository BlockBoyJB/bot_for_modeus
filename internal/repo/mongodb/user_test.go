package mongodb

import (
	"bot_for_modeus/internal/model/dbmodel"
	"bot_for_modeus/internal/repo/mongoerrs"
	"go.mongodb.org/mongo-driver/bson"
	mgo "go.mongodb.org/mongo-driver/mongo"
)

func (s *mongodbTestSuite) TestUserRepo_Create() {
	testCases := []struct {
		testName  string
		user      dbmodel.User
		expectErr error
	}{
		{
			testName: "correct test",
			user: dbmodel.User{
				UserId:     1,
				FullName:   "vasya",
				Login:      "login",
				Password:   "password",
				ScheduleId: "abc",
				GradesId:   "abc",
				Friends:    []dbmodel.Friend{},
			},
			expectErr: nil,
		},
	}

	for _, tc := range testCases {
		err := s.user.Create(s.ctx, tc.user)
		s.Assert().Equal(tc.expectErr, err)

		if tc.expectErr == nil {
			var actualUser dbmodel.User
			err = s.user.pool.FindOne(s.ctx, bson.D{{"user_id", tc.user.UserId}}).Decode(&actualUser)
			s.Assert().Nil(err)

			s.Assert().Equal(tc.user, actualUser)
		}
	}
}

func (s *mongodbTestSuite) TestUserRepo_FindById() {
	user := dbmodel.User{
		UserId:     1,
		FullName:   "vasya",
		Login:      "login",
		Password:   "password",
		ScheduleId: "abc",
		GradesId:   "abc",
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
	}
	if _, err := s.user.pool.InsertOne(s.ctx, user); err != nil {
		panic(err)
	}

	testCases := []struct {
		testName   string
		userId     int64
		expectUser dbmodel.User
		expectErr  error
	}{
		{
			testName:   "correct test",
			userId:     user.UserId,
			expectUser: user,
			expectErr:  nil,
		},
		{
			testName:  "user not exist",
			userId:    1123123,
			expectErr: mongoerrs.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		u, err := s.user.FindById(s.ctx, tc.userId)
		s.Assert().Equal(tc.expectErr, err)
		s.Assert().Equal(tc.expectUser, u)
	}
}

func (s *mongodbTestSuite) TestUserRepo_Update() {
	user := dbmodel.User{
		UserId:     1,
		FullName:   "vasya",
		Login:      "login",
		Password:   "password",
		ScheduleId: "abc",
		GradesId:   "abc",
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
	}
	if _, err := s.user.pool.InsertOne(s.ctx, user); err != nil {
		panic(err)
	}

	testCases := []struct {
		testName   string
		userId     int64
		update     bson.D
		expectUser dbmodel.User
		expectErr  error
	}{
		{
			testName: "correct test",
			userId:   user.UserId,
			update:   bson.D{{"$set", bson.D{{"full_name", "petya"}}}},
			expectUser: dbmodel.User{
				UserId:     user.UserId,
				FullName:   "petya",
				Login:      user.Login,
				Password:   user.Password,
				ScheduleId: user.ScheduleId,
				GradesId:   user.GradesId,
				Friends:    user.Friends,
			},
			expectErr: nil,
		},
		{
			testName:  "user not exist",
			userId:    13123123,
			update:    bson.D{{"$set", bson.D{{"full_name", "petya"}}}},
			expectErr: mongoerrs.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		err := s.user.Update(s.ctx, tc.userId, tc.update)
		s.Assert().Equal(tc.expectErr, err)

		if tc.expectErr == nil {
			var actualUser dbmodel.User
			err = s.user.pool.FindOne(s.ctx, bson.D{{"user_id", tc.userId}}).Decode(&actualUser)
			s.Assert().Nil(err)
			s.Assert().Equal(tc.expectUser, actualUser)
		}
	}
}

func (s *mongodbTestSuite) TestUserRepo_Delete() {
	user := dbmodel.User{
		UserId:     1,
		FullName:   "vasya",
		Login:      "login",
		Password:   "password",
		ScheduleId: "abc",
		GradesId:   "abc",
		Friends:    []dbmodel.Friend{},
	}
	if _, err := s.user.pool.InsertOne(s.ctx, user); err != nil {
		panic(err)
	}

	testCases := []struct {
		testName  string
		userId    int64
		expectErr error
	}{
		{
			testName:  "correct test",
			userId:    user.UserId,
			expectErr: nil,
		},
		{
			testName:  "user not exits",
			userId:    1231231,
			expectErr: mongoerrs.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		err := s.user.Delete(s.ctx, tc.userId)
		s.Assert().Equal(tc.expectErr, err)

		if tc.expectErr == nil {
			err = s.user.pool.FindOne(s.ctx, bson.D{{"user_id", tc.userId}}).Err()
			s.Assert().Equal(mgo.ErrNoDocuments, err)
		}
	}
}
