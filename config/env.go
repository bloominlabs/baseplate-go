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

func GetEnvInt(key string, def int) (int, error) {
	strValue := os.Getenv(key)
	if strValue == "" {
		return def, nil
	}
	v, err := strconv.Atoi(strValue)
	if err != nil {
		return 0, err
	}
	return v, nil
}
