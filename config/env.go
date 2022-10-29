package config

import (
	"fmt"
	"os"
)

func GetEnvStr(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("getenv: environment variable empty - %s", key)
	}
	return value, nil
}

func GetEnvDefault(key string, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
}

func MustGetEnvStr(key string) string {
	value, err := GetEnvStr(key)
	if err != nil {
		panic("Failed to get environment variable: " + key)
	}
	return value
}
