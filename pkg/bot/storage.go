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
	delData(id int64, keys ...string) error
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

func (s *memoryStorage) delData(id int64, keys ...string) error {
	data := s.__getData(id)
	for _, k := range keys {
		delete(data, k)
	}
	s.__setData(id, data)
	return nil
}

func (s *memoryStorage) clear(id int64) error {
	s.Lock()
	defer s.Unlock()
	delete(s.state, id)
	delete(s.data, id)
	return nil
}

const (
	redisStoragePrefix = "fsm:%d:%s"
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
	state, err := s.Get(s.ctx, s.stateKey(id)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrKeyNotExists
		}
		return "", err
	}
	return state, nil
}

func (s *redisStorage) setData(id int64, key string, v any) error {
	return s.setTempData(id, key, v, 0)
}

func (s *redisStorage) setTempData(id int64, key string, v any, d time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.Set(s.ctx, s.dataKey(id, key), b, d).Err()
}

func (s *redisStorage) getData(id int64, key string, v any) error {
	b, err := s.Get(s.ctx, s.dataKey(id, key)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrKeyNotExists
		}
		return err
	}
	return json.Unmarshal(b, v)
}

// Удаляет данные по заданному ключу. Не работает с паттернами по типу fsm:id:*. Нужно точное соответствие.
// Не возвращает ошибку redis.Nil
func (s *redisStorage) delData(id int64, keys ...string) error {
	var f []string
	for _, k := range keys {
		f = append(f, s.dataKey(id, k))
	}
	return s.Del(s.ctx, f...).Err()
}

// Работает достаточно медленно, потому что стираются все пользовательские ключи, поэтому преимущественно использовать delData.
// Не возвращает ошибку redis.Nil
func (s *redisStorage) clear(id int64) error {
	var (
		cursor uint64
		keys   []string
	)

	for {
		k, nc, err := s.Scan(s.ctx, cursor, s.dataKey(id, "*"), 20).Result()
		if err != nil {
			return err
		}

		keys = append(keys, k...)
		cursor = nc

		if cursor == 0 {
			break
		}
	}

	for _, k := range keys {
		err := s.Del(s.ctx, k).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *redisStorage) stateKey(id int64) string {
	return s.dataKey(id, "state")
}

func (s *redisStorage) dataKey(id int64, key string) string {
	return fmt.Sprintf(redisStoragePrefix, id, key)
}
