package parser

import (
	"bot_for_modeus/pkg/modeus"
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

type Lesson struct {
	Name          string
	Subject       string
	Type          string
	Time          string
	AuditoriumNum string
	BuildingAddr  string
	Lector        string
}

func (s *Service) DaySchedule(ctx context.Context, scheduleId string) ([]Lesson, error) {
	token, err := s.rootToken(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	input := modeus.ScheduleRequest{
		Size:             500, // 14 05
		TimeMin:          time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		TimeMax:          time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()),
		AttendeePersonId: []string{scheduleId},
	}
	return s.parseSchedule(token, input)
}

func (s *Service) WeekSchedule(ctx context.Context, scheduleId string) (map[int][]Lesson, error) {
	token, err := s.rootToken(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	start := now.Day() - int(now.Weekday()) + 1
	result := make(map[int][]Lesson)
	// TODO поменять формат поиска расписания
	// На данный момент такой вариант, потому что я не понял каким образом отфильтровать расписание по дням (я тупой)
	for i := start; i <= start+6; i++ {
		currDay := time.Date(now.Year(), now.Month(), i, 0, 0, 0, 0, now.Location())
		input := modeus.ScheduleRequest{
			Size:             500,
			TimeMin:          currDay,
			TimeMax:          time.Date(now.Year(), now.Month(), i+1, 0, 0, 0, 0, now.Location()),
			AttendeePersonId: []string{scheduleId},
		}
		lessons, _ := s.parseSchedule(token, input)
		result[int(currDay.Weekday())] = lessons
	}
	return result, nil
}

func (s *Service) parseSchedule(token string, input modeus.ScheduleRequest) ([]Lesson, error) {
	schedule, err := s.modeus.Schedule(token, input)
	if err != nil {
		log.Errorf("%s/parseSchedule error find user schedule from modeus: %s", serviceParserPrefixLog, err)
		return nil, err
	}
	var result []Lesson
	for _, event := range schedule.Embedded.Events {
		auditoriumNum, buildingAddr := "Неизвестно", "Неизвестно"
		room, err := s.parseLessonLocation(event, schedule)
		if err == nil {
			auditoriumNum = room.Name
			buildingAddr = room.Building.Address
		}
		lesson := Lesson{
			Name:          event.Name,
			Subject:       s.parseLessonSubject(event, schedule),
			Type:          s.parseLessonType(event),
			Time:          s.parseLessonTime(event),
			AuditoriumNum: auditoriumNum,
			BuildingAddr:  buildingAddr,
			Lector:        s.parseLessonLector(event, schedule),
		}
		result = append(result, lesson)
	}
	return result, nil
}

func (s *Service) parseLessonTime(event modeus.Event) string {
	return fmt.Sprintf("%s - %s", event.Start.Format("15:04"), event.End.Format("15:04"))
}

func (s *Service) parseLessonType(event modeus.Event) string {
	t, ok := lessonTypes[event.TypeId]
	if !ok {
		return "Неизвестно"
	}
	return t
}

func (s *Service) parseLessonLocation(event modeus.Event, response modeus.ScheduleResponse) (modeus.Room, error) {
	eventId := event.Links.Self.Href
	var roomId string
	for _, eventRoom := range response.Embedded.EventRooms {
		if eventRoom.Links.Event.Href == eventId {
			roomId = eventRoom.Links.Room.Href
			break
		}
	}
	if len(roomId) == 0 {
		return modeus.Room{}, errors.New("cannot find event room")
	}
	for _, room := range response.Embedded.Rooms {
		if room.Links.Self.Href == roomId {
			return room, nil
		}
	}
	return modeus.Room{}, errors.New("cannot find event room")
}

func (s *Service) parseLessonLector(event modeus.Event, response modeus.ScheduleResponse) string {
	eventId := event.Links.Self.Href
	var attendeeId string
	for _, eventAttendee := range response.Embedded.EventAttendees {
		if eventAttendee.Links.Event.Href == eventId {
			attendeeId = eventAttendee.Links.Person.Href
			break
		}
	}
	if len(attendeeId) == 0 {
		return "Неизвестно"
	}
	for _, person := range response.Embedded.Persons {
		if person.Links.Self.Href == attendeeId {
			return person.FullName
		}
	}
	return "Неизвестно"
}

func (s *Service) parseLessonSubject(event modeus.Event, response modeus.ScheduleResponse) string {
	courseId := event.Links.CourseUnitRealization.Href
	for _, courseUnit := range response.Embedded.CourseUnitRealizations {
		if courseUnit.Links.Self.Href == courseId {
			return courseUnit.Name
		}
	}
	return "Неизвестно"
}
