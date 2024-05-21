// Package parser Такого количества ошибок я еще не хэндлил :/
// Процентов 70 кода это if err != nil {return err}
// Да простят меня те, кто это будет читать
package parser

import (
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"time"
)

const defaultModeusUrl = "https://utmn.modeus.org/"

type Parser interface {
	InitRemote() (selenium.WebDriver, error)
	DaySchedule(driver selenium.WebDriver, login, password, fullName string, timeout time.Duration) (string, error)
	WeekSchedule(driver selenium.WebDriver, login, password, fullName string, timeout time.Duration) (string, error)
	DayScheduleWithLoginPass(driver selenium.WebDriver, login, password string, timeout time.Duration) (string, error)
	WeekScheduleWithLoginPass(driver selenium.WebDriver, login, password string, timeout time.Duration) (string, error)
	UserGrades(driver selenium.WebDriver, login, password string, timeout time.Duration) (map[int]int, string, error)
	SubjectDetailedInfo(driver selenium.WebDriver, login, password string, index int, timeout time.Duration) (string, error)
}

type Selenium struct {
	url     string
	caps    selenium.Capabilities
	service *selenium.Service // for local remote
}

func NewClient(url string) *Selenium {
	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{
		Args: []string{
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

// InitRemote создаем каждый раз снова, тк я не придумал оригинального способа логиниться и парсить данные
func (s *Selenium) InitRemote() (selenium.WebDriver, error) {
	driver, err := selenium.NewRemote(s.caps, s.url)
	if err != nil {
		return nil, err
	}
	return driver, nil
}

// NewLocalClient локальная инициализация
func NewLocalClient(path string) (*Selenium, error) {
	service, err := selenium.NewChromeDriverService(path, 4444)
	if err != nil {
		return nil, err
	}
	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{
		Args: []string{
			"--incognito",
			"--no-sandbox",
			"--disable-gpu",
			"window-size=1024,768",
		},
		//W3C: true,
	})
	return &Selenium{caps: caps, url: "http://127.0.0.1:4444/wd/hub", service: service}, nil
}

func (s *Selenium) CloseLocalClient() {
	_ = s.service.Stop()
}
