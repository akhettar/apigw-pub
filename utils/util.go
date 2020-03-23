package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func FetchEnvVar(key, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if ok {
		return value
	}
	log.WithFields(log.Fields{
		"environment variable key": key,
		"default":                  defaultValue,
	}).Warn("Environment variable not found returning default value")
	return defaultValue
}

func RetrieveEnvVar(key string) string {
	value, ok := os.LookupEnv(key)
	if ok {
		return value
	}
	log.WithFields(log.Fields{
		"environment variable key": key,
	}).Warn("Failed to retrieve environment variable")
	return ""
}
