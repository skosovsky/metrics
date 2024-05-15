package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidReceiverAddress = errors.New("invalid receiver address")
	ErrInvalidReceiverPort    = errors.New("invalid receiver port")
)

type Value interface {
	String() string
	Set(string) error
}

func (s *Receiver) String() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

func (s *Receiver) Set(flagValue string) error {
	const flagParts = 2

	flagValues := strings.Split(flagValue, ":")

	if len(flagValues) != flagParts {
		return fmt.Errorf("parsing address error - %s: %w", flagValue, ErrInvalidReceiverAddress)
	}

	port, err := strconv.Atoi(flagValues[1])
	if err != nil {
		return fmt.Errorf("parsing port error - %s: %w", flagValue, ErrInvalidReceiverPort)
	}

	s.Port = port
	s.Host = flagValues[0]

	return nil
}

func (t *Transmitter) String() string {
	return t.Host + ":" + strconv.Itoa(t.Port)
}

func (t *Transmitter) Set(flagValue string) error {
	const flagParts = 2

	flagValues := strings.Split(flagValue, ":")

	if len(flagValues) != flagParts {
		return fmt.Errorf("parsing address error - %s: %w", flagValue, ErrInvalidReceiverAddress)
	}

	port, err := strconv.Atoi(flagValues[1])
	if err != nil {
		return fmt.Errorf("parsing port error - %s: %w", flagValue, ErrInvalidReceiverPort)
	}

	t.Port = port
	t.Host = flagValues[0]

	return nil
}
