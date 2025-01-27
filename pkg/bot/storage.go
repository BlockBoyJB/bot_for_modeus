package bot

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
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
	setCommonData(key string, v any, d time.Duration) error
	getData(id int64, key string, v any) error
	getCommonData(key string, v any) error
	delData(id int64, keys ...string) error
	delCommonData(keys ...string) error
	clear(id int64) error
}

// Чтобы был для удобства
type memoryStorage struct {
	sync.RWMutex
	state  map[int64]string
	data   map[int64]map[string][]byte
	common map[string][]byte
}

func newMemoryStorage() *memoryStorage {
	return &memoryStorage{
		state:  map[int64]string{},
		data:   make(map[int64]map[string][]byte),
		common: make(map[string][]byte),
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

func (s *memoryStorage) setCommonData(key string, v any, d time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	s.Lock()
	defer s.Unlock()
	s.common[key] = b
	if d > 0 {
		go func() {
			time.Sleep(d)
			s.Lock()
			delete(s.common, key)
			s.Unlock()
		}()
	}
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

func (s *memoryStorage) getCommonData(key string, v any) error {
	s.RLock()
	defer s.RUnlock()
	b, ok := s.common[key]
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

func (s *memoryStorage) delCommonData(keys ...string) error {
	s.Lock()
	defer s.Unlock()
	for _, k := range keys {
		delete(s.common, k)
	}
	return nil
}

func (s *memoryStorage) clear(id int64) error {
	s.Lock()
	defer s.Unlock()
	delete(s.state, id)
	delete(s.data, id)
	return nil
}

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
	return s.setCommonData(s.dataKey(id, key), v, 0)
}

func (s *redisStorage) setTempData(id int64, key string, v any, d time.Duration) error {
	return s.setCommonData(s.dataKey(id, key), v, d)
}

func (s *redisStorage) setCommonData(key string, v any, d time.Duration) error {
	b, err := sonic.Marshal(v)
	if err != nil {
		return err
	}
	return s.Set(s.ctx, s.normalizeKey(key), b, d).Err()
}

func (s *redisStorage) getData(id int64, key string, v any) error {
	return s.getCommonData(s.dataKey(id, key), v)
}

func (s *redisStorage) getCommonData(key string, v any) error {
	b, err := s.Get(s.ctx, s.normalizeKey(key)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrKeyNotExists
		}
		return err
	}
	return sonic.UnmarshalString(b, v)
}

// Удаляет данные по заданному ключу. Не работает с паттернами по типу fsm:id:*. Нужно точное соответствие.
// Не возвращает ошибку redis.Nil
func (s *redisStorage) delData(id int64, keys ...string) error {
	for i, k := range keys {
		keys[i] = s.dataKey(id, k)
	}
	return s.Del(s.ctx, keys...).Err()
}

func (s *redisStorage) delCommonData(keys ...string) error {
	for i, k := range keys {
		keys[i] = s.normalizeKey(k)
	}
	return s.Del(s.ctx, keys...).Err()
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

	if len(keys) == 0 {
		return nil
	}
	return s.Del(s.ctx, keys...).Err()
}

func (s *redisStorage) stateKey(id int64) string {
	return s.dataKey(id, "state")
}

// Ключ в формате fsm:id:key
// Создаем буфер размером 24 + len(key) из которых 4 байта на префикс "fsm:", 19 на int64 (не 20, потому что только числа > 0) и еще 1 на символ ":" = 24
// Уменьшаем аллокации и стреляем в ногу - делаем строку через unsafe.
func (s *redisStorage) dataKey(id int64, key string) string {
	buf := make([]byte, 0, 24+len(key)) // fsm: (4) + int64 (19, т.к. id > 0, поэтому не включаем -) + : (1) = 24

	buf = append(buf, "fsm:"...)
	buf = strconv.AppendInt(buf, id, 10)
	buf = append(buf, ':')
	buf = append(buf, key...)

	return *(*string)(unsafe.Pointer(&buf))
}

func (s *redisStorage) normalizeKey(key string) string {
	if strings.HasPrefix(key, "fsm:") {
		return key
	}
	return "fsm:" + key
}
