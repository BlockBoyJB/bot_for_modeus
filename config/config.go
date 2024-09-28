package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Bot          Bot
	MongoDB      MongoDB
	Redis        Redis
	Log          Log
	Root         Root
	Crypter      Crypter
	TokenService TokenService
}

type (
	Bot struct {
		Token     string `env-required:"true" env:"BOT_TOKEN"`
		IsWebhook bool   `env-required:"true" env:"BOT_WEBHOOK"`
	}
	MongoDB struct {
		Url string `env-required:"true" env:"MONGO_URL"`
		DB  string `env-required:"true" env:"MONGO_DB"`
	}
	Redis struct {
		Url string `env-required:"true" env:"REDIS_URL"`
	}
	Log struct {
		Level  string `env-required:"true" env:"LOG_LEVEL"`
		Output string `env-required:"true" env:"LOG_OUTPUT"`
	}
	Root struct {
		Login    string `env-required:"true" env:"MAIN_USER"`
		Password string `env-required:"true" env:"MAIN_PASS"`
	}
	Crypter struct {
		Secret string `env-required:"true" env:"SECRET"`
	}
	TokenService struct {
		Url string `env-required:"true" env:"TOKEN_SERVICE_URL"`
	}
)

func NewConfig() (*Config, error) {
	c := &Config{}
	if err := cleanenv.ReadEnv(c); err != nil {
		return nil, fmt.Errorf("error reading config env: %w", err)
	}
	return c, nil
}
