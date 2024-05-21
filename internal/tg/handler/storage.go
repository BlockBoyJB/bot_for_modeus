package handler

import "encoding/json"

// storage is simple implementation of aiogram (python tg framework) storage
type storage struct {
	state map[int64]string
	data  map[int64]map[string][]byte
}

func newStorage() *storage {
	return &storage{
		state: make(map[int64]string),
		data:  make(map[int64]map[string][]byte),
	}
}

func (s *storage) setState(id int64, state string) {
	s.state[id] = state
}

func (s *storage) getState(id int64) string {
	return s.state[id]
}

func _setData(s *storage, id int64, data map[string][]byte) {
	s.data[id] = data
}

func (s *storage) updateData(id int64, key string, value any) error {
	currentData := s.data[id]
	if currentData == nil {
		currentData = make(map[string][]byte)
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	currentData[key] = b
	_setData(s, id, currentData)
	return nil
}

func (s *storage) deleteData(id int64, key string) {
	currentData := s.data[id]
	if currentData == nil {
		return // nothing to delete
	}
	delete(currentData, key)
}

func (s *storage) getData(id int64, key string) []byte {
	return s.data[id][key]
}

func (s *storage) _getData(id int64, key string, v interface{}) error {
	data := s.data[id][key]
	return json.Unmarshal(data, v)
}

func (s *storage) clear(id int64) {
	delete(s.state, id)
	delete(s.data, id)
}
