package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	App struct {
		Mode string `env:"APP_MODE" validate:"required,oneof=development production test"`
	}
	Server struct {
		Host string `env:"SRV_HOST" validate:"required"`
		Port int    `env:"SRV_PORT" validate:"required,min=0,max=65535"`
	}
	Store struct {
		DBDriver  string `env:"DB_DRIVER"  validate:"required,oneof=sqlite3 memory"`
		DBAddress string `env:"DB_ADDRESS"`
	}
}

func NewConfig() (Config, error) {
	var config Config

	if err := env.Parse(&config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := config.validate(); err != nil {
		return Config{}, fmt.Errorf("failed to validate config: %w", err)
	}

	return config, nil
}

func (c Config) validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(c)
	if err != nil {
		return fmt.Errorf("failed to validate config %v: %w", c, err)
	}

	return nil
}
