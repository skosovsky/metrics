package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
)

const (
	devMode  = "development"
	testMode = "test"
	prodMode = "production"
)

type (
	Address string

	App struct {
		Mode string `env:"APP_MODE" validate:"required,oneof=development production test"`
	}

	Consumer struct {
		Address Address `env:"ADDRESS" validate:"url"`
	}

	Producer struct {
		Address        Address `env:"ADDRESS"         validate:"url"`
		ReportInterval int     `env:"REPORT_INTERVAL" validate:"min=1"`
		PollInterval   int     `env:"POLL_INTERVAL"   validate:"min=1"`
	}

	Store struct {
		DBDriver  string `env:"DB_DRIVER"  validate:"required,oneof=sqlite3 memory"`
		DBAddress string `env:"DB_ADDRESS"`
	}

	ConsumerConfig struct {
		App      App
		Consumer Consumer
		Store    Store
	}

	ProducerConfig struct {
		App      App
		Producer Producer
	}
)

func NewConsumerConfig() (ConsumerConfig, error) {
	var config ConsumerConfig

	err := config.Consumer.Address.Set("localhost:8080")
	if err != nil {
		return ConsumerConfig{}, fmt.Errorf("failed to set default value: %w", err)
	}

	flag.Var(&config.Consumer.Address, "a", "Server address host:port")
	flag.Parse()

	if err = env.Parse(&config); err != nil {
		return ConsumerConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	if err = config.validate(); err != nil {
		return ConsumerConfig{}, fmt.Errorf("failed to validate config: %w", err)
	}

	return config, nil
}

func (c ConsumerConfig) validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(c)
	if err != nil {
		return fmt.Errorf("failed to validate config %v: %w", c, err)
	}

	return nil
}

func NewProducerConfig() (ProducerConfig, error) {
	var config ProducerConfig

	err := config.Producer.Address.Set("localhost:8080")
	if err != nil {
		return ProducerConfig{}, fmt.Errorf("failed to set default value: %w", err)
	}

	flag.Var(&config.Producer.Address, "a", "Server address host:port")
	flag.IntVar(&config.Producer.PollInterval, "p", 2, "Polling interval in seconds")
	flag.IntVar(&config.Producer.ReportInterval, "r", 10, "Reporting interval in seconds")

	flag.Parse()

	if err = env.Parse(&config); err != nil {
		return ProducerConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	if err = config.validate(); err != nil {
		return ProducerConfig{}, fmt.Errorf("failed to validate config: %w", err)
	}

	return config, nil
}

func (c ProducerConfig) validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(c)
	if err != nil {
		return fmt.Errorf("failed to validate config %v: %w", c, err)
	}

	return nil
}
