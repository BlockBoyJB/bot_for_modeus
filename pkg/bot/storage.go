package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

// Ключи, по которым сохраняются данные пользователя (привет aiogram)
const (
	defaultStateKey = "fsm:%d:state" // fsm:userId:type (state/data)
	defaultDataKey  = "fsm:%d:data"
)

const (
	storagePrefixLog = "bot/storage"
)

// Реализация aiogram redis fsm (машины состояний)
// для удобства состояния хранятся в строках, а все данные - в виде мапы, где ключ - строка для удобства поиска,
// а значение - массив байтов для гибкого хранения разных структур
type storage struct {
	*redis.Client
}

func newStorage(redis *redis.Client) *storage {
	return &storage{redis}
}

func (s *storage) SetState(ctx context.Context, id int64, state string) error {
	if err := s.Set(ctx, stateKey(id), state, 0).Err(); err != nil {
		log.Errorf("%s/SetState error set state: %s", storagePrefixLog, err)
		return err
	}
	return nil
}

func (s *storage) GetState(ctx context.Context, id int64) (string, error) {
	state, err := s.Get(ctx, stateKey(id)).Result()
	if err != nil {
		log.Errorf("%s/GetState error get state: %s", storagePrefixLog, err)
		return "", err
	}
	return state, nil
}

func (s *storage) _setData(ctx context.Context, id int64, data map[string][]byte) error {
	b, err := json.Marshal(data)
	if err != nil {
		log.Errorf("%s/_setData error marshal data: %s", storagePrefixLog, err)
		return err
	}
	if err = s.Set(ctx, dataKey(id), b, 0).Err(); err != nil {
		log.Errorf("%s/_setData error set data: %s", storagePrefixLog, err)
		return err
	}
	return nil
}

func (s *storage) _getData(ctx context.Context, id int64) (map[string][]byte, error) {
	var data map[string][]byte
	ok, err := s.Exists(ctx, dataKey(id)).Result()
	if err != nil {
		log.Errorf("%s/_getData error check exist data: %s", storagePrefixLog, err)
		return nil, err
	}
	if ok == 0 {
		return make(map[string][]byte), nil
	}
	b, err := s.Get(ctx, dataKey(id)).Bytes()
	if err != nil {
		log.Errorf("%s/_getData error get data: %s", storagePrefixLog, err)
		return nil, err
	}
	if err = json.Unmarshal(b, &data); err != nil {
		log.Errorf("%s/_getData error unmarshal data: %s", storagePrefixLog, err)
		return nil, err
	}
	return data, nil
}

func (s *storage) GetData(ctx context.Context, id int64, key string, v interface{}) error {
	data, err := s._getData(ctx, id)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data[key], v); err != nil {
		log.Errorf("%s/GetData error unmarshal data: %s", storagePrefixLog, err)
		return err
	}
	return nil
}

func (s *storage) UpdateData(ctx context.Context, id int64, key string, value interface{}) error {
	data, err := s._getData(ctx, id)
	if err != nil {
		return err
	}
	b, err := json.Marshal(value)
	if err != nil {
		log.Errorf("%s/UpdateData error marshal data: %s", storagePrefixLog, err)
		return err
	}
	data[key] = b
	return s._setData(ctx, id, data)
}

func (s *storage) Clear(ctx context.Context, id int64) error {
	if err := s.Del(ctx, stateKey(id)).Err(); err != nil {
		log.Errorf("%s/Clear error clear state: %s", storagePrefixLog, err)
		return err
	}
	if err := s.Del(ctx, dataKey(id)).Err(); err != nil {
		log.Errorf("%s/Clear error clear data: %s", storagePrefixLog, err)
		return err
	}
	return nil
}

func stateKey(id int64) string {
	return fmt.Sprintf(defaultStateKey, id)
}

func dataKey(id int64) string {
	return fmt.Sprintf(defaultDataKey, id)
}
