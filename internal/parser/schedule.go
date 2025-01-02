package parser

import (
	"bot_for_modeus/pkg/modeus"
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type Lesson struct {
	Name          string    // Название пары
	Subject       string    // Предмет
	Type          string    // Тип занятия
	Time          string    // Время проведения
	AuditoriumNum string    // Номер аудитории
	BuildingAddr  string    // Адрес корпуса
	Lector        string    // ФИО преподавателя
	start         time.Time // Вспомогательное поле для сортировки
}

func (p *parser) DaySchedule(ctx context.Context, scheduleId string, now time.Time) ([]Lesson, error) {
	input := modeus.ScheduleRequest{
		Size:             500,
		TimeMin:          time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		TimeMax:          time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()),
		AttendeePersonId: []string{scheduleId},
	}
	return p.parseSchedule(ctx, input)
}

// Обновил расписание на неделю. Теперь вместо 6 запросов делается 1.

func (p *parser) WeekSchedule(ctx context.Context, scheduleId string, now time.Time) (map[int][]Lesson, error) {
	start := now.Day() - int(now.Weekday()) + 1
	input := modeus.ScheduleRequest{
		Size:             500,
		TimeMin:          time.Date(now.Year(), now.Month(), start, 0, 0, 0, 0, now.Location()),
		TimeMax:          time.Date(now.Year(), now.Month(), start+6, 0, 0, 0, 0, now.Location()),
		AttendeePersonId: []string{scheduleId},
	}
	schedule, err := p.parseSchedule(ctx, input)
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
		key := int(l.start.Weekday())
		result[key] = append(result[key], l) // можем так делать, потому что занятия идут друг за другом по времени (порядок не будет нарушен)
	}
	return result, nil
}

// Теперь сортирует итоговый слайс. Была проблема (в DaySchedule), что иногда расписание приходит не в отсортированном порядке.
func (p *parser) parseSchedule(ctx context.Context, input modeus.ScheduleRequest) ([]Lesson, error) {
	token, err := p.rootToken(ctx)
	if err != nil {
		return nil, err
	}
	schedule, err := p.modeus.Schedule(token, input)
	if err != nil {
		var e *modeus.ErrModeusUnavailable
		if errors.As(err, &e) {
			log.Errorf("%s/parseSchedule modeus error: %s", parserServicePrefixLog, e)
			return nil, ErrModeusUnavailable
		}
		log.Errorf("%s/parseSchedule error find user schedule from modeus: %s", parserServicePrefixLog, err)
		return nil, err
	}

	var result []Lesson
	for _, event := range schedule.Embedded.Events {
		auditoriumNum, buildingAddr := parseLessonLocation(event, schedule)

		result = append(result, Lesson{
			Name:          event.Name,
			Subject:       parseLessonSubject(event, schedule),
			Type:          parseLessonType(event),
			Time:          parseLessonTime(event),
			AuditoriumNum: auditoriumNum,
			BuildingAddr:  buildingAddr,
			Lector:        parseLessonLector(event, schedule),
			start:         event.Start,
		})
	}
	return sort(result), nil
}

func parseLessonLocation(event modeus.Event, response modeus.ScheduleResponse) (auditoriumNum string, buildingAddr string) {
	eventId := event.Links.Self.Href
	var roomId string
	for _, eventRoom := range response.Embedded.EventRooms {
		if eventRoom.Links.Event.Href == eventId {
			roomId = eventRoom.Links.Room.Href
			break
		}
	}
	if len(roomId) == 0 {
		// надо проверить, что это не ошибка, а кастомная локация
		for _, l := range response.Embedded.EventLocations {
			if l.EventId == eventId[1:] {
				if l.CustomLocation == "" {
					return "Ошибка", "Ошибка"
				}
				return l.CustomLocation, "💻 Онлайн" // TODO онлайн или неизвестно?
			}
		}
		return "Ошибка", "Ошибка"
	}
	for _, room := range response.Embedded.Rooms {
		if room.Links.Self.Href == roomId {
			return room.Name, room.Building.Address
		}
	}
	return "Ошибка", "Ошибка"
}

func parseLessonSubject(event modeus.Event, response modeus.ScheduleResponse) string {
	courseId := event.Links.CourseUnitRealization.Href
	for _, courseUnit := range response.Embedded.CourseUnitRealizations {
		if courseUnit.Links.Self.Href == courseId {
			return courseUnit.Name
		}
	}
	return "Ошибка"
}

func parseLessonType(event any) string {
	var (
		t  string
		ok bool
	)
	switch event.(type) {
	case modeus.Event:
		e := event.(modeus.Event)
		t, ok = lessonTypes[e.TypeId]
	case modeus.Lesson:
		e := event.(modeus.Lesson)
		t, ok = lessonTypes[e.LessonType]
	}
	if !ok {
		return "Ошибка"
	}
	return t
}

func parseLessonTime(event any) string {
	var (
		start, end time.Time
	)
	switch event.(type) {
	case modeus.Event:
		e := event.(modeus.Event)
		start, end = e.Start, e.End
	case modeus.Lesson:
		e := event.(modeus.Lesson)
		start, _ = time.Parse("2006-01-02T15:04:05", e.EventStartsAtLocal)
		end, _ = time.Parse("2006-01-02T15:04:05", e.EventEndsAtLocal)
	}
	return fmt.Sprintf("%s - %s", start.Format("15:04"), end.Format("15:04"))
}

// Находит всех преподавателей для занятия (их может быть несколько, особенно на повторных аттестациях)
// Результатом функции будут ФИО преподавателей через запятую (ФИО, ФИО, ...)
func parseLessonLector(event modeus.Event, response modeus.ScheduleResponse) string {
	eventId := event.Links.Self.Href
	exist := make(map[string]bool) // мапа, в которой ключ будет id преподавателя (значение bool для удобства)
	for _, l := range response.Embedded.EventAttendees {
		if l.Links.Event.Href == eventId {
			exist[l.Links.Person.Href] = true
		}
	}
	if len(exist) == 0 {
		return "Неизвестно"
	}

	result := make([]string, 0, len(exist))

	for _, person := range response.Embedded.Persons {
		if exist[person.Links.Self.Href] {
			result = append(result, person.FullName)

			if len(result) == len(exist) { // чтобы не делать лишние итерации в цикле
				break
			}
		}
	}
	if len(result) == 0 {
		return "Ошибка"
	}
	return strings.Join(result, ", ")
}

// Для сортировки расписания выбрал merge_sort, потому что данные очень часто приходят абсолютно не в том порядке
// Поэтому сортировка за O(n^2) точно не подходит.
// Был выбран именно этот метод, потому что он стабильно быстро сортирует за O(n*log(n))
// (в отличие от quick_sort, где в худшем случае будет O(n^2))
func sort(l []Lesson) []Lesson {
	if len(l) < 2 {
		return l
	}
	first := sort(l[:len(l)/2])
	second := sort(l[len(l)/2:])
	return merge(first, second)
}

func merge(first, second []Lesson) []Lesson {
	var (
		result []Lesson
		i, j   int
	)
	for i < len(first) && j < len(second) {
		if first[i].start.Before(second[j].start) {
			result = append(result, first[i])
			i++
		} else {
			result = append(result, second[j])
			j++
		}
	}

	for ; i < len(first); i++ {
		result = append(result, first[i])
	}
	for ; j < len(second); j++ {
		result = append(result, second[j])
	}
	return result
}
