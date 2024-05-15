package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
)

type (
	App struct {
		Mode string `env:"APP_MODE" validate:"required,oneof=development production test"`
	}

	Receiver struct {
		Host string `env:"RCV_HOST" validate:"required"`
		Port int    `env:"RCV_PORT" validate:"required,min=0,max=65535"`
	}

	Transmitter struct {
		Host           string        `env:"TSM_HOST"            validate:"required"`
		Port           int           `env:"TSM_PORT"            validate:"required,min=0,max=65535"`
		ReportInterval time.Duration `env:"TSM_REPORT_INTERVAL" validate:"required,min=1s"`
		PollInterval   time.Duration `env:"TSM_POLL_INTERVAL"   validate:"required,min=1s"`
	}

	Store struct {
		DBDriver  string `env:"DB_DRIVER"  validate:"required,oneof=sqlite3 memory"`
		DBAddress string `env:"DB_ADDRESS"`
	}

	ReceiverConfig struct {
		App
		Receiver
		Store
	}

	TransmitterConfig struct {
		App
		Transmitter
	}
)

func NewReceiverConfig() (ReceiverConfig, error) {
	var config ReceiverConfig

	if err := env.Parse(&config); err != nil {
		return ReceiverConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	var configReceiver Receiver

	err := configReceiver.Set("localhost:8080")
	if err != nil {
		return ReceiverConfig{}, fmt.Errorf("failed to set default value: %w", err)
	}

	flag.Var(&configReceiver, "a", "Server address host:port")
	flag.Parse()

	config.Receiver = configReceiver

	if err = config.validate(); err != nil {
		return ReceiverConfig{}, fmt.Errorf("failed to validate config: %w", err)
	}

	return config, nil
}

func (c ReceiverConfig) validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(c)
	if err != nil {
		return fmt.Errorf("failed to validate config %v: %w", c, err)
	}

	return nil
}

func NewTransmitterConfig() (TransmitterConfig, error) {
	var config TransmitterConfig

	if err := env.Parse(&config); err != nil {
		return TransmitterConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	var configTransmitter Transmitter
	var pollInterval int
	var reportInterval int

	err := configTransmitter.Set("localhost:8080")
	if err != nil {
		return TransmitterConfig{}, fmt.Errorf("failed to set default value: %w", err)
	}

	flag.Var(&configTransmitter, "a", "Server address host:port")
	flag.IntVar(&pollInterval, "p", 2, "Polling interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "Reporting interval in seconds")
	flag.Parse()

	configTransmitter.PollInterval = time.Duration(pollInterval) * time.Second
	configTransmitter.ReportInterval = time.Duration(reportInterval) * time.Second
	config.Transmitter = configTransmitter

	if err = config.validate(); err != nil {
		return TransmitterConfig{}, fmt.Errorf("failed to validate config: %w", err)
	}

	return config, nil
}

func (c TransmitterConfig) validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(c)
	if err != nil {
		return fmt.Errorf("failed to validate config %v: %w", c, err)
	}

	return nil
}
