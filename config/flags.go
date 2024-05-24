package config

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidAddress = errors.New("invalid address")
)

type Value interface {
	String() string
	Set(string) error
}

func (a *Address) String() string {
	return string(*a)
}

func (a *Address) Set(flagValue string) error {
	const flagParts = 2

	flagValues := strings.Split(flagValue, ":")

	if len(flagValues) != flagParts {
		return fmt.Errorf("parsing address error - %s: %w", flagValue, ErrInvalidAddress)
	}

	*a = Address(flagValues[0] + ":" + flagValues[1])

	return nil
}
