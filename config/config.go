package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Bot      Bot
	PG       PG
	Log      Log
	Selenium Selenium
	Broker   Broker
	Root     Root
}

type (
	Bot struct {
		Token string `env:"BOT_TOKEN"`
	}
	PG struct {
		MaxPoolSize int    `env:"PG_MAX_POOL_SIZE"`
		Url         string `env:"PG_URL"`
	}
	Log struct {
		Level  string `env:"LOG_LEVEL"`
		Output string `env:"LOG_OUTPUT"`
	}
	Selenium struct {
		Url      string `env:"SELENIUM_URL"`
		LocalUrl string `env:"LOCAL_CLIENT"`
	}
	Root struct {
		Login    string `env:"MAIN_USER"`
		Password string `env:"MAIN_PASS"`
	}
	Broker struct {
		Url string `env:"BROKER_URL"`
	}
)

func NewConfig() (*Config, error) {
	c := &Config{}
	if err := cleanenv.ReadEnv(c); err != nil {
		return nil, fmt.Errorf("error reading config env: %w", err)
	}
	return c, nil
}
