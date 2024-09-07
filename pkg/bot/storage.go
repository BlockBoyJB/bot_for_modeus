package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

var (
	ErrKeyNotExists = errors.New("specified key does not exists")
)

// Реализация aiogram fsm (машины состояний)
// для удобства состояния хранятся в строках, а все данные - в виде мапы, где ключ - строка для удобства поиска,
// а значение - массив байтов для гибкого хранения разных структур
type storage interface {
	setState(id int64, state string) error
	getState(id int64) (string, error)
	setData(id int64, key string, v any) error
	setTempData(id int64, key string, v any, d time.Duration) error
	getData(id int64, key string, v any) error
	clear(id int64) error
}

// Чтобы был для удобства
type memoryStorage struct {
	sync.RWMutex
	state map[int64]string
	data  map[int64]map[string][]byte
}

func newMemoryStorage() *memoryStorage {
	return &memoryStorage{
		state: map[int64]string{},
		data:  make(map[int64]map[string][]byte),
	}
}

func (s *memoryStorage) setState(id int64, state string) error {
	s.Lock()
	defer s.Unlock()
	s.state[id] = state
	return nil
}

func (s *memoryStorage) getState(id int64) (string, error) {
	s.RLock()
	defer s.RUnlock()
	state, ok := s.state[id]
	if !ok {
		return "", ErrKeyNotExists
	}
	return state, nil
}

func (s *memoryStorage) __setData(id int64, data map[string][]byte) {
	s.Lock()
	defer s.Unlock()
	s.data[id] = data
}

func (s *memoryStorage) __getData(id int64) map[string][]byte {
	s.RLock()
	defer s.RUnlock()
	data, ok := s.data[id]
	if !ok {
		return map[string][]byte{}
	}
	return data
}

func (s *memoryStorage) setData(id int64, key string, v any) error {
	data := s.__getData(id)
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data[key] = b
	s.__setData(id, data)
	return nil
}

func (s *memoryStorage) setTempData(id int64, key string, v any, d time.Duration) error {
	data := s.__getData(id)
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data[key] = b
	s.__setData(id, data)
	go func() {
		time.Sleep(d)
		data := s.__getData(id)
		delete(data, key)
		s.__setData(id, data)
	}()
	return nil
}

func (s *memoryStorage) getData(id int64, key string, v any) error {
	data := s.__getData(id)
	b, ok := data[key]
	if !ok {
		return ErrKeyNotExists
	}
	return json.Unmarshal(b, v)
}

func (s *memoryStorage) clear(id int64) error {
	s.Lock()
	defer s.Unlock()
	delete(s.state, id)
	delete(s.data, id)
	return nil
}

// Ключи, по которым сохраняются данные пользователя (привет aiogram)
const (
	defaultStateKey = "fsm:%d:state" // fsm:userId:type (state/data)
	defaultDataKey  = "fsm:%d:data"
)

type redisStorage struct {
	*redis.Client
	ctx context.Context
}

func newRedisStorage(ctx context.Context, redis *redis.Client) *redisStorage {
	if ctx == nil {
		ctx = context.Background()
	}
	return &redisStorage{
		Client: redis,
		ctx:    ctx,
	}
}

func (s *redisStorage) setState(id int64, state string) error {
	return s.Set(s.ctx, s.stateKey(id), state, 0).Err()
}

func (s *redisStorage) getState(id int64) (string, error) {
	return s.Get(s.ctx, s.stateKey(id)).Result()
}

func (s *redisStorage) __setData(id int64, data map[string][]byte, d time.Duration) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.Set(s.ctx, s.dataKey(id), b, d).Err()
}

func (s *redisStorage) __getData(id int64) (map[string][]byte, error) {
	ok, err := s.Exists(s.ctx, s.dataKey(id)).Result()
	if err != nil {
		return nil, err
	}
	if ok == 0 {
		return make(map[string][]byte), nil
	}

	var data map[string][]byte
	b, err := s.Get(s.ctx, s.dataKey(id)).Bytes()
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *redisStorage) setData(id int64, key string, v any) error {
	data, err := s.__getData(id)
	if err != nil {
		return err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data[key] = b
	return s.__setData(id, data, 0)
}

func (s *redisStorage) setTempData(id int64, key string, v any, d time.Duration) error {
	data, err := s.__getData(id)
	if err != nil {
		return err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data[key] = b
	return s.__setData(id, data, d)
}

func (s *redisStorage) getData(id int64, key string, v any) error {
	data, err := s.__getData(id)
	if err != nil {
		return err
	}
	b, ok := data[key]
	if !ok {
		return ErrKeyNotExists
	}
	if err = json.Unmarshal(b, v); err != nil {
		return err
	}
	return nil
}

func (s *redisStorage) clear(id int64) error {
	if err := s.Del(s.ctx, s.stateKey(id)).Err(); err != nil {
		return err
	}
	if err := s.Del(s.ctx, s.dataKey(id)).Err(); err != nil {
		return err
	}
	return nil
}

func (s *redisStorage) stateKey(id int64) string {
	return fmt.Sprintf(defaultStateKey, id)
}

func (s *redisStorage) dataKey(id int64) string {
	return fmt.Sprintf(defaultDataKey, id)
}
