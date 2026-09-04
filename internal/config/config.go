package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppID       int
	AppHash     string
	SessionFile string
}

func Load() (Config, error) {
	appIDRaw := os.Getenv("APP_ID")
	if appIDRaw == "" {
		return Config{}, fmt.Errorf("APP_ID is required")
	}

	appID, err := strconv.Atoi(appIDRaw)
	if err != nil || appID <= 0 {
		return Config{}, fmt.Errorf("APP_ID must be a positive integer")
	}

	appHash := os.Getenv("APP_HASH")
	if appHash == "" {
		return Config{}, fmt.Errorf("APP_HASH is required")
	}

	if os.Getenv("TELEGRAM_PHONE") == "" {
		return Config{}, fmt.Errorf("TELEGRAM_PHONE is required")
	}

	sessionFile := os.Getenv("SESSION_FILE")
	if sessionFile == "" {
		sessionFile = ".data/session.json"
	}

	return Config{
		AppID:       appID,
		AppHash:     appHash,
		SessionFile: sessionFile,
	}, nil
}
