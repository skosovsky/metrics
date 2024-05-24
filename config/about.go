package config

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"metrics/internal/log"
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

func LogAppInfo() error {
	if os.Getenv("APP_MODE") == testMode || os.Getenv("APP_MODE") == prodMode {
		return nil
	}

	appInfo, err := NewAppInfo()
	if err != nil {
		return fmt.Errorf("get app info: %w", err)
	}

	log.Info("appInfo",
		log.AnyAttr("app", fmt.Sprint(appInfo)))

	return nil
}
