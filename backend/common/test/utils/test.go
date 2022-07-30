package utils

import (
	"os"
	"testing"
)

func MustSkip(t *testing.T, msg string) {
	t.Helper()
	if os.Getenv(envNoSkip) != "" {
		panic("test environment disallows a test to be skipped")
	}
	t.Skip(msg)
}
