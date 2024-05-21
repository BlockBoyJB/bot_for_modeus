package parser

import (
	"fmt"
	"github.com/tebeka/selenium"
	"strings"
	"time"
)

const (
	//loginEmailPlaceholder    = "//input[@placeholder='proverka@example.com']" // Некорректно работает
	//loginPasswordPlaceholder = "//input[@placeholder='Пароль']" // Некорректно работает
	loginEmailPlaceholder    = "//input[@id='userNameInput']"
	loginPasswordPlaceholder = "//input[@id='passwordInput']"
	loginButtonPlaceholder   = "//span[@class='submit']"
	filterButton             = "//button[@class='btn-filter screen-only']"                                                  // Кнопка "фильтр"
	clearButton              = "//button[@class='btn btn-clear']"                                                           // Кнопка "очистить"
	arrowDown                = "//span[@class='ui-multiselect-trigger-icon ui-clickable ng-tns-c56-10 pi pi-chevron-down']" // Стрелочка вниз, нажимаем и открывается меню с поиском пользователя
	searchField              = "//input[@placeholder='Поиск...']"                                                           // Поле для поиска пользователя
	userNotFoundField        = "//li[@class='ui-multiselect-empty-message ng-tns-c56-10 ng-star-inserted']"                 // Если пользователь не найден, то появляется такое поле
	executeButton            = "//button[@class='btn btn-apply']"                                                           // Кнопка "применить"
	multiselectField         = "//p-multiselectitem/li[@aria-label='%s']"                                                   // Поле найденного пользователя. Вместо %s внутри aria-label подставляем ФИО искомого пользователя
	// TODO aria-label выдает баг, когда имя пользователя введено с ошибкой или есть лишние пробелы (пользователь не находится)

	defaultRedirectTimeout = time.Second * 7 // Пользователь может ввести некорректный пароль в бота, поэтому будет беда, потому что будет не редирект, а сообщение с ошибкой о неправильном пароле
	defaultFindUserField   = time.Second * 5
)

// Решил каждый раз при запросе открывать новое окно вместо одного окна при запуске, чтобы избежать проблем с сессионными токенами (can be expired)
func (s *Selenium) loginPage(driver selenium.WebDriver, login, password string, timeout time.Duration) error {
	if err := driver.Get(defaultModeusUrl); err != nil {
		return err
	}
	err := driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		form, _ := wd.FindElement(selenium.ByXPATH, loginEmailPlaceholder)
		if form != nil {
			return form.IsDisplayed()
		}
		return false, nil
	}, timeout)
	if err != nil {
		return ErrFindElementTimeout
	}
	email, err := driver.FindElement(selenium.ByXPATH, loginEmailPlaceholder)
	if err != nil {
		return err
	}
	if err = email.SendKeys(login); err != nil {
		return err
	}
	pass, err := driver.FindElement(selenium.ByXPATH, loginPasswordPlaceholder)
	if err != nil {
		return err
	}
	if err = pass.SendKeys(password); err != nil {
		return err
	}
	btn, err := driver.FindElement(selenium.ByXPATH, loginButtonPlaceholder)
	if err != nil {
		return err
	}
	if err = btn.Click(); err != nil {
		return err
	}
	err = driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		url, err := driver.CurrentURL()
		if err != nil {
			return false, err
		}
		return strings.Contains(url, defaultModeusUrl), nil
	}, defaultRedirectTimeout)
	if err != nil {
		return ErrIncorrectUserData
	}
	return nil
}

// SetupUser вообще не нужен, если у нас есть логин и пароль от пользователя
func (s *Selenium) SetupUser(driver selenium.WebDriver, login, password, fullName string, timeout time.Duration) error {
	if err := s.loginPage(driver, login, password, timeout); err != nil {
		return err
	}
	err := driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		filter, _ := wd.FindElement(selenium.ByXPATH, filterButton)
		if filter != nil {
			return filter.IsDisplayed()
		}
		return false, nil
	}, timeout)
	if err != nil {
		return ErrFindElementTimeout
	}
	filter, err := driver.FindElement(selenium.ByXPATH, filterButton)
	if err != nil {
		return err
	}
	if err = filter.Click(); err != nil {
		return err
	}
	clearBtn, err := driver.FindElement(selenium.ByXPATH, clearButton)
	if err != nil {
		return err
	}
	if err = clearBtn.Click(); err != nil {
		return err
	}
	arrow, err := driver.FindElement(selenium.ByXPATH, arrowDown)
	if err != nil {
		return err
	}
	if err = arrow.Click(); err != nil {
		return err
	}
	input, err := driver.FindElement(selenium.ByXPATH, searchField)
	if err != nil {
		return err
	}
	if err = input.SendKeys(fullName); err != nil {
		return err
	}

	// старый вариант проверки
	//userField := fmt.Sprintf(multiselectField, fullName)
	//err = driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) { // проверяем, что пользователь в строке поиска загрузился
	//	user, _ := wd.FindElement(selenium.ByXPATH, userField)
	//	if user != nil {
	//		return user.IsDisplayed()
	//	}
	//	return false, nil
	//}, defaultFindUserField)
	var fieldFullName string
	err = driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		user, _ := wd.FindElement(selenium.ByXPATH, "//div[@class='label']") // TODO надо как то проверять что пользователь ввел свое ФИО без ошибок, потому что такой подход отображает всех найденных пользователей
		if user != nil {
			u, _ := user.Text()
			if strings.ToLower(u) == strings.ToLower(fullName) {
				fieldFullName = u
				return true, nil
			}
		}
		return false, nil
	}, defaultFindUserField)

	// старый вариант проверки
	//if err != nil {
	//	err = driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
	//		notFound, _ := wd.FindElement(selenium.ByXPATH, userNotFoundField)
	//		if notFound != nil {
	//			return notFound.IsDisplayed()
	//		}
	//		return false, nil
	//	}, defaultFindUserField)
	//	if err == nil {
	//		return ErrIncorrectFullName
	//	}
	//}
	if len(fieldFullName) == 0 {
		return ErrIncorrectFullName
	}
	userField := fmt.Sprintf(multiselectField, fieldFullName)
	user, err := driver.FindElement(selenium.ByXPATH, userField)
	if err != nil {
		return err
	}
	if err = user.Click(); err != nil {
		return err
	}
	executeBtn, err := driver.FindElement(selenium.ByXPATH, executeButton)
	if err != nil {
		return err
	}
	if err = executeBtn.Click(); err != nil {
		return err
	}
	return nil
}
