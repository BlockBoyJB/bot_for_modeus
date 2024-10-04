package bot

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type redisStorageTestSuite struct {
	suite.Suite
	ctx     context.Context
	redis   *redis.Client
	storage *redisStorage
}

func (s *redisStorageTestSuite) SetupTest() {
	testRedisUrl := "127.0.0.1:6379"
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: testRedisUrl,
	})
	s.ctx = ctx
	s.storage = newRedisStorage(ctx, rdb)
	s.redis = rdb
}

func (s *redisStorageTestSuite) TearDownTest() {
	_ = s.redis.Del(s.ctx, "*")
	_ = s.redis.Close()
}

func TestRedisStorage(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	suite.Run(t, new(redisStorageTestSuite))
}

func (s *redisStorageTestSuite) Test_setState() {
	testCases := []struct {
		testName  string
		id        int64
		state     string
		expectErr error
	}{
		{
			testName:  "correct test",
			id:        123,
			state:     "someState",
			expectErr: nil,
		},
	}

	for _, tc := range testCases {
		err := s.storage.setState(tc.id, tc.state)
		s.Assert().Equal(tc.expectErr, err)

		actualState, err := s.redis.Get(s.ctx, s.storage.stateKey(tc.id)).Result()
		s.Assert().Nil(err)

		s.Assert().Equal(tc.state, actualState)
	}
}

func (s *redisStorageTestSuite) Test_getState() {
	var (
		defaultState       = "foobar"
		defaultId    int64 = 123
	)
	if err := s.redis.Set(s.ctx, s.storage.stateKey(defaultId), defaultState, 0).Err(); err != nil {
		s.T().Fatalf("setup test data err: %s", err)
	}

	testCases := []struct {
		testName    string
		id          int64
		expectState string
		expectErr   error
	}{
		{
			testName:    "correct test",
			id:          defaultId,
			expectState: defaultState,
			expectErr:   nil,
		},
		{
			testName:    "id with state not exist",
			id:          999,
			expectState: "",
			expectErr:   ErrKeyNotExists,
		},
	}

	for _, tc := range testCases {
		state, err := s.storage.getState(tc.id)
		s.Assert().Equal(tc.expectState, state)
		s.Assert().Equal(tc.expectErr, err)
	}
}

func (s *redisStorageTestSuite) Test_setData() {
	var (
		defaultId  int64 = 1
		defaultKey       = "foobar"
	)

	testCases := []struct {
		testName string
		data     any
	}{
		{
			testName: "test string",
			data:     "hello world",
		},
		{
			testName: "test int",
			data:     123,
		},
		{
			testName: "test array",
			data:     []int{1, 2, 3, 4},
		},
		{
			testName: "test map",
			data: map[string]string{
				"message": "hello world",
				"foo":     "bar",
			},
		},
		{
			testName: "test some structure",
			data: struct {
				Name string
				Age  int
			}{
				Name: "Petya",
				Age:  25,
			},
		},
	}

	for _, tc := range testCases {
		err := s.storage.setData(defaultId, defaultKey, tc.data)
		s.Assert().Nil(err)

		b, err := s.redis.Get(s.ctx, s.storage.dataKey(defaultId, defaultKey)).Bytes()
		s.Assert().Nil(err)

		expectData, err := json.Marshal(tc.data)
		s.Assert().Nil(err)

		// разницы нет в каком варианте их сравнивать: слайс байтов или через тип данных + значение
		s.Assert().Equal(expectData, b)
	}
}

func (s *redisStorageTestSuite) Test_setTempData() {
	var (
		defaultId  int64 = 1
		defaultKey       = "foobar"
	)

	testCases := []struct {
		testName string
		data     any
		ttl      time.Duration
	}{
		{
			testName: "test string",
			data:     "hello world",
			ttl:      time.Minute,
		},
		{
			testName: "test int",
			data:     123,
			ttl:      time.Hour,
		},
		{
			testName: "test slice",
			data:     []int{1, 2, 10, 3},
			ttl:      time.Minute * 10,
		},
		{
			testName: "test map",
			data: map[int]string{
				1: "hello",
				2: "world",
			},
			ttl: time.Hour * 10,
		},
	}

	for _, tc := range testCases {
		err := s.storage.setTempData(defaultId, defaultKey, tc.data, tc.ttl)
		s.Assert().Nil(err)

		actualTTL, err := s.redis.TTL(s.ctx, s.storage.dataKey(defaultId, defaultKey)).Result()
		s.Assert().Nil(err)

		// проверка булевая, потому что может быть задержка между записью в бд
		s.Assert().True(tc.ttl-actualTTL <= time.Second*2, tc.testName)
	}
}

func (s *redisStorageTestSuite) Test_getData() {
	var (
		defaultId   int64 = 1
		defaultKey        = "foobar"
		defaultData       = map[string]string{
			"message": "hello world",
		}
	)

	b, err := json.Marshal(defaultData)
	if err != nil {
		s.T().Fatalf("marhsall test data err: %s", err)
	}

	if err = s.redis.Set(s.ctx, s.storage.dataKey(defaultId, defaultKey), b, 0).Err(); err != nil {
		s.T().Fatalf("setup test data into redis err: %s", err)
	}

	testCases := []struct {
		testName   string
		id         int64
		key        string
		expectErr  error
		expectData map[string]string
	}{
		{
			testName:   "correct test",
			id:         defaultId,
			key:        defaultKey,
			expectErr:  nil,
			expectData: defaultData,
		},
		{
			testName:  "id not exist",
			id:        999,
			key:       defaultKey,
			expectErr: ErrKeyNotExists,
		},
		{
			testName:  "key not exist",
			id:        defaultId,
			key:       "not_exist_key",
			expectErr: ErrKeyNotExists,
		},
	}

	for _, tc := range testCases {
		var actualData map[string]string

		err = s.storage.getData(tc.id, tc.key, &actualData)

		s.Assert().Equal(tc.expectErr, err)
		s.Assert().Equal(tc.expectData, actualData)
	}
}

func (s *redisStorageTestSuite) Test_delData() {
	var (
		defaultId  int64 = 1
		defaultKey       = "foobar"
	)

	if err := s.redis.Set(s.ctx, s.storage.dataKey(defaultId, defaultKey), "hello world", 0).Err(); err != nil {
		s.T().Fatalf("setup test data (data) into redis err: %s", err)
	}

	// set other user data
	if err := s.redis.Set(s.ctx, s.storage.dataKey(2, "other_user"), "hello world", 0).Err(); err != nil {
		s.T().Fatalf("setup test data (data) into redis err: %s", err)
	}

	testCases := []struct {
		testName  string
		id        int64
		key       string
		expectErr error
	}{
		{
			testName:  "correct test",
			id:        defaultId,
			key:       defaultKey,
			expectErr: nil,
		},
		{
			testName:  "id not exist",
			id:        999,
			key:       defaultKey,
			expectErr: nil,
		},
		{
			testName:  "key not exist",
			id:        defaultId,
			key:       "key_not_exist",
			expectErr: nil,
		},
		{
			testName:  "id and key not exist",
			id:        999,
			key:       "key_not_exist",
			expectErr: nil,
		},
	}

	for _, tc := range testCases {
		err := s.storage.delData(tc.id, tc.key)
		s.Assert().Equal(tc.expectErr, err)

		if tc.expectErr == nil {
			var exist int64
			exist, err = s.redis.Exists(s.ctx, s.storage.dataKey(tc.id, tc.key)).Result()
			s.Assert().Nil(err)

			s.Assert().Zero(exist)

			// also check other user data. It must not be deleted
			exist, err = s.redis.Exists(s.ctx, s.storage.dataKey(2, "other_user")).Result()
			s.Assert().Nil(err)

			s.Assert().NotZero(exist)
		}
	}
}

func (s *redisStorageTestSuite) Test_clear() {
	var (
		defaultId  int64 = 1
		defaultKey       = "foobar"
	)

	if err := s.redis.Set(s.ctx, s.storage.dataKey(defaultId, defaultKey), "hello world", 0).Err(); err != nil {
		s.T().Fatalf("setup test data (data) into redis err: %s", err)
	}
	if err := s.redis.Set(s.ctx, s.storage.stateKey(defaultId), "someState", 0).Err(); err != nil {
		s.T().Fatalf("setup test data (state) into redis err: %s", err)
	}

	// set other user data
	if err := s.redis.Set(s.ctx, s.storage.dataKey(2, "other_user"), "hello world", 0).Err(); err != nil {
		s.T().Fatalf("setup test data (data) into redis err: %s", err)
	}
	if err := s.redis.Set(s.ctx, s.storage.stateKey(2), "someState", 0).Err(); err != nil {
		s.T().Fatalf("setup test data (state) into redis err: %s", err)
	}

	testCases := []struct {
		testName  string
		id        int64
		key       string
		expectErr error
	}{
		{
			testName:  "correct test",
			id:        defaultId,
			key:       defaultKey,
			expectErr: nil,
		},
		{
			testName:  "id not exist",
			id:        999,
			key:       defaultKey,
			expectErr: nil,
		},
		{
			testName:  "key not exist",
			id:        defaultId,
			key:       "key_not_exist",
			expectErr: nil,
		},
		{
			testName:  "id and key not exist",
			id:        999,
			key:       "key_not_exist",
			expectErr: nil,
		},
	}

	for _, tc := range testCases {
		err := s.storage.clear(tc.id)
		s.Assert().Equal(tc.expectErr, err)

		if tc.expectErr == nil {
			var exist int64
			exist, err = s.redis.Exists(s.ctx, s.storage.dataKey(tc.id, tc.key)).Result()
			s.Assert().Nil(err)

			s.Assert().Zero(exist)

			exist, err = s.redis.Exists(s.ctx, s.storage.stateKey(tc.id)).Result()
			s.Assert().Nil(err)

			s.Assert().Zero(exist)

			// also check other user data
			exist, err = s.redis.Exists(s.ctx, s.storage.dataKey(2, "other_user")).Result()
			s.Assert().Nil(err)

			s.Assert().NotZero(exist)

			exist, err = s.redis.Exists(s.ctx, s.storage.stateKey(2)).Result()
			s.Assert().Nil(err)

			s.Assert().NotZero(exist)

		}
	}
}
