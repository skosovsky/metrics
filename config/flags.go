package config

import (
	"errors"
	"fmt"
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

func (r *Receiver) String() string {
	return r.Address
}

func (r *Receiver) Set(flagValue string) error {
	const flagParts = 2

	flagValues := strings.Split(flagValue, ":")

	if len(flagValues) != flagParts {
		return fmt.Errorf("parsing address error - %s: %w", flagValue, ErrInvalidReceiverAddress)
	}

	// port, err := strconv.Atoi(flagValues[1])
	// if err != nil {
	// 	return fmt.Errorf("parsing port error - %r: %w", flagValue, ErrInvalidReceiverPort)
	// }
	//
	// r.Port = port
	// r.Host = flagValues[0]

	r.Address = flagValues[0] + ":" + flagValues[1]

	return nil
}

func (t *Transmitter) String() string {
	return t.Address
}

func (t *Transmitter) Set(flagValue string) error {
	const flagParts = 2

	flagValues := strings.Split(flagValue, ":")

	if len(flagValues) != flagParts {
		return fmt.Errorf("parsing address error - %s: %w", flagValue, ErrInvalidReceiverAddress)
	}

	// port, err := strconv.Atoi(flagValues[1])
	// if err != nil {
	//	return fmt.Errorf("parsing port error - %s: %w", flagValue, ErrInvalidReceiverPort)
	// }
	//
	// t.Port = port
	// t.Host = flagValues[0]

	t.Address = flagValues[0] + ":" + flagValues[1]

	return nil
}
