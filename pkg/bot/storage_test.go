package bot

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
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
				"foo":     "bar",
				"message": "hello world",
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

		var actual, expect any

		expectData, err := sonic.Marshal(tc.data)
		s.Assert().Nil(err)
		s.Assert().Nil(sonic.Unmarshal(expectData, &expect))

		s.Assert().Nil(sonic.Unmarshal(b, &actual))
		s.Assert().NotNil(actual)

		s.Assert().Equal(expect, actual)
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
				2: "world",
				1: "hello",
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

func (s *redisStorageTestSuite) Test_setCommonData() {
	testCases := []struct {
		testName string
		key      string
		data     any
		ttl      time.Duration
	}{
		{
			testName: "test string",
			key:      "foobar",
			data:     "hello_world",
			ttl:      time.Minute,
		},
		{
			testName: "test int",
			data:     123,
			key:      "abc",
			ttl:      time.Hour,
		},
		{
			testName: "test slice",
			key:      "foo",
			data:     []int{1, 2, 10, 3},
			ttl:      time.Minute * 10,
		},
		{
			testName: "test map",
			key:      "bar",
			data: map[int]string{
				2: "world",
				1: "hello",
			},
			ttl: time.Hour * 10,
		},
		{
			testName: "test struct",
			key:      "some_struct",
			data: struct {
				Name string
				Age  int
			}{
				Name: "Petya",
				Age:  18,
			},
			ttl: time.Second * 50,
		},
	}

	for _, tc := range testCases {
		err := s.storage.setCommonData(tc.key, tc.data, tc.ttl)
		s.Assert().Nil(err)

		b, err := s.redis.Get(s.ctx, s.storage.normalizeKey(tc.key)).Bytes()
		s.Assert().Nil(err)

		var actual, expect any

		expectData, err := sonic.Marshal(tc.data)
		s.Assert().Nil(err)
		s.Assert().Nil(sonic.Unmarshal(expectData, &expect))

		s.Assert().Nil(sonic.Unmarshal(b, &actual))
		s.Assert().NotNil(actual)

		s.Assert().Equal(expect, actual)

		if tc.ttl > 0 {
			actualTTL, err := s.redis.TTL(s.ctx, s.storage.normalizeKey(tc.key)).Result()
			s.Assert().Nil(err)
			s.Assert().True(tc.ttl-actualTTL <= time.Second*2)
		}
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

	b, err := sonic.Marshal(defaultData)
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

func (s *redisStorageTestSuite) Test_getCommonData() {
	var (
		defaultKey  = "foobar"
		defaultData = map[string]string{
			"message": "hello world",
		}
	)

	b, err := sonic.Marshal(defaultData)
	if err != nil {
		s.T().Fatalf("marhsal test data err: %s", err)
	}

	if err = s.redis.Set(s.ctx, s.storage.normalizeKey(defaultKey), b, 0).Err(); err != nil {
		s.T().Fatalf("setup test data into redis err: %s", err)
	}

	testCases := []struct {
		testName   string
		key        string
		expectErr  error
		expectData map[string]string
	}{
		{
			testName:   "correct test",
			key:        defaultKey,
			expectErr:  nil,
			expectData: defaultData,
		},
		{
			testName:   "key not exist",
			key:        "not_exist_key",
			expectErr:  ErrKeyNotExists,
			expectData: nil,
		},
	}

	for _, tc := range testCases {
		var actualData map[string]string

		err = s.storage.getCommonData(tc.key, &actualData)

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

func (s *redisStorageTestSuite) Test_delCommonData() {
	var (
		defaultKey   = "foobar"
		otherDataKey = "other_data"
	)

	if err := s.redis.Set(s.ctx, s.storage.normalizeKey(defaultKey), "hello world", 0).Err(); err != nil {
		s.T().Fatalf("setup test data into redis err: %s", err)
	}

	if err := s.redis.Set(s.ctx, s.storage.normalizeKey(otherDataKey), "hello world 2", 0).Err(); err != nil {
		s.T().Fatalf("setup test data into redis err: %s", err)
	}

	testCases := []struct {
		testName string
		key      string
	}{
		{
			testName: "correct test",
			key:      defaultKey,
		},
		{
			testName: "key not exist",
			key:      "key_not_exist",
		},
	}

	for _, tc := range testCases {
		err := s.storage.delCommonData(tc.key)
		s.Assert().Nil(err)

		exist, err := s.redis.Exists(s.ctx, s.storage.normalizeKey(tc.key)).Result()
		s.Assert().Nil(err)

		s.Assert().Zero(exist)

		exist, err = s.redis.Exists(s.ctx, s.storage.normalizeKey(otherDataKey)).Result()
		s.Assert().Nil(err)
		s.Assert().NotZero(exist)
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

func Benchmark_dataKey(b *testing.B) {
	s := &redisStorage{}

	for i := 0; i < b.N; i++ {
		key := s.dataKey(10000000000000, "state")
		if key != "fsm:10000000000000:state" {
			b.Fatalf("not equal expect fsm:10000000000000:state, got %s", key)
		}
	}
}

// Старый вариант создания ключа для redis storage. (через fmt Sprintf)
func Benchmark_oldDataKey(b *testing.B) {
	f := func(id int64, key string) string {
		return fmt.Sprintf("fsm:%d:%s", id, key)
	}

	for i := 0; i < b.N; i++ {
		key := f(10000000000000, "state")
		if key != "fsm:10000000000000:state" {
			b.Fatalf("not equal expect fsm:10000000000000:state, got %s", key)
		}
	}
}
