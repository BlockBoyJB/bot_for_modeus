package parser

import (
	"bot_for_modeus/pkg/modeus"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type SubjectGrades struct {
	Subject          string
	CurrentResult    string
	CourseUnitResult string
	PresentRate      string
	AbsentRate       string
	UndefinedRate    string
}

// TODO можно сейвить текущий aprId и периодически его обновлять???

type GradesInput struct {
	GradesId   string
	ScheduleId string
	Login      string
	Password   string
}

func (s *Service) CourseTotalGrades(ctx context.Context, input GradesInput) ([]SubjectGrades, error) {
	token, err := s.userToken(ctx, input.Login, input.Password)
	if err != nil {
		return nil, err
	}
	apr, err := s.modeus.FindCurrentAPR(token, input.GradesId)
	if err != nil {
		log.Errorf("%s/CourseTotalGrades error find currect academic period realization: %s", serviceParserPrefixLog, err)
		return nil, err
	}
	courseUnits, err := s.modeus.CourseUnits(token, modeus.PrimaryGradesRequest{
		PersonId:                    input.ScheduleId,
		WithMidcheckModulesIncluded: false,
		AprId:                       apr.AcademicPeriodRealization.Id,
		AcademicPeriodStartDate:     apr.AcademicPeriodRealization.StartDate,
		AcademicPeriodEndDate:       apr.AcademicPeriodRealization.EndDate,
		StudentId:                   input.GradesId,
		CurriculumFlowId:            apr.AcademicPeriodRealization.CurriculumFlowId,
		CurriculumPlanId:            apr.AcademicPeriodRealization.CurriculumPlanId,
	})
	if err != nil {
		log.Errorf("%s/CourseTotalGrades error find courses from current APR: %s", serviceParserPrefixLog, err)
		return nil, err
	}
	userSubjects := make(map[string]*SubjectGrades)
	var reqSubjects []string

	for _, course := range courseUnits {
		_, ok := userSubjects[course.Id]
		if !ok {
			userSubjects[course.Id] = &SubjectGrades{}
		}
		userSubjects[course.Id].Subject = course.Name
		reqSubjects = append(reqSubjects, course.Id)
	}

	aprResults, err := s.modeus.CoursesTotalResults(token, modeus.SecondaryGradesRequest{
		CourseUnitRealizationId: reqSubjects,
		AcademicCourseId:        []string{},
		LessonId:                []string{},
		PersonId:                input.ScheduleId,
		AprId:                   apr.AcademicPeriodRealization.Id,
		AcademicPeriodStartDate: apr.AcademicPeriodRealization.StartDate,
		AcademicPeriodEndDate:   apr.AcademicPeriodRealization.EndDate,
		StudentId:               input.GradesId,
	})
	if err != nil {
		log.Errorf("%s/CourseTotalGrades error find total results: %s", serviceParserPrefixLog, err)
		return nil, err
	}
	var result []SubjectGrades

	for _, attendanceRate := range aprResults.CourseUnitRealizationAttendanceRates {
		userSubjects[attendanceRate.CourseUnitRealizationId].AbsentRate = convertToPresent(attendanceRate.AbsentRate)
		userSubjects[attendanceRate.CourseUnitRealizationId].PresentRate = convertToPresent(attendanceRate.PresentRate)
		userSubjects[attendanceRate.CourseUnitRealizationId].UndefinedRate = convertToPresent(attendanceRate.UndefinedRate)
	}
	for _, totalResult := range aprResults.CourseUnitRealizationControlObjects {
		if totalResult.TypeCode == "CURRENT_COURSE_UNIT_RESULT" {
			result = append(result, SubjectGrades{
				Subject:          userSubjects[totalResult.CourseUnitRealizationId].Subject,
				CurrentResult:    s.parseCourseCurrentResult(totalResult.ResultCurrent),
				CourseUnitResult: s.parseCourseFinalResult(totalResult.ResultFinal),
				PresentRate:      userSubjects[totalResult.CourseUnitRealizationId].PresentRate,
				AbsentRate:       userSubjects[totalResult.CourseUnitRealizationId].AbsentRate,
				UndefinedRate:    userSubjects[totalResult.CourseUnitRealizationId].UndefinedRate,
			})
		}
	}
	return result, nil
}

func (s *Service) parseCourseCurrentResult(result modeus.ResultCurrent) string {
	if result.Id == "" {
		return "Неизвестно"
	}
	return result.ResultValue
}

func (s *Service) parseCourseFinalResult(result modeus.ResultFinal) string {
	if result.ControlObjectId == "" {
		return "Неизвестно"
	}
	return result.ResultValue
}

type AcademicPeriod struct {
	Name          string
	PresentRate   string
	AbsentRate    string
	UndefinedRate string
	GPA           string
}

func (s *Service) UserRatings(ctx context.Context, input GradesInput) (string, map[int]*AcademicPeriod, error) {
	token, err := s.userToken(ctx, input.Login, input.Password)
	if err != nil {
		return "", nil, err
	}
	APRs, err := s.modeus.FindAPR(token, input.GradesId)
	if err != nil {
		log.Errorf("%s/UserRatings error find all user academic period realizations: %s", serviceParserPrefixLog, err)
		return "", nil, err
	}
	var aprId []string
	for _, apr := range APRs {
		aprId = append(aprId, apr.AcademicPeriodRealization.Id)
	}
	ratings, err := s.modeus.StudentRatings(token, modeus.RatingsRequest{
		StudentId: input.GradesId,
		AprId:     aprId,
	})
	if err != nil {
		log.Errorf("%s/UserRatings error request student ratings: %s", serviceParserPrefixLog, err)
		return "", nil, err
	}
	result := make(map[int]*AcademicPeriod)
	for _, apr := range APRs {
		p := &AcademicPeriod{
			Name:          apr.AcademicPeriodRealization.FullName,
			PresentRate:   convertToPresent(apr.PresentRate),
			AbsentRate:    convertToPresent(apr.AbsentRate),
			UndefinedRate: convertToPresent(apr.UndefinedRate),
			GPA:           "Не подсчитан",
		}
		for _, gpa := range ratings.GpaRatings {
			if gpa.AprId == apr.AcademicPeriodRealization.Id {
				p.GPA = fmt.Sprintf("%.2f", gpa.Score)
				break
			}
		}
		result[apr.AcademicPeriodRealization.Number] = p
	}
	return fmt.Sprintf("%.2f", ratings.CgpaRating.Score), result, nil
}

func convertToPresent(d float64) string {
	return fmt.Sprintf("%d", int(d*100)) + "%"
}
