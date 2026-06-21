package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `env:"APP_ENV" env-default:"development"`
	Log      LogConfig
	Telegram TelegramConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

type JWTConfig struct {
	Secret string        `env:"JWT_SECRET" env-required:"true"`
	TTL    time.Duration `env:"JWT_TTL" env-default:"72h"`
}

type LogConfig struct {
	Level string `env:"LOG_LEVEL" env-default:"info"`
	JSON  bool   `env:"LOG_JSON" env-default:"false"`
}

type TelegramConfig struct {
	AppID       int    `env:"TG_APP_ID" env-required:"true"`
	AppHash     string `env:"TG_APP_HASH" env-required:"true"`
	ProxyAddr   string `env:"TG_PROXY_ADDR" env-default:""`
	ProxySecret string `env:"TG_PROXY_SECRET" env-default:""`
}

type RedisConfig struct {
	Host     string        `env:"REDIS_HOST" env-default:"localhost"`
	Port     string        `env:"REDIS_PORT" env-default:"6379"`
	Password string        `env:"REDIS_PASSWORD" env-default:""`
	DB       int           `env:"REDIS_DB" env-default:"0"`
	Timeout  time.Duration `env:"REDIS_TIMEOUT" env-default:"5s"`
}

var (
	instance *Config
	once     sync.Once
)

func NewConfig() (*Config, error) {
	var err error
	once.Do(func() {
		var cfg Config
		if err = cleanenv.ReadConfig(".env", &cfg); err != nil {
			err = cleanenv.ReadEnv(&cfg)
		}
		instance = &cfg
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return instance, nil
}