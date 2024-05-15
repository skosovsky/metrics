package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
)

type (
	App struct {
		Mode string `env:"APP_MODE" validate:"required,oneof=development production test"`
	}

	Receiver struct {
		Address string `env:"ADDRESS" validate:"url"`
		// Host string `env:"RCV_HOST"       validate:"required"`
		// Port int    `env:"RCV_PORT"       validate:"required,min=0,max=65535"`
	}

	Transmitter struct {
		Address        string `env:"ADDRESS"         validate:"url"`
		ReportInterval int    `env:"REPORT_INTERVAL" validate:"min=1"`
		PollInterval   int    `env:"POLL_INTERVAL"   validate:"min=1"`
		// Host        string `env:"TSM_HOST"        validate:"required"`
		// Port        int    `env:"TSM_PORT"        validate:"required,min=0,max=65535"`
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

	var configReceiver Receiver
	err := configReceiver.Set("localhost:8080")
	if err != nil {
		return ReceiverConfig{}, fmt.Errorf("failed to set default value: %w", err)
	}

	flag.Var(&configReceiver, "a", "Server address host:port")
	flag.Parse()

	config.Receiver = configReceiver

	if err = env.Parse(&config); err != nil {
		return ReceiverConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

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

	configTransmitter.PollInterval = pollInterval
	configTransmitter.ReportInterval = reportInterval
	config.Transmitter = configTransmitter

	if err = env.Parse(&config); err != nil {
		return TransmitterConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

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
