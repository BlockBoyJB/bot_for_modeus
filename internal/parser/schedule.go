package parser

import (
	"net/http"
	"time"
)

const (
	findScheduleUri = "/schedule"
)

type Lesson struct {
	Name          string    `json:"name"`           // Название пары
	Subject       string    `json:"subject"`        // Предмет
	Type          string    `json:"type"`           // Тип занятия
	Time          string    `json:"time"`           // Время проведения
	AuditoriumNum string    `json:"auditorium_num"` // Номер аудитории
	BuildingAddr  string    `json:"building_addr"`  // Адрес корпуса
	Lector        string    `json:"lector"`         // ФИО преподавателя
	Start         time.Time `json:"start"`          // Вспомогательное поле для сортировки. Сейчас необходимо для разделения расписания по дням
}

type scheduleRequest struct {
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	ScheduleId string    `json:"schedule_id"`
}

func (p *parser) DaySchedule(scheduleId string, now time.Time) ([]Lesson, error) {
	return p.parseSchedule(scheduleRequest{ // TODO fix timezone
		Start:      time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		End:        time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()),
		ScheduleId: scheduleId,
	})
}

func (p *parser) WeekSchedule(scheduleId string, now time.Time) (map[int][]Lesson, error) {
	start := now.Day() - int(now.Weekday()) + 1
	input := scheduleRequest{ // TODO fix timezone
		Start:      time.Date(now.Year(), now.Month(), start, 0, 0, 0, 0, now.Location()),
		End:        time.Date(now.Year(), now.Month(), start+6, 0, 0, 0, 0, now.Location()),
		ScheduleId: scheduleId,
	}
	schedule, err := p.parseSchedule(input)
	if err != nil {
		return nil, err
	}
	result := make(map[int][]Lesson, 6)

	if len(schedule) == 0 {
		return result, nil
	}

	// Идем циклом по всему расписанию (оно отсортированное).
	// Заполняем мапу с недельным расписанием, вычисляя ключ для каждого занятия через time.Time{}.Weekday() (Понедельник начинается с 1)
	for _, l := range schedule {
		key := int(l.Start.Weekday())
		result[key] = append(result[key], l) // можем так делать, потому что занятия идут друг за другом по времени (порядок не будет нарушен)
	}
	return result, nil
}

func (p *parser) parseSchedule(input scheduleRequest) ([]Lesson, error) {
	resp, err := p.makeRequest(http.MethodPost, findScheduleUri, input)
	if err != nil {
		return nil, err
	}

	var result []Lesson
	if err = parseBody(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}
