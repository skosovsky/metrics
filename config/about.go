package config

import (
	"errors"
	"runtime/debug"
	"strings"
)

var ErrInfoUnknown = errors.New("app name or revision unknown")

type Application struct {
	name     string
	revision string
}

func NewAppInfo() (Application, error) {
	var application Application

	if info, ok := debug.ReadBuildInfo(); ok {
		application.name = strings.ToUpper(info.Main.Path)

		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				application.revision = setting.Value

				break
			}
		}
	}

	if application.name == "" || application.revision == "" {
		return Application{}, ErrInfoUnknown
	}

	return application, nil
}
