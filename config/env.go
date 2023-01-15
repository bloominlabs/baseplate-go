package config

import (
	"fmt"
	"os"
	"strconv"
)

func GetEnvStr(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("getenv: environment variable empty - %s", key)
	}
	return value, nil
}

func GetEnvStrDefault(key string, def string) string {
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

func GetEnvBool(key string) (bool, error) {
	strValue := os.Getenv(key)
	if strValue == "" {
		return false, fmt.Errorf("environment varialbe empty - %s", key)
	}
	v, err := strconv.ParseBool(strValue)
	if err != nil {
		return false, err
	}
	return v, nil
}

func GetEnvBoolDefault(key string, def bool) bool {
	value, err := GetEnvBool(key)
	if err != nil {
		return def
	}

	return value
}

func MustGetEnvBool(key string) bool {
	v, err := GetEnvBool(key)
	if err != nil {
		panic(err)
	}
	return v
}
