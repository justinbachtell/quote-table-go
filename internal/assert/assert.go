package assert

import (
	"strings"
	"testing"
)

// Check if the actual value is equal to the expected value
func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper()

	if actual != expected {
		t.Errorf("got %v, want %v", actual, expected)
	}
}

// Check if the actual string contains the expected substring
func StringContains(t *testing.T, actual, expectedSubstring string) {
	t.Helper()

	if !strings.Contains(actual, expectedSubstring) {
		t.Errorf("got %s, want %s", actual, expectedSubstring)
	}
}

// Check if the actual error is nil
func NilError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("got : %v; expected: nil", actual)
	}
}