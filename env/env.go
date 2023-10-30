package env

import (
	"fmt"
	"os"
)

// Must returns the value if it exists else it panics
func Must(key string) string {
	s, ok := os.LookupEnv(key)
	if err := fmt.Errorf("no environment variable with the key %s", key); !ok {
		panic(err)
	}
	return s
}

// WithValue gives the option for a default value to be set
func WithValue(key, value string) string {
	s, ok := os.LookupEnv(key)
	if !ok {
		return value
	}
	return s
}
