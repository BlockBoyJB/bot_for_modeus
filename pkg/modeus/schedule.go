package modeus

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const (
	defaultScheduleUri = "/schedule-calendar-v2/api/calendar/events/search"
)

type Event struct {
	Name          string      `json:"name"`        // Название пары
	NameShort     string      `json:"nameShort"`   // Чаще всего тут тип занятия (например Лекционное занятие)
	Description   interface{} `json:"description"` // ???
	TypeId        string      `json:"typeId"`      // Тип занятия (лекция, практика, лабораторная и тд)
	FormatId      string      `json:"formatId"`
	Start         time.Time   `json:"start"`         // Начало занятия
	End           time.Time   `json:"end"`           // Конец занятия
	StartsAtLocal string      `json:"startsAtLocal"` // Начало по местному (тюменскому)
	EndsAtLocal   string      `json:"endsAtLocal"`   // Конец по местному
	StartsAt      string      `json:"startsAt"`      // Непонятно зачем дублирование
	EndsAt        string      `json:"endsAt"`        // Непонятно зачем дублирование
	HoldingStatus struct {
		Id                  string      `json:"id"`   // DRAFT не проведено, HELD проведено
		Name                string      `json:"name"` // Статус
		AudModifiedAt       interface{} `json:"audModifiedAt"`
		AudModifiedBy       interface{} `json:"audModifiedBy"`
		AudModifiedBySystem interface{} `json:"audModifiedBySystem"`
	} `json:"holdingStatus"`
	RepeatedLessonRealization interface{} `json:"repeatedLessonRealization"`
	UserRoleIds               []string    `json:"userRoleIds"`
	LessonTemplateId          string      `json:"lessonTemplateId"`
	Version                   int         `json:"__version"`
	Links                     struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Type struct {
			Href string `json:"href"`
		} `json:"type"`
		Format struct {
			Href string `json:"href"`
		} `json:"format"`
		TimeZone struct {
			Href string `json:"href"`
		} `json:"time-zone"`
		Grid struct {
			Href string `json:"href"`
		} `json:"grid"`
		CourseUnitRealization struct {
			Href string `json:"href"` // Совпадает с CourseUnitRealizations Id
		} `json:"course-unit-realization"`
		CycleRealization struct {
			Href string `json:"href"`
		} `json:"cycle-realization"`
		LessonRealization struct {
			Href string `json:"href"`
		} `json:"lesson-realization"`
		LessonRealizationTeam struct {
			Href string `json:"href"`
		} `json:"lesson-realization-team"`
		LessonRealizationTemplate struct {
			Href string `json:"href"`
		} `json:"lesson-realization-template"`
		Location struct {
			Href string `json:"href"`
		} `json:"location"`
		Duration struct {
			Href string `json:"href"`
		} `json:"duration"`
		Team struct {
			Href string `json:"href"`
		} `json:"team"`
		Organizers struct {
			Href string `json:"href"`
		} `json:"organizers"`
		HoldingStatusModifiedBy struct {
			Href string `json:"href"`
		} `json:"holding-status-modified-by,omitempty"`
	} `json:"_links"`
	Id string `json:"id"`
}

type CourseUnitRealization struct {
	Name        string `json:"name"`
	NameShort   string `json:"nameShort"`
	PrototypeId string `json:"prototypeId"`
	Links       struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		PlanningPeriod struct {
			Href string `json:"href"`
		} `json:"planning-period"`
	} `json:"_links"`
	Id string `json:"id"`
}

type CycleRealization struct {
	Name                           string `json:"name"`
	NameShort                      string `json:"nameShort"`
	Code                           string `json:"code"`
	CourseUnitRealizationNameShort string `json:"courseUnitRealizationNameShort"`
	Links                          struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		CourseUnitRealization struct {
			Href string `json:"href"`
		} `json:"course-unit-realization"`
	} `json:"_links"`
	Id string `json:"id"`
}

type EventRoom struct {
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Event struct {
			Href string `json:"href"`
		} `json:"event"`
		Room struct {
			Href string `json:"href"`
		} `json:"room"`
	} `json:"_links"`
	Id string `json:"id"`
}

type Room struct {
	Name      string `json:"name"`
	NameShort string `json:"nameShort"`
	Building  struct {
		Id           string `json:"id"`
		Name         string `json:"name"`
		NameShort    string `json:"nameShort"`
		Address      string `json:"address"`
		DisplayOrder int    `json:"displayOrder"`
	} `json:"building"`
	ProjectorAvailable interface{} `json:"projectorAvailable"`
	TotalCapacity      int         `json:"totalCapacity"`
	WorkingCapacity    int         `json:"workingCapacity"`
	DeletedAtUtc       interface{} `json:"deletedAtUtc"`
	Links              struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Type struct {
			Href string `json:"href"`
		} `json:"type"`
		Building struct {
			Href string `json:"href"`
		} `json:"building"`
	} `json:"_links"`
	Id string `json:"id"`
}

type ScheduleResponse struct { // Как хорошо что у goland есть возможность автоматически парсить json в структуру...
	Embedded struct {
		Events                 []Event                 `json:"events"`
		CourseUnitRealizations []CourseUnitRealization `json:"course-unit-realizations"`
		//CycleRealizations      []CycleRealization      `json:"cycle-realizations"`
		//LessonRealizationTeams []struct {
		//	Name  string `json:"name"`
		//	Links struct {
		//		Self struct {
		//			Href string `json:"href"`
		//		} `json:"self"`
		//	} `json:"_links"`
		//	Id string `json:"id"`
		//} `json:"lesson-realization-teams"`
		LessonRealizations []struct {
			Name        string `json:"name"`
			NameShort   string `json:"nameShort"`
			PrototypeId string `json:"prototypeId"`
			Ordinal     int    `json:"ordinal"`
			Links       struct {
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"_links"`
			Id string `json:"id"`
		} `json:"lesson-realizations"`
		EventLocations []struct {
			EventId        string `json:"eventId"`
			CustomLocation string `json:"customLocation"`
			Links          struct {
				Self []struct {
					Href string `json:"href"`
				} `json:"self"`
				EventRooms struct {
					Href string `json:"href"`
				} `json:"event-rooms,omitempty"`
			} `json:"_links"`
		} `json:"event-locations"`
		//Durations []struct {
		//	EventId    string `json:"eventId"`
		//	Value      int    `json:"value"`
		//	TimeUnitId string `json:"timeUnitId"`
		//	Minutes    int    `json:"minutes"`
		//	Links      struct {
		//		Self []struct {
		//			Href string `json:"href"`
		//		} `json:"self"`
		//		TimeUnit struct {
		//			Href string `json:"href"`
		//		} `json:"time-unit"`
		//	} `json:"_links"`
		//} `json:"durations"`
		EventRooms []EventRoom `json:"event-rooms"`
		Rooms      []Room      `json:"rooms"`
		//Buildings  []struct {
		//	Name              string `json:"name"`
		//	NameShort         string `json:"nameShort"`
		//	Address           string `json:"address"`
		//	SearchableAddress string `json:"searchableAddress"`
		//	DisplayOrder      int    `json:"displayOrder"`
		//	Links             struct {
		//		Self struct {
		//			Href string `json:"href"`
		//		} `json:"self"`
		//	} `json:"_links"`
		//	Id string `json:"id"`
		//} `json:"buildings"`
		//EventTeams []struct {
		//	EventId string `json:"eventId"`
		//	Size    int    `json:"size"`
		//	Links   struct {
		//		Self struct {
		//			Href string `json:"href"`
		//		} `json:"self"`
		//		Event struct {
		//			Href string `json:"href"`
		//		} `json:"event"`
		//	} `json:"_links"`
		//} `json:"event-teams"`
		//EventOrganizers []struct {
		//	EventId string `json:"eventId"`
		//	Links   struct {
		//		Self struct {
		//			Href string `json:"href"`
		//		} `json:"self"`
		//		Event struct {
		//			Href string `json:"href"`
		//		} `json:"event"`
		//		EventAttendees interface{} `json:"event-attendees"`
		//	} `json:"_links"`
		//} `json:"event-organizers"`
		EventAttendees []struct {
			RoleId           string `json:"roleId"`
			RoleName         string `json:"roleName"`
			RoleNamePlural   string `json:"roleNamePlural"`
			RoleDisplayOrder int    `json:"roleDisplayOrder"`
			Links            struct {
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
				Event struct {
					Href string `json:"href"`
				} `json:"event"`
				Person struct {
					Href string `json:"href"`
				} `json:"person"`
			} `json:"_links"`
			Id string `json:"id"`
		} `json:"event-attendees"`
		Persons []struct {
			LastName   string `json:"lastName"`
			FirstName  string `json:"firstName"`
			MiddleName string `json:"middleName"`
			FullName   string `json:"fullName"`
			Links      struct {
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"_links"`
			Id string `json:"id"`
		} `json:"persons"`
	} `json:"_embedded"`
	//Page struct {
	//	Size          int `json:"size"`
	//	TotalElements int `json:"totalElements"`
	//	TotalPages    int `json:"totalPages"`
	//	Number        int `json:"number"`
	//} `json:"page"`
}

type ScheduleRequest struct {
	Size             int       `json:"size"`
	TimeMin          time.Time `json:"timeMin"`
	TimeMax          time.Time `json:"timeMax"`
	AttendeePersonId []string  `json:"attendeePersonId"`
}

func (s *modeus) Schedule(token string, input ScheduleRequest) (ScheduleResponse, error) {
	resp, err := s.makeRequest(token, http.MethodPost, defaultScheduleUri, input)
	if err != nil {
		return ScheduleResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ScheduleResponse{}, err
	}
	var response ScheduleResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return ScheduleResponse{}, err
	}
	return response, nil
}
