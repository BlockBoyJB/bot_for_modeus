package parser

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	findSemesterGradesUri      = "/grades/total"
	findRatingsUri             = "/grades/ratings"
	findDayGradesUri           = "/grades/day"
	findSemestersUri           = "/grades/semesters"
	findSemesterSubjects       = "/grades/semesters/subjects"
	findSubjectDetailedInfoUri = "/grades/subjects"
)

type GradesInput struct {
	Login      string `json:"login"`
	Password   string `json:"password"`
	ScheduleId string `json:"schedule_id"`
	GradesId   string `json:"grades_id"`
}

type Semester struct {
	Id               string `json:"id"`
	Number           int    `json:"number"`
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date"`
	CurriculumFlowId string `json:"curriculum_flow_id"`
	CurriculumPlanId string `json:"curriculum_plan_id"`
}

type semesterTotalRequest struct {
	GradesInput
	Semester
}

type SubjectGrades struct {
	Name           string `json:"name"`            // Название предмета
	Status         string `json:"status"`          // Декоративный элемент. Зависит от текущего количества баллов
	CurrentResult  string `json:"current_result"`  // Текущее количество баллов
	SemesterResult string `json:"semester_result"` // Итог семестра (оценка отлично, хорошо и тд)
	PresentRate    string `json:"present_rate"`    // Процент посещений
	AbsentRate     string `json:"absent_rate"`     // Процент пропусков
	UndefinedRate  string `json:"undefined_rate"`  // Процент без отметки о посещении
}

func (p *parser) SemesterTotalGrades(gi GradesInput, semester Semester) ([]SubjectGrades, error) {
	resp, err := p.makeRequest(http.MethodPost, findSemesterGradesUri, semesterTotalRequest{
		GradesInput: gi,
		Semester:    semester,
	})
	if err != nil {
		return nil, err
	}

	var result []SubjectGrades
	if err = parseBody(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

type ratingsResponse struct {
	CGPA    string            `json:"cgpa"`
	Ratings []SemesterRatings `json:"ratings"`
}

type SemesterRatings struct {
	Name          string `json:"name"`           // Название семестра
	PresentRate   string `json:"present_rate"`   // Процент посещений
	AbsentRate    string `json:"absent_rate"`    // Процент пропусков
	UndefinedRate string `json:"undefined_rate"` // Процент без отметок посещаемости
	GPA           string `json:"gpa"`            // gpa рейтинг
}

func (p *parser) Ratings(gi GradesInput) (string, []SemesterRatings, error) {
	resp, err := p.makeRequest(http.MethodPost, findRatingsUri, gi)
	if err != nil {
		return "", nil, err
	}

	var result ratingsResponse
	if err = parseBody(resp, &result); err != nil {
		return "", nil, err
	}
	return result.CGPA, result.Ratings, nil
}

type dayGradesRequest struct {
	GradesInput
	Day time.Time `json:"day"`
}

type DayGrades struct {
	Subject    string `json:"subject"`    // Название предмета
	Name       string `json:"name"`       // Название пары
	Type       string `json:"type"`       // Тип занятия
	Time       string `json:"time"`       // Время проведения
	Attendance string `json:"attendance"` // Отметка посещения
	Grades     string `json:"grades"`     // Баллы, поставленные за пару (может быть несколько)
}

func (p *parser) DayGrades(day time.Time, gi GradesInput) ([]DayGrades, error) {
	resp, err := p.makeRequest(http.MethodPost, findDayGradesUri, dayGradesRequest{
		GradesInput: gi,
		Day:         day,
	})
	if err != nil {
		return nil, err
	}

	var result []DayGrades
	if err = parseBody(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *parser) FindAllSemesters(gi GradesInput) ([]Semester, error) {
	resp, err := p.makeRequest(http.MethodPost, findSemestersUri, gi)
	if err != nil {
		return nil, err
	}

	var semesters []Semester
	if err = parseBody(resp, &semesters); err != nil {
		return nil, err
	}
	return semesters, nil
}

func (p *parser) FindCurrentSemester(gi GradesInput) (Semester, error) {
	semesters, err := p.FindAllSemesters(gi)
	if err != nil {
		return Semester{}, err
	}

	if len(semesters) < 1 {
		return Semester{}, errors.New("semesters not found")
	}
	return semesters[len(semesters)-1], nil // в ответе они отсортированы по возрастанию
}

func (p *parser) FindSemesterSubjects(gi GradesInput, semester Semester) (map[string]string, error) {
	resp, err := p.makeRequest(http.MethodPost, findSemesterSubjects, semesterTotalRequest{
		GradesInput: gi,
		Semester:    semester,
	})
	if err != nil {
		return nil, err
	}

	var result map[string]string // TODO сменить на слайс структур (чтобы был ordered)?
	if err = parseBody(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

type LessonGrades struct {
	Name       string `json:"name"`       // Название пары
	Type       string `json:"type"`       // Тип занятия
	Time       string `json:"time"`       // Время проведения
	Attendance string `json:"attendance"` // Отметка посещения
	Grades     string `json:"grades"`     // Оценки (может быть несколько)
}

func (p *parser) SubjectDetailedInfo(gi GradesInput, semester Semester, subjectId string) ([]LessonGrades, error) {
	uri := fmt.Sprintf("%s?subject_id=%s", findSubjectDetailedInfoUri, subjectId)
	resp, err := p.makeRequest(http.MethodPost, uri, semesterTotalRequest{
		GradesInput: gi,
		Semester:    semester,
	})
	if err != nil {
		return nil, err
	}

	var result []LessonGrades
	if err = parseBody(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}
