package modeus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	academicPeriodsUri      = "/students-app/api/pages/student-card/my/attendance-rates"
	studentRatingsUri       = "/students-app/api/pages/student-card/my/ratings"
	courseUnitsUri          = "/students-app/api/pages/student-card/my/academic-period-results-table/primary"
	coursersTotalResultsUri = "/students-app/api/pages/student-card/my/academic-period-results-table/secondary"
	eventResultsUri         = "/schedule-calendar-v2/api/calendar/events/%s/results/search"
)

type Lesson struct {
	Id                          string `json:"id"`
	LessonRealizationTemplateId string `json:"lessonRealizationTemplateId"`
	Name                        string `json:"name"`
	OrderIndex                  int    `json:"orderIndex"`
	LessonType                  string `json:"lessonType"`
	TypeName                    string `json:"typeName"`
	EventFormat                 string `json:"eventFormat"`
	EventHoldingStatus          string `json:"eventHoldingStatus"`
	EventStartsAtLocal          string `json:"eventStartsAtLocal"`
	EventEndsAtLocal            string `json:"eventEndsAtLocal"`
	LearningPathElementActive   bool   `json:"learningPathElementActive"`
}

type CourseUnit struct {
	Id            string   `json:"id"` // Для поиска итога модуля
	CourseUnitId  string   `json:"courseUnitId"`
	Name          string   `json:"name"`
	Lessons       []Lesson `json:"lessons"`
	TrainingTeams []struct {
		Id                 string `json:"id"`
		CycleRealizationId string `json:"cycleRealizationId"`
	} `json:"trainingTeams"`
	LearningPathElementActive bool `json:"learningPathElementActive"`
	MidCheckModule            bool `json:"midCheckModule"` // TODO если false, то предмет не оцениваемый?
}

type PrimaryGradesResponse struct {
	//AcademicCourses []struct { // По сути не нужен, потому что это названия "папок" для courseUnitRealizations
	//	Id                       string   `json:"id"`
	//	Name                     string   `json:"name"`
	//	CourseUnitRealizationIds []string `json:"courseUnitRealizationIds"`
	//} `json:"academicCourses"`
	CourseUnitRealizations []CourseUnit  `json:"courseUnitRealizations"`
	Errors                 []interface{} `json:"errors"`
}

type PrimaryGradesRequest struct {
	PersonId                    string `json:"personId"`
	WithMidcheckModulesIncluded bool   `json:"withMidcheckModulesIncluded"`
	AprId                       string `json:"aprId"`
	AcademicPeriodStartDate     string `json:"academicPeriodStartDate"`
	AcademicPeriodEndDate       string `json:"academicPeriodEndDate"`
	StudentId                   string `json:"studentId"`
	CurriculumFlowId            string `json:"curriculumFlowId"`
	CurriculumPlanId            string `json:"curriculumPlanId"`
}

type ResultCurrent struct {
	Id              string `json:"id"`
	ControlObjectId string `json:"controlObjectId"`
	ResultValue     string `json:"resultValue"`
	CreatedTs       string `json:"createdTs"`
	CreatedBy       string `json:"createdBy"`
	UpdatedTs       string `json:"updatedTs"`
	UpdatedBy       string `json:"updatedBy"`
}

type ResultFinal struct {
	CourseUnitRealizationId string `json:"courseUnitRealizationId"`
	ControlObjectId         string `json:"controlObjectId"`
	ResultValue             string `json:"resultValue"`
	UpdatedTs               string `json:"updatedTs"`
	UpdatedBy               string `json:"updatedBy"`
}

type EventPersonAttendance struct { // Отметки посещения за пару
	ResultId                    string      `json:"resultId"`
	EventId                     string      `json:"eventId"`
	LessonId                    interface{} `json:"lessonId"`
	LessonRealizationTemplateId string      `json:"lessonRealizationTemplateId"`
	CreatedTs                   string      `json:"createdTs"`
	CreatedBy                   string      `json:"createdBy"`
	UpdatedTs                   string      `json:"updatedTs"`
	UpdatedBy                   string      `json:"updatedBy"`
}

type SecondaryGradesResponse struct {
	Errors                               []interface{} `json:"errors"`
	CourseUnitRealizationAttendanceRates []struct {
		CourseUnitRealizationId string  `json:"courseUnitRealizationId"`
		PresentRate             float64 `json:"presentRate"`
		AbsentRate              float64 `json:"absentRate"`
		UndefinedRate           float64 `json:"undefinedRate"`
	} `json:"courseUnitRealizationAttendanceRates"`
	EventPersonAttendances       []EventPersonAttendance `json:"eventPersonAttendances"`
	AcademicCourseControlObjects []struct {              // Это баллы "папки"
		AcademicCourseId string `json:"academicCourseId"`
		Value            string `json:"value"`
	} `json:"academicCourseControlObjects"`
	CourseUnitRealizationControlObjects []struct {
		ControlObjectId         string        `json:"controlObjectId"`
		TypeName                string        `json:"typeName"`
		TypeShortName           interface{}   `json:"typeShortName"`
		TypeCode                string        `json:"typeCode"`
		OrderIndex              int           `json:"orderIndex"`
		CourseUnitRealizationId string        `json:"courseUnitRealizationId"`
		MainGradingScaleCode    string        `json:"mainGradingScaleCode"`
		ResultCurrent           ResultCurrent `json:"resultCurrent"`
		ResultFinal             ResultFinal   `json:"resultFinal"`
	} `json:"courseUnitRealizationControlObjects"`
	LessonControlObjects []struct {
		ControlObjectId      string      `json:"controlObjectId"`
		TypeName             string      `json:"typeName"`
		TypeShortName        interface{} `json:"typeShortName"`
		TypeCode             string      `json:"typeCode"`
		OrderIndex           int         `json:"orderIndex"`
		LessonId             string      `json:"lessonId"`
		MainGradingScaleCode string      `json:"mainGradingScaleCode"`
		Result               struct {
			Id              string `json:"id"`
			ControlObjectId string `json:"controlObjectId"`
			ResultValue     string `json:"resultValue"` // оценка за отдельную пару
			CreatedTs       string `json:"createdTs"`
			CreatedBy       string `json:"createdBy"`
			UpdatedTs       string `json:"updatedTs"`
			UpdatedBy       string `json:"updatedBy"`
		} `json:"result"`
		ResultRequired bool `json:"resultRequired"`
	} `json:"lessonControlObjects"`
}

type SecondaryGradesRequest struct {
	CourseUnitRealizationId []string `json:"courseUnitRealizationId"` // тут указываем CourseUnit.Id
	AcademicCourseId        []string `json:"academicCourseId"`        // TODO поэкспериментировать
	LessonId                []string `json:"lessonId"`                // Для детального просмотра оценок (посещаемость + баллы за каждую пару)
	PersonId                string   `json:"personId"`
	AprId                   string   `json:"aprId"`
	AcademicPeriodStartDate string   `json:"academicPeriodStartDate"`
	AcademicPeriodEndDate   string   `json:"academicPeriodEndDate"`
	StudentId               string   `json:"studentId"`
}

type APRealization struct {
	Id                        string `json:"id"`
	AcademicPeriodRealization struct {
		Id               string `json:"id"`
		Name             string `json:"name"`
		FullName         string `json:"fullName"`
		CurriculumId     string `json:"curriculumId"`
		CurriculumPlanId string `json:"curriculumPlanId"`
		CurriculumFlowId string `json:"curriculumFlowId"`
		StartYear        int    `json:"startYear"`
		StartDate        string `json:"startDate"` // PrimaryGradesRequest.AcademicPeriodStartDate
		EndDate          string `json:"endDate"`   // PrimaryGradesRequest.AcademicPeriodEndDate
		Number           int    `json:"number"`
		NumberInYear     int    `json:"numberInYear"`
		YearNumber       int    `json:"yearNumber"`
		Type             string `json:"type"`
		PlanningPeriodId string `json:"planningPeriodId"`
	} `json:"academicPeriodRealization"`
	StudentId     string  `json:"studentId"`
	PersonId      string  `json:"personId"` // TODO Приходит null
	PresentRate   float64 `json:"presentRate"`
	AbsentRate    float64 `json:"absentRate"`
	UndefinedRate float64 `json:"undefinedRate"`
}

type RatingsRequest struct {
	StudentId string   `json:"studentId"`
	AprId     []string `json:"aprId"`
}

type RatingsResponse struct {
	CgpaRating struct {
		Score                 float64 `json:"score"`
		ByCurriculumFlow      int     `json:"byCurriculumFlow"`
		TotalByCurriculumFlow int     `json:"totalByCurriculumFlow"`
	} `json:"cgpaRating"`
	GpaRatings []struct {
		AprId                 string  `json:"aprId"`
		Score                 float64 `json:"score"`
		ByCurriculumFlow      int     `json:"byCurriculumFlow"`
		TotalByCurriculumFlow int     `json:"totalByCurriculumFlow"`
	} `json:"gpaRatings"`
	Errors []interface{} `json:"errors"`
}

// FindAPR находит все прошедшие семестры + текущий
func (s *modeus) FindAPR(token, gradesId string) ([]APRealization, error) {
	type request struct {
		StudentId string `json:"studentId"`
	}
	req := request{StudentId: gradesId}
	resp, err := s.makeRequest(token, http.MethodPost, academicPeriodsUri, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response []APRealization

	if err = json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response, nil
}

func (s *modeus) FindCurrentAPR(token, gradesId string) (APRealization, error) {
	response, err := s.FindAPR(token, gradesId)
	if err != nil {
		return APRealization{}, err
	}
	maxNum := 0
	ind := 0
	for i, apr := range response {
		if apr.AcademicPeriodRealization.Number > maxNum {
			ind = i
			maxNum = apr.AcademicPeriodRealization.Number
		}
	}
	return response[ind], nil
}

func (s *modeus) StudentRatings(token string, input RatingsRequest) (RatingsResponse, error) {
	resp, err := s.makeRequest(token, http.MethodPost, studentRatingsUri, input)
	if err != nil {
		return RatingsResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return RatingsResponse{}, err
	}
	var response RatingsResponse

	if err = json.Unmarshal(body, &response); err != nil {
		return RatingsResponse{}, err
	}
	return response, nil
}

// TODO При детальном просмотре предмета обратить внимание на пары, в которых может быть несколько полей для оценок

// CourseUnits возвращает PrimaryGradesResponse.CourseUnitRealizations (id, название предмета + все запланированные встречи). На вход нужна информация о семестре и schedule + grades id пользователя (
func (s *modeus) CourseUnits(token string, input PrimaryGradesRequest) ([]CourseUnit, error) {
	resp, err := s.makeRequest(token, http.MethodPost, courseUnitsUri, input)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response PrimaryGradesResponse

	if err = json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.CourseUnitRealizations, nil
}

// CoursesTotalResults возвращает информацию о посещаемости (по id предметов), текущий итог баллов + итог семестра (если есть), оценки за встречи.
// На вход надо указать id всех предметов, id всех встреч (для детальной информации баллов по встречам), информацию о текущем семестре
func (s *modeus) CoursesTotalResults(token string, input SecondaryGradesRequest) (SecondaryGradesResponse, error) {
	resp, err := s.makeRequest(token, http.MethodPost, coursersTotalResultsUri, input)
	if err != nil {
		return SecondaryGradesResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SecondaryGradesResponse{}, err
	}
	var response SecondaryGradesResponse

	if err = json.Unmarshal(body, &response); err != nil {
		return SecondaryGradesResponse{}, err
	}
	return response, nil
}

type EventGradesRequest struct {
	StudentId               []string `json:"studentId"`
	EventTypeId             string   `json:"eventTypeId"` // Тот же, что и в ScheduleResponse.Embedded.Events[i].TypeId
	LessonId                string   `json:"lessonId"`
	LessonTemplateId        string   `json:"lessonTemplateId"`
	CourseUnitRealizationId string   `json:"courseUnitRealizationId"`
}

type Attendance struct {
	ResultId           string      `json:"resultId"`
	UpdatedByPersonId  string      `json:"updatedByPersonId"`
	UpdatedByFullName  string      `json:"updatedByFullName"`
	UpdatedByShortName interface{} `json:"updatedByShortName"`
	UpdatedByTs        string      `json:"updatedByTs"`
}

type Result struct {
	Id                 string      `json:"id"`
	Value              string      `json:"value"`
	Required           bool        `json:"required"`
	ControlObjectId    string      `json:"controlObjectId"`
	Name               string      `json:"name"`
	UpdatedByPersonId  string      `json:"updatedByPersonId"`
	UpdatedByFullName  string      `json:"updatedByFullName"`
	UpdatedByShortName interface{} `json:"updatedByShortName"`
	UpdatedByTs        string      `json:"updatedByTs"`
}

type EventGradesResponse struct {
	Embedded struct {
		AttendanceResult []Attendance `json:"attendance-result"`
		Results          []Result     `json:"results"`
		//MidCheckResults []struct {
		//	Id                 string      `json:"id"`
		//	Value              string      `json:"value"`
		//	Required           interface{} `json:"required"`
		//	ControlObjectId    interface{} `json:"controlObjectId"`
		//	Name               string      `json:"name"`
		//	UpdatedByPersonId  string      `json:"updatedByPersonId"`
		//	UpdatedByFullName  string      `json:"updatedByFullName"`
		//	UpdatedByShortName interface{} `json:"updatedByShortName"`
		//	UpdatedByTs        string      `json:"updatedByTs"`
		//} `json:"mid-check-results"`
	} `json:"_embedded"`
}

func (s *modeus) EventGrades(token, eventId string, input EventGradesRequest) (EventGradesResponse, error) {
	resp, err := s.makeRequest(token, http.MethodPost, fmt.Sprintf(eventResultsUri, eventId), input)
	if err != nil {
		return EventGradesResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return EventGradesResponse{}, err
	}
	var response EventGradesResponse

	if err = json.Unmarshal(body, &response); err != nil {
		return EventGradesResponse{}, err
	}
	return response, nil
}
