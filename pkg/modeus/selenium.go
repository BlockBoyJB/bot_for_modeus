package modeus

import (
	"errors"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"strings"
	"time"
)

const (
	loginEmailPlaceholder    = "//input[@id='userNameInput']"
	loginPasswordPlaceholder = "//input[@id='passwordInput']"
	loginButtonPlaceholder   = "//span[@class='submit']"
	incorrectInputData       = "//span[@id='errorText']"
	defaultRedirectTimeout   = time.Second * 10 // Пользователь может ввести некорректный пароль в бота, поэтому будет беда, потому что будет не редирект, а сообщение с ошибкой о неправильном пароле
)

type Selenium struct {
	url     string
	caps    selenium.Capabilities
	service *selenium.Service // for local remote
}

func NewSeleniumFromConfig(config, url, path string) (*Selenium, error) {
	switch config {
	case "local":
		return newLocalClient(path)
	case "remote":
		return newRemoteClient(url), nil
	default:
		return nil, errors.New("cannot parse selenium client config")
	}
}

// newRemoteClient инициализация с уже существующим сервисом
func newRemoteClient(url string) *Selenium {
	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{
		Args: []string{
			"--headless",
			"--incognito",
			"--no-sandbox",
			"--disable-gpu",
			"window-size=1024,768",
		},
		W3C: true,
	})
	return &Selenium{
		url:  url,
		caps: caps,
	}
}

// newLocalClient локальная инициализация
func newLocalClient(path string) (*Selenium, error) {
	service, err := selenium.NewChromeDriverService(path, 4444)
	if err != nil {
		return nil, err
	}
	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{
		Args: []string{
			"--headless",
			"--incognito",
			"--no-sandbox",
			"--disable-gpu",
			"window-size=1024,768",
		},
		//W3C: true,
	})
	return &Selenium{
		url:     "http://127.0.0.1:4444/wd/hub",
		caps:    caps,
		service: service,
	}, nil
}

func (s *Selenium) CloseClient() {
	if s.service != nil {
		_ = s.service.Stop()
	}
}

// ExtractToken забирает токен из текущей сессии модеуса.
// Спустя огромное количество проб и ошибок (и 403 кодов) было принято решение реализовать авторизацию в модеусе через браузерное окно
// Модеус ТюмГУ имеет какую то странную защиту из множества редиректов, из-за которой авторизация только через http становится какой-то нереальной
func (s *Selenium) ExtractToken(login, password string, timeout time.Duration) (string, error) {
	driver, err := selenium.NewRemote(s.caps, s.url)
	if err != nil {
		return "", err
	}
	defer func() { _ = driver.Quit() }()

	if err = driver.Get(defaultModeusUrl); err != nil {
		return "", err
	}
	err = driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		form, _ := wd.FindElement(selenium.ByXPATH, loginEmailPlaceholder)
		if form != nil {
			return form.IsDisplayed()
		}
		return false, nil
	}, timeout)
	if err != nil {
		return "", err
	}
	email, err := driver.FindElement(selenium.ByXPATH, loginEmailPlaceholder)
	if err != nil {
		return "", err
	}
	if err = email.SendKeys(login); err != nil {
		return "", err
	}
	pass, err := driver.FindElement(selenium.ByXPATH, loginPasswordPlaceholder)
	if err != nil {
		return "", err
	}
	if err = pass.SendKeys(password); err != nil {
		return "", err
	}
	btn, err := driver.FindElement(selenium.ByXPATH, loginButtonPlaceholder)
	if err != nil {
		return "", err
	}
	if err = btn.Click(); err != nil {
		return "", err
	}
	// После заполнения формы и нажатия на кнопку мы ждем, пока нас редиректнет на страницу модеуса.
	// Тут нужно повесить ожидание, потому что это делается не мгновенно
	// К тому же, если пароль/логин неверные, то редирект не произойдет
	err = driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		url, e := driver.CurrentURL()
		if e != nil {
			return false, e
		}
		return strings.Contains(url, defaultModeusUrl), nil
	}, defaultRedirectTimeout)
	if err != nil {
		// это может быть не только ошибка пользователя, а какие-то внешние проблемы,
		// поэтому проверяем поле, которое отображается после неправильного ввода
		errField, e := driver.FindElement(selenium.ByXPATH, incorrectInputData)
		if e != nil {
			return "", e
		}
		text, e := errField.Text()
		if e != nil {
			return "", e
		}
		// Если надпись появилась, значит это точно ошибка пользователя
		if text != "" {
			return "", ErrIncorrectInputData
		}
		return "", err
	}

	var result interface{}
	err = driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		result, err = driver.ExecuteScript(`return window.sessionStorage.getItem("id_token");`, nil)
		if err != nil {
			return false, err
		}
		return result != nil, nil
	}, timeout)
	if err != nil {
		return "", err
	}
	token, ok := result.(string)
	if !ok {
		return "", errors.New("cannot extract token")
	}
	return token, nil
}
