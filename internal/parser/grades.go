package parser

import (
	"bot_for_modeus/pkg/modeus"
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type GradesInput struct {
	Login      string
	Password   string
	ScheduleId string
	GradesId   string
}

type SubjectGrades struct {
	Name           string // Название предмета
	Status         string // Декоративный элемент. Зависит от текущего количества баллов
	CurrentResult  string // Текущее количество баллов
	SemesterResult string // Итог семестра (оценка отлично, хорошо и тд)
	PresentRate    string // Процент посещений
	AbsentRate     string // Процент пропусков
	UndefinedRate  string // Процент без отметки о посещении
}

type Semester struct {
	Id               string
	Number           int
	StartDate        string
	EndDate          string
	CurriculumFlowId string
	CurriculumPlanId string
}

//func (p *parser) CurrentSemesterGrades(ctx context.Context, input GradesInput) ([]SubjectGrades, error) {
//	token, err := p.userToken(ctx, input.Login, input.Password)
//	if err != nil {
//		return nil, err
//	}
//
//	apr, err := p.modeus.FindCurrentAPR(token, input.GradesId)
//	if err != nil {
//		log.Errorf("%s/CurrentSemesterGrades error find current semester: %s", parserServicePrefixLog, err)
//		return nil, err
//	}
//	semester := Semester{
//		Id:               apr.AcademicPeriodRealization.Id,
//		Number:           apr.AcademicPeriodRealization.Number,
//		StartDate:        apr.AcademicPeriodRealization.StartDate,
//		EndDate:          apr.AcademicPeriodRealization.EndDate,
//		CurriculumFlowId: apr.AcademicPeriodRealization.CurriculumFlowId,
//		CurriculumPlanId: apr.AcademicPeriodRealization.CurriculumPlanId,
//	}
//	return p.SemesterTotalGrades(ctx, input, semester)
//
//}

func (p *parser) SemesterTotalGrades(ctx context.Context, input GradesInput, semester Semester) ([]SubjectGrades, error) {
	token, err := p.userToken(ctx, input.Login, input.Password)
	if err != nil {
		return nil, err
	}

	// Сначала нужно найти все предметы в семестре.
	// К тому же, чтобы получать оценки детально по проведенным парам, нужно из первого запроса забирать id пар...
	semesterSubjects, err := p.modeus.CourseUnits(token, modeus.PrimaryGradesRequest{
		PersonId:                    input.ScheduleId,
		WithMidcheckModulesIncluded: false,
		AprId:                       semester.Id,
		AcademicPeriodStartDate:     semester.StartDate,
		AcademicPeriodEndDate:       semester.EndDate,
		StudentId:                   input.GradesId,
		CurriculumFlowId:            semester.CurriculumFlowId,
		CurriculumPlanId:            semester.CurriculumPlanId,
	})
	if err != nil {
		log.Errorf("%s/SemesterTotalGrades error find semester subjects: %s", parserServicePrefixLog, err)
		return nil, err
	}

	type attendanceRates struct {
		subject       string
		presentRate   string
		absentRate    string
		undefinedRate string
	}
	tempResult := make(map[string]attendanceRates)

	// Для второго запроса необходимо предоставить id всех предметов в семестре
	var requestSubjects []string
	for _, subject := range semesterSubjects {
		// во втором запросе нет названия предметов, а в первом нет отметок посещения...
		tempResult[subject.Id] = attendanceRates{
			subject: subject.Name,
		}
		requestSubjects = append(requestSubjects, subject.Id)
	}

	semesterResults, err := p.modeus.CoursesTotalResults(token, modeus.SecondaryGradesRequest{
		CourseUnitRealizationId: requestSubjects,
		AcademicCourseId:        []string{},
		LessonId:                []string{},
		PersonId:                input.ScheduleId,
		AprId:                   semester.Id,
		AcademicPeriodStartDate: semester.StartDate,
		AcademicPeriodEndDate:   semester.EndDate,
		StudentId:               input.GradesId,
	})
	if err != nil {
		log.Errorf("%s/SemesterTotalGrades error find semester total results: %s", parserServicePrefixLog, err)
		return nil, err
	}

	var result []SubjectGrades

	for _, ar := range semesterResults.CourseUnitRealizationAttendanceRates {
		tempResult[ar.CourseUnitRealizationId] = attendanceRates{
			subject:       tempResult[ar.CourseUnitRealizationId].subject,
			presentRate:   convertToPresent(ar.PresentRate),
			absentRate:    convertToPresent(ar.AbsentRate),
			undefinedRate: convertToPresent(ar.UndefinedRate),
		}
	}

	for _, subjectResult := range semesterResults.CourseUnitRealizationControlObjects {
		if subjectResult.TypeCode == "CURRENT_COURSE_UNIT_RESULT" { // такой фильтр на оцениваемый предмет
			result = append(result, SubjectGrades{
				Name:           tempResult[subjectResult.CourseUnitRealizationId].subject,
				Status:         formatCourseCurrentResult(subjectResult.ResultCurrent),
				CurrentResult:  parseCourseCurrentResult(subjectResult.ResultCurrent),
				SemesterResult: parseCourseFinalResult(subjectResult.ResultFinal),
				PresentRate:    tempResult[subjectResult.CourseUnitRealizationId].presentRate,
				AbsentRate:     tempResult[subjectResult.CourseUnitRealizationId].absentRate,
				UndefinedRate:  tempResult[subjectResult.CourseUnitRealizationId].undefinedRate,
			})
		}
	}
	return result, nil
}

type SemesterResult struct {
	Name          string // Название семестра
	PresentRate   string // Процент посещений
	AbsentRate    string // Процент пропусков
	UndefinedRate string // Процент без отметок посещаемости
	GPA           string // gpa рейтинг
}

func (p *parser) Ratings(ctx context.Context, input GradesInput) (string, map[int]SemesterResult, error) {
	token, err := p.userToken(ctx, input.Login, input.Password)
	if err != nil {
		return "", nil, err
	}
	semesters, err := p.modeus.FindAPR(token, input.GradesId)
	if err != nil {
		log.Errorf("%s/Ratings error find all user semesters: %s", parserServicePrefixLog, err)
		return "", nil, err
	}

	// для получения рейтинга всех семестров на нужны их id
	var requestSem []string
	for _, s := range semesters {
		requestSem = append(requestSem, s.AcademicPeriodRealization.Id)
	}
	ratings, err := p.modeus.StudentRatings(token, modeus.RatingsRequest{
		StudentId: input.GradesId,
		AprId:     requestSem,
	})
	if err != nil {
		log.Errorf("%s/Ratings error find student ratings: %s", parserServicePrefixLog, err)
		return "", nil, err
	}

	result := make(map[int]SemesterResult)
	for _, s := range semesters {
		sr := SemesterResult{
			Name:          s.AcademicPeriodRealization.FullName,
			PresentRate:   convertToPresent(s.PresentRate),
			AbsentRate:    convertToPresent(s.AbsentRate),
			UndefinedRate: convertToPresent(s.UndefinedRate),
			GPA:           "Не подсчитан",
		}
		for _, gpa := range ratings.GpaRatings {
			if gpa.AprId == s.AcademicPeriodRealization.Id {
				sr.GPA = fmt.Sprintf("%.2f", gpa.Score)
				break
			}
		}
		result[s.AcademicPeriodRealization.Number] = sr
	}
	cgpa := fmt.Sprintf("%.2f", ratings.CgpaRating.Score)
	return cgpa, result, nil
}

type SubjectDayGrades struct {
	Subject    string // Название предмета
	Name       string // Название пары
	Type       string // Тип занятия
	Time       string // Время проведения
	Attendance string // Отметка посещения
	Grades     string // Баллы, поставленные за пару (может быть несколько)
	start      time.Time
}

func (p *parser) DayGrades(ctx context.Context, day time.Time, input GradesInput) ([]SubjectDayGrades, error) {
	token, err := p.userToken(ctx, input.Login, input.Password)
	if err != nil {
		return nil, err
	}

	schedule, err := p.modeus.Schedule(token, modeus.ScheduleRequest{
		Size:             500,
		TimeMin:          time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location()),
		TimeMax:          time.Date(day.Year(), day.Month(), day.Day()+1, 0, 0, 0, 0, day.Location()),
		AttendeePersonId: []string{input.ScheduleId},
	})
	if err != nil {
		log.Errorf("%s/DayGrades error find user schedule: %s", parserServicePrefixLog, err)
		return nil, err
	}

	var result []SubjectDayGrades
	for _, event := range schedule.Embedded.Events {
		grades, e := p.modeus.EventGrades(token, event.Id, modeus.EventGradesRequest{
			StudentId:               []string{input.GradesId},
			EventTypeId:             event.TypeId,
			LessonId:                event.Links.LessonRealization.Href[1:],
			LessonTemplateId:        event.LessonTemplateId,
			CourseUnitRealizationId: event.Links.CourseUnitRealization.Href[1:],
		})
		if e != nil {
			log.Errorf("%s/DayGrades error find lesson grades: %s", parserServicePrefixLog, e)
			return nil, e
		}

		result = append(result, SubjectDayGrades{
			Subject:    parseLessonSubject(event, schedule),
			Name:       event.Name,
			Type:       parseLessonType(event),
			Time:       parseLessonTime(event),
			Attendance: parseLessonAttendance(grades.Embedded.AttendanceResult),
			Grades:     parseLessonGrades(grades.Embedded.Results),
			start:      event.Start,
		})
	}
	return bubble(result), nil
}

func (p *parser) FindCurrentSemester(ctx context.Context, input GradesInput) (Semester, error) {
	token, err := p.userToken(ctx, input.Login, input.Password)
	if err != nil {
		return Semester{}, err
	}

	s, err := p.modeus.FindCurrentAPR(token, input.GradesId)
	if err != nil {
		log.Errorf("%s/FindCurrentSemester error find current user semester: %s", parserServicePrefixLog, err)
		return Semester{}, err
	}

	return Semester{
		Id:               s.AcademicPeriodRealization.Id,
		Number:           s.AcademicPeriodRealization.Number,
		StartDate:        s.AcademicPeriodRealization.StartDate,
		EndDate:          s.AcademicPeriodRealization.EndDate,
		CurriculumFlowId: s.AcademicPeriodRealization.CurriculumFlowId, // TODO различие с предыдущим вариантом (был CurriculumId)
		CurriculumPlanId: s.AcademicPeriodRealization.CurriculumPlanId,
	}, nil
}

func (p *parser) FindAllSemesters(ctx context.Context, input GradesInput) (map[string]Semester, error) {
	token, err := p.userToken(ctx, input.Login, input.Password)
	if err != nil {
		return nil, err
	}

	semesters, err := p.modeus.FindAPR(token, input.GradesId)
	if err != nil {
		log.Errorf("%s/FindAllSemesters erorr find user semesters: %s", parserServicePrefixLog, err)
		return nil, err
	}

	result := make(map[string]Semester)
	for _, s := range semesters {
		result[s.Id] = Semester{
			Id:               s.AcademicPeriodRealization.Id,
			Number:           s.AcademicPeriodRealization.Number,
			StartDate:        s.AcademicPeriodRealization.StartDate,
			EndDate:          s.AcademicPeriodRealization.EndDate,
			CurriculumFlowId: s.AcademicPeriodRealization.CurriculumFlowId, // TODO различие с предыдущим вариантом (был CurriculumId)
			CurriculumPlanId: s.AcademicPeriodRealization.CurriculumPlanId,
		}
	}
	return result, nil
}

// FindSemesterSubjects возвращает в формате subjectId - название
func (p *parser) FindSemesterSubjects(ctx context.Context, input GradesInput, semester Semester) (map[string]string, error) {
	token, err := p.userToken(ctx, input.Login, input.Password)
	if err != nil {
		return nil, err
	}

	semesterSubjects, err := p.modeus.CourseUnits(token, modeus.PrimaryGradesRequest{
		PersonId:                    input.ScheduleId,
		WithMidcheckModulesIncluded: false,
		AprId:                       semester.Id,
		AcademicPeriodStartDate:     semester.StartDate,
		AcademicPeriodEndDate:       semester.EndDate,
		StudentId:                   input.GradesId,
		CurriculumFlowId:            semester.CurriculumFlowId,
		CurriculumPlanId:            semester.CurriculumPlanId,
	})
	if err != nil {
		log.Errorf("%s/FindSemesterSubjects error find subjects from modeus: %s", parserServicePrefixLog, err)
		return nil, err
	}

	result := make(map[string]string)
	for _, s := range semesterSubjects {
		result[s.Id] = s.Name
	}
	return result, nil
}

type LessonGrades struct {
	Name       string // Название пары
	Type       string // Тип занятия
	Time       string // Время проведения
	Attendance string // Отметка посещения
	Grades     string // Оценки (может быть несколько)
}

// в первом запросе забираем lessonId (для второго запроса), название пары, начало-конец, тип
// можно сортировать по orderIndex (начинается с 0)
// формат мапа int - структура. В начале циклом от 0 до длины
// нам нужно ограничение максимального orderIndex статус проведен. Все что идут выше и не проведены удаляем

// в начале временная мапа lessonId - структура. Потом разворачиваем в формат int - структура и удаляем не проведенные пары

func (p *parser) SubjectDetailedInfo(ctx context.Context, input GradesInput, semester Semester, subjectId string) (map[int]LessonGrades, error) {
	token, err := p.userToken(ctx, input.Login, input.Password)
	if err != nil {
		return nil, err
	}

	semesterSubject, err := p.modeus.CourseUnits(token, modeus.PrimaryGradesRequest{
		PersonId:                    input.ScheduleId,
		WithMidcheckModulesIncluded: false,
		AprId:                       semester.Id,
		AcademicPeriodStartDate:     semester.StartDate,
		AcademicPeriodEndDate:       semester.EndDate,
		StudentId:                   input.GradesId,
		CurriculumFlowId:            semester.CurriculumFlowId,
		CurriculumPlanId:            semester.CurriculumPlanId,
	})
	if err != nil {
		log.Errorf("%s/SubjectDetailedInfo error find semester subjects: %s", parserServicePrefixLog, err)
		return nil, err
	}

	type tempLessonGrades struct {
		*LessonGrades
		orderIndex int
	}

	var subject modeus.CourseUnit
	for _, s := range semesterSubject {
		if s.Id == subjectId {
			subject = s
			break
		}
	}
	if subject.Id == "" {
		return nil, errors.New("cannot find subject with specified id")
	}

	// ключ - lesson.LessonRealizationTemplateId, значение - lesson.Id
	attendanceResults := make(map[string]string)

	tempResult := make(map[string]*tempLessonGrades)
	var requestLesson []string
	for _, l := range subject.Lessons {
		tempResult[l.Id] = &tempLessonGrades{
			LessonGrades: &LessonGrades{
				Name: l.Name,
				Type: parseLessonType(l),
				Time: parseLessonTime(l),
			},
			orderIndex: l.OrderIndex,
		}
		requestLesson = append(requestLesson, l.Id)

		attendanceResults[l.LessonRealizationTemplateId] = l.Id
	}

	semesterResults, err := p.modeus.CoursesTotalResults(token, modeus.SecondaryGradesRequest{
		CourseUnitRealizationId: []string{subjectId},
		AcademicCourseId:        []string{},
		LessonId:                requestLesson,
		PersonId:                input.ScheduleId,
		AprId:                   semester.Id,
		AcademicPeriodStartDate: semester.StartDate,
		AcademicPeriodEndDate:   semester.EndDate,
		StudentId:               input.GradesId,
	})
	if err != nil {
		log.Errorf("%s/SubjectDetailedInfo error find semester total results: %s", parserServicePrefixLog, err)
		return nil, err
	}
	for _, lesson := range semesterResults.LessonControlObjects {
		l := tempResult[lesson.LessonId]
		if l.Grades != "" {
			l.Grades += lesson.Result.ResultValue + " "
		} else {
			l.Grades = lesson.Result.ResultValue + " "
		}
	}

	// TODO баг не находит пару по lesson.EventId
	for _, lesson := range semesterResults.EventPersonAttendances {
		lessonId, ok := attendanceResults[lesson.LessonRealizationTemplateId]
		if !ok {
			continue
		}
		tempResult[lessonId].Attendance = parseLessonAttendance(lesson)
	}

	result := make(map[int]LessonGrades)
	for _, lesson := range tempResult {
		result[lesson.orderIndex] = LessonGrades{
			Name:       lesson.Name,
			Type:       lesson.Type,
			Time:       lesson.Time,
			Attendance: lesson.Attendance,
			Grades:     lesson.Grades,
		}
	}

	return result, nil
}

func formatCourseCurrentResult(result modeus.ResultCurrent) string {
	if result.Id == "" {
		return "❓"
	}
	v, err := strconv.ParseFloat(result.ResultValue, 64)
	if err != nil {
		return ""
	}
	if v < 61 {
		return "📕"
	}
	if v < 76 {
		return "📒"
	}
	if v < 91 {
		return "📗"
	}
	if v >= 91 {
		return "📘"
	}
	return "Ошибка"
}

func parseCourseCurrentResult(result modeus.ResultCurrent) string {
	if result.Id == "" {
		return "Неизвестно"
	}
	return result.ResultValue
}

func parseCourseFinalResult(result modeus.ResultFinal) string {
	if result.ControlObjectId == "" {
		return "Неизвестно"
	}
	return result.ResultValue
}

func convertToPresent(d float64) string {
	return fmt.Sprintf("%d", int(d*100)) + "%"
}

func parseLessonAttendance(attendances any) string {
	var result string
	switch attendances.(type) {
	case []modeus.Attendance:
		a := attendances.([]modeus.Attendance)
		if len(a) == 0 {
			return "❓"
		}
		result = a[0].ResultId
	case modeus.EventPersonAttendance:
		a := attendances.(modeus.EventPersonAttendance)
		result = a.ResultId
	default:
		return "❓"
	}
	if result == "PRESENT" {
		return "✅ П"
	} else if result == "ABSENT" {
		return "❌ Н"
	}
	return "❓"

}

func parseLessonGrades(results []modeus.Result) string {
	if len(results) == 0 {
		return "Не поставлены"
	}
	var result string
	for _, r := range results {
		if len(r.Value) == 0 {
			result += "Не поставлены "
			continue
		}
		result += r.Value + " "
	}
	return result
}

// Решил повеселиться и использовать разные варианты сортировки (в schedule.go есть merge_sort)
// Выбрал пузырек для сортировки расписания на день, потому что чаще всего расписание на день приходит почти упорядоченным,
// в то время как расписание на неделю всегда неупорядоченное. Поэтому возможный O(n) намного интереснее постоянного O(n*log(n))
// Хотя это все глупости, потому что за день пар явно меньше 10
func bubble(a []SubjectDayGrades) []SubjectDayGrades {
	swapped := true
	for swapped {
		swapped = false
		for i := 1; i < len(a); i++ {
			if a[i].start.Before(a[i-1].start) {
				a[i], a[i-1] = a[i-1], a[i]
				swapped = true
			}
		}
	}
	return a
}
