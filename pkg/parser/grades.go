package parser

import (
	"fmt"
	"github.com/tebeka/selenium"
	"strings"
	"time"
)

// lesson-realization-wrap mds-event-type-lab mds-event-status-draft ng-star-inserted - не прошла пара
// lesson-realization-wrap mds-event-type-lab mds-event-status-held ng-star-inserted - прошла пара

// currentSubjectField
// первое поле %d: case(1 or 2) = 1 - названия предметов, 2 - колонка оценок;
// второе поле %d = конкретная строчка с предметом (разделом предметов)
// + /td[%d] для получения списка оценок in range(1, 37)
const (
	subjectNamesTable       = "//div[@class='ui-table-scrollable-wrapper ng-star-inserted']/div[1]/div[2]/table/tbody/tr" // for find all user subjects
	gradesLoadedInfoField   = "//div[@class='p-d-flex p-ai-center p-jc-between ng-star-inserted']"                        // Он отображается только когда все оценки загрузились
	currentSubjectField     = "//div[@class='ui-table-scrollable-wrapper ng-star-inserted']/div[%d]/div[2]/table/tbody/tr[%d]"
	currentResultField      = "/td[@class='current-result ng-star-inserted']"  // Текущее количество баллов
	finalResultField        = "/td[@class='final-result ng-star-inserted']"    // Итог модуля (оценка)
	currentAttendanceField  = "/td[@class='rates-present ng-star-inserted']"   // Текущий процент посещаемости (отметка "П" за пару)
	currentMissingField     = "/td[@class='rates-absebt ng-star-inserted']"    // Текущий процент пропусков (отметка "Н" за пару)
	currentWithoutMarkField = "/td[@class='rates-undefined ng-star-inserted']" // Текущий процент без отметок (преподаватель не поставил ничего)
	txtUserGrades           = "%d. Предмет: %s\nТекущий итог: %s\nИтог модуля: %s\nПосещаемость: %s\nПропуск: %s\nНе отмечено: %s"
	txtDetailedLessonInfo   = "Встреча %d\nТип встречи: %s\nОтметка посещаемости: %s\nПоставленные баллы: %s"

	//meetingStatusHeld    = "mds-event-status-held"  // Встреча проведена
	meetingStatusNotHeld = "mds-event-status-draft" // Встреча еще не проведена
)

// UserGrades возвращает карту с доступными для просмотра предметов (ключ - номер кнопки, значение - индекс предмета в модеусе)
func (s *Selenium) UserGrades(driver selenium.WebDriver, login, password string, timeout time.Duration) (map[int]int, string, error) {
	if err := s.setupGradesPage(driver, login, password, timeout); err != nil {
		return nil, "", err
	}
	allSubjects, err := driver.FindElements(selenium.ByXPATH, subjectNamesTable)
	if err != nil {
		return nil, "", err
	}
	var subjects []int // создаем слайс, куда будем добавлять те индексы, в которых есть информация о баллах предмета
	for i := range allSubjects {
		currSubject := fmt.Sprintf(currentSubjectField, 1, i+1)
		subject, err := driver.FindElement(selenium.ByXPATH, currSubject+"/td")
		if err != nil {
			return nil, "", err
		}
		classItems, err := subject.GetAttribute("class")
		if err != nil {
			return nil, "", err
		}
		// ng-star-inserted - это предмет, в котором есть оценки -> нужно добавить в слайс,
		// academic-course-td - что-то типо тематического раздела предметов (оценок нет)
		// также есть вариант, когда у предмета нет раздела (он нам тоже нужен), но это та же проверка на != academic-course-td
		if classItems != "academic-course-td" {
			subjects = append(subjects, i+1)
		}
	}
	text := "Вот все Ваши оценки по предметам:"
	//var stateData map[int]int
	stateData := make(map[int]int)
	for i, index := range subjects {
		currGrade, err := parseGrades(driver, i+1, index)
		if err != nil {
			return nil, "", err
		}
		text += "\n" + currGrade
		stateData[i+1] = index
	}
	return stateData, text, nil
}

func (s *Selenium) setupGradesPage(driver selenium.WebDriver, login, password string, timeout time.Duration) error {
	if err := s.loginPage(driver, login, password, timeout); err != nil {
		return err
	}
	if err := driver.Get(defaultModeusUrl + "students-app/my-results/"); err != nil {
		return err
	}
	if err := driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		gradesLoaded, _ := wd.FindElement(selenium.ByXPATH, gradesLoadedInfoField)
		if gradesLoaded != nil {
			return gradesLoaded.IsEnabled()
		}
		return false, nil
	}, timeout); err != nil {
		return err
	}
	return nil
}

// SubjectDetailedInfo для просмотра информации о всех проведенных занятиях (баллы и отметка посещения).
// На вход (помимо основных параметров) принимает индекс предмета, информацию о котором будем искать.
func (s *Selenium) SubjectDetailedInfo(driver selenium.WebDriver, login, password string, index int, timeout time.Duration) (string, error) {
	if err := s.setupGradesPage(driver, login, password, timeout); err != nil {
		return "", err
	}
	return getDetailedSubjectInfo(driver, index)
}

func parseGrades(driver selenium.WebDriver, i, index int) (string, error) {
	currSubject := fmt.Sprintf(currentSubjectField, 1, index) + "//span[2]" // Прибавляем //span[2], в котором находится нужное значение
	currMarks := fmt.Sprintf(currentSubjectField, 2, index)
	subject, err := driver.FindElement(selenium.ByXPATH, currSubject)
	if err != nil {
		return "", err
	}
	subjectName, _ := subject.Text()
	marks, err := driver.FindElement(selenium.ByXPATH, currMarks+currentResultField)
	if err != nil {
		marks, err = driver.FindElement(selenium.ByXPATH, currMarks+"/td[@class='not-active current-result ng-star-inserted']")
		if err != nil {
			return "", err
		}
	}
	currResult, _ := marks.Text()
	marks, err = driver.FindElement(selenium.ByXPATH, currMarks+finalResultField)
	if err != nil {
		marks, err = driver.FindElement(selenium.ByXPATH, currMarks+"/td[@class='not-active final-result ng-star-inserted']")
		if err != nil {
			return "", err
		}
	}
	finalResult, _ := marks.Text()
	marks, err = driver.FindElement(selenium.ByXPATH, currMarks+currentAttendanceField)
	if err != nil {
		marks, err = driver.FindElement(selenium.ByXPATH, currMarks+"/td[@class='not-active rates-present ng-star-inserted']")
		if err != nil {
			return "", err
		}
	}
	currAttendance, _ := marks.Text()
	marks, err = driver.FindElement(selenium.ByXPATH, currMarks+currentMissingField)
	if err != nil {
		marks, err = driver.FindElement(selenium.ByXPATH, currMarks+"/td[@class='not-active rates-absebt ng-star-inserted']")
		if err != nil {
			return "", err
		}
	}
	currMissing, _ := marks.Text()
	marks, err = driver.FindElement(selenium.ByXPATH, currMarks+currentWithoutMarkField)
	if err != nil {
		marks, err = driver.FindElement(selenium.ByXPATH, currMarks+"/td[@class='not-active rates-undefined ng-star-inserted']")
		if err != nil {
			return "", err
		}
	}
	currWithoutMark, _ := marks.Text()
	return fmt.Sprintf(txtUserGrades, i, subjectName, currResult, finalResult, currAttendance+"%", currMissing+"%", currWithoutMark+"%"), nil
}

func getDetailedSubjectInfo(driver selenium.WebDriver, index int) (string, error) {
	currSubjectMarks := fmt.Sprintf(currentSubjectField, 2, index) // + /td[%d] in range(1, len(allMeetings))
	allMeetings, err := driver.FindElements(selenium.ByXPATH, currSubjectMarks+"/td")
	if err != nil {
		return "", err
	}
	text := "Вот вся информация о пройденных встречах:" // TODO добавлять название предмета
	for i := 1; i < len(allMeetings); i++ {             // идем циклом от 1 до максимального количества встреч в семестре
		meetingField := currSubjectMarks + fmt.Sprintf("/td[%d]", i) // Находим текущую (i) встречу. Внутри может быть /span с /span с контентом, либо ничего
		meeting, _ := driver.FindElement(selenium.ByXPATH, meetingField+"/span")
		if meeting != nil {
			classItems, err := meeting.GetAttribute("class")
			if err != nil {
				continue
			}
			if strings.Contains(classItems, meetingStatusNotHeld) {
				break // нас интересует информация только о проведенных встречах
			}
			lessonType := "Неизвестно"
			for _, item := range strings.Fields(classItems) {
				if s, ok := lessonTypes[item]; ok {
					lessonType = s
					break
				}
			}
			marks, err := driver.FindElements(selenium.ByXPATH, meetingField+"/span/span") // внутри есть несколько span с информацией о баллах за пару и один span с отметкой посещаемости
			if err != nil {
				continue
			}
			lessonStatus := "Неизвестно"
			lessonGrades := ""

			for _, el := range marks {
				classItems, err = el.GetAttribute("class")
				// может быть либо lesson-realization ng-star-inserted - оценка за пару (может несколько)
				// либо lesson-realization present - отметка посещаемости
				if err != nil {
					continue
				}
				if strings.Contains(classItems, "present") { // проверяем, что это span с отметкой посещения
					lessonStatus, _ = el.Text()
					continue
				}
				// значит это span с оценкой
				grade, _ := el.Text()
				lessonGrades += " " + grade
			}
			if len(lessonGrades) == 0 {
				lessonGrades = "Баллов нет"
			}
			text += "\n" + fmt.Sprintf(txtDetailedLessonInfo, i, lessonType, lessonStatus, lessonGrades)
		}
	}
	return text, nil
}
