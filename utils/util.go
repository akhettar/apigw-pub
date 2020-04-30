package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
)

// FetchEnvVar returns environment variable if not found it will return the given default value
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

// RetrieveEnvVar returns environment and if not found it logs a fatal error
func RetrieveEnvVar(key string) string {
	value, ok := os.LookupEnv(key)
	if ok {
		return value
	}
	log.WithFields(log.Fields{
		"Environment variable": key,
	}).Fatal("Failed to retrieve required environment variable")
	return ""
}
