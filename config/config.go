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
		Mode string `validate:"required,oneof=development production test" env:"APP_MODE"`
	}

	Receiver struct {
		Address Address `validate:"url" env:"ADDRESS"`
	}

	Transmitter struct {
		Address        Address `validate:"url"   env:"ADDRESS"`
		ReportInterval int     `validate:"min=1" env:"REPORT_INTERVAL"`
		PollInterval   int     `validate:"min=1" env:"POLL_INTERVAL"`
	}

	Store struct {
		DBDriver  string `validate:"required,oneof=sqlite3 memory" env:"DB_DRIVER"`
		DBAddress string `                                         env:"DB_ADDRESS"`
	}

	ReceiverConfig struct {
		App      App
		Receiver Receiver
		Store    Store
	}

	TransmitterConfig struct {
		App         App
		Transmitter Transmitter
	}
)

func NewReceiverConfig() (ReceiverConfig, error) {
	var config ReceiverConfig

	err := config.Receiver.Address.Set("localhost:8080")
	if err != nil {
		return ReceiverConfig{}, fmt.Errorf("failed to set default value: %w", err)
	}

	flag.Var(&config.Receiver.Address, "a", "Server address host:port")
	flag.Parse()

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

	err := config.Transmitter.Address.Set("localhost:8080")
	if err != nil {
		return TransmitterConfig{}, fmt.Errorf("failed to set default value: %w", err)
	}

	flag.Var(&config.Transmitter.Address, "a", "Server address host:port")
	flag.IntVar(&config.Transmitter.PollInterval, "p", 2, "Polling interval in seconds")
	flag.IntVar(&config.Transmitter.ReportInterval, "r", 10, "Reporting interval in seconds")

	flag.Parse()

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
