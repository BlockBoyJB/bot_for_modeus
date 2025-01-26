package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Bot     Bot
	MongoDB MongoDB
	Redis   Redis
	Log     Log
	Crypter Crypter
	Parser  Parser
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
	Crypter struct {
		Secret string `env-required:"true" env:"SECRET"`
	}
	Parser struct {
		Host string `env-required:"true" env:"PARSER_HOST"`
	}
)

func NewConfig() (*Config, error) {
	c := &Config{}
	if err := cleanenv.ReadEnv(c); err != nil {
		return nil, fmt.Errorf("error reading config env: %w", err)
	}
	return c, nil
}
