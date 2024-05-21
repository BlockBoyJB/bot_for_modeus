package parser

import (
	"fmt"
	"github.com/tebeka/selenium"
	"strings"
	"time"
)

const (
	modeusTableField      = "//mds-fullcalendar"
	dayTableField         = "//td[%s]//div[@class='fc-content']"
	subjectInfoClassType  = "//td[%s]//a[%d]"
	subjectInfoField      = "//td[%s]//a[%d]/div[@class='fc-content']" // /td[%s] - номер дня + 1, /a[%s] - номер пары (из общего числа за день)
	subjectDurationLesson = "/div[@class='fc-time']/span"
	subjectAudienceNumber = "/div[@class='fc-time']/small"
	subjectSubjectName    = "/div[@class='fc-title']"
	txtDaySchedule        = "%d. Пара по предмету: %s\nТип занятия: %s\nВремя: %s\nКорпус и номер аудитории: %s\n"
	txtNoSubjectsToday    = "Занятий нет"
)

// Именно такое форматирование, потому что в модеусе дни недели отличаются на +1
var dates = map[int][]string{
	1: {"2", "Понедельник"},
	2: {"3", "Вторник"},
	3: {"4", "Среда"},
	4: {"5", "Четверг"},
	5: {"6", "Пятница"},
	6: {"7", "Суббота"},
	0: {"8", "Воскресенье"},
}

// Все виды занятий
var lessonTypes = map[string]string{
	"mds-event-type-semi":                   "Практическое занятие",
	"mds-event-type-lab":                    "Лабораторное занятие",
	"mds-event-type-lect":                   "Лекционное занятие",
	"mds-event-type-cons":                   "Консультация",
	"mds-event-type-cur_check":              "Текущий контроль",
	"mds-event-type-mid_check":              "Аттестация",
	"mds-event-type-mid_check-examination":  "Аттестация (экзамен)",
	"mds-event-type-mid_check-diff_pretest": "Аттестация (дифференцированный зачет)",
	"mds-event-type-mid_check-pretest":      "Аттестация (зачет)",
	"mds-event-type-self":                   "Самостоятельная работа",
	"mds-event-type-other":                  "Прочее",
}

func parseDaySchedule(driver selenium.WebDriver, dayNumber int, timeout time.Duration) (string, error) {
	// waiting until schedule loads
	err := driver.WaitWithTimeout(func(w selenium.WebDriver) (bool, error) {
		// mds-fullcalendar[@class='fc fc-unthemed fc-ltr loading-events'] - незагруженное состояние
		// если внутри класса есть loading-events, значит что таблица еще не загрузилась
		table, _ := w.FindElement(selenium.ByXPATH, modeusTableField)
		if table != nil {
			classItems, err := table.GetAttribute("class")
			if err != nil {
				return false, err
			}
			return classItems == "fc fc-unthemed fc-ltr", nil
		}
		return false, nil
	}, timeout)
	if err != nil {
		return "", ErrFindElementTimeout
	}
	d := dates[dayNumber]
	number, weekDay := d[0], d[1]
	table := fmt.Sprintf(dayTableField, number)

	text := "День недели: " + weekDay + "\n"
	subjects, err := driver.FindElements(selenium.ByXPATH, table)
	if err != nil {
		return "", err
	}
	if len(subjects) == 0 {
		return text + txtNoSubjectsToday + "\n", nil
	}
	// получаем все элементы и снова делаем поочередные запросы, чтобы получить возможность сплитить информацию
	for i, subject := range subjects {
		// время /div[@class='fc-time']/span
		// аудитория /div[@class='fc-time']/small
		// название /div[@class='fc-title']

		// currentSubject формируется следующим образом: делаем запрос и получаем все пары на день. Идем циклом и формируем Xpath для текущего предмета
		currentSubject := fmt.Sprintf(subjectInfoField, number, i+1)
		currentField := fmt.Sprintf(subjectInfoClassType, number, i+1)
		classItems, err := subject.FindElement(selenium.ByXPATH, currentField)
		if err != nil {
			return "", err
		}
		cssStyles, err := classItems.GetAttribute("class")
		if err != nil {
			return "", err
		}
		lessonType := "Неизвестно"
		listStyles := strings.Fields(cssStyles)
		for _, el := range listStyles {
			if s, ok := lessonTypes[el]; ok {
				lessonType = s
				break
			}
		}
		durationLesson, err := subject.FindElement(selenium.ByXPATH, currentSubject+subjectDurationLesson)
		if err != nil {
			return "", err
		}
		audienceNumber, err := subject.FindElement(selenium.ByXPATH, currentSubject+subjectAudienceNumber)
		if err != nil {
			return "", err
		}
		subjectName, err := subject.FindElement(selenium.ByXPATH, currentSubject+subjectSubjectName)
		if err != nil {
			return "", err
		}
		// ну они просто издеваются мне прописывать три if err != nil {return "", err} удовольствия вообще не приносит
		dl, _ := durationLesson.Text()
		an, _ := audienceNumber.Text()
		sn, _ := subjectName.Text()

		text += fmt.Sprintf(txtDaySchedule, i+1, sn, lessonType, dl, an)
	}
	return text, nil
}

func (s *Selenium) DaySchedule(driver selenium.WebDriver, login, password, fullName string, timeout time.Duration) (string, error) {
	if err := s.SetupUser(driver, login, password, fullName, timeout); err != nil {
		return "", err
	}
	today := time.Now().Weekday()
	text, err := parseDaySchedule(driver, int(today), timeout)
	if err != nil {
		return "", err
	}
	return text, nil
}

func (s *Selenium) WeekSchedule(driver selenium.WebDriver, login, password, fullName string, timeout time.Duration) (string, error) {
	if err := s.SetupUser(driver, login, password, fullName, timeout); err != nil {
		return "", err
	}
	text := "Расписание на неделю: "
	for i := 1; i < 7; i++ { // Возможно, стоит идти циклом (0, 6), потому что в воскресение у кого то могут быть занятия
		daySchedule, err := parseDaySchedule(driver, i, timeout)
		if err != nil {
			return "", err
		}
		text += "\n" + daySchedule
	}
	return text, nil
}

// Глобально функции ниже заточенные под логин + пароль по скорости получения данных сильно не отличаются от дефолтных выше
// Различие явно заметно при низкой скорости соединения на сервере, а при скорости 100+ мбит/сек разница получается ~1-1.5 секунды

func (s *Selenium) DayScheduleWithLoginPass(driver selenium.WebDriver, login, password string, timeout time.Duration) (string, error) {
	if err := s.loginPage(driver, login, password, timeout); err != nil {
		return "", err
	}
	today := time.Now().Weekday()
	text, err := parseDaySchedule(driver, int(today), timeout)
	if err != nil {
		return "", err
	}
	return text, nil
}

func (s *Selenium) WeekScheduleWithLoginPass(driver selenium.WebDriver, login, password string, timeout time.Duration) (string, error) {
	if err := s.loginPage(driver, login, password, timeout); err != nil {
		return "", err
	}

	text := "Расписание на неделю: "
	for i := 1; i < 7; i++ {
		daySchedule, err := parseDaySchedule(driver, i, timeout)
		if err != nil {
			return "", err
		}
		text += "\n" + daySchedule
	}
	return text, nil
}
