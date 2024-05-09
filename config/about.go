package config

import (
	"errors"
	"runtime/debug"
	"strings"
)

var ErrInfoUnknown = errors.New("app name or revision unknown")

type App struct {
	name     string
	revision string
}

func NewAppInfo() (App, error) {
	var app App

	if info, ok := debug.ReadBuildInfo(); ok {
		app.name = strings.ToUpper(info.Main.Path)

		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				app.revision = setting.Value

				break
			}
		}
	}

	if app.name == "" || app.revision == "" {
		return App{}, ErrInfoUnknown
	}

	return app, nil
}
