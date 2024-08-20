package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// Define a Validator type which holds a map of field names and their error messages
type Validator struct {
	NonFieldErrors []string
	FieldErrors map[string]string
	JSONErrors map[string]string
}

// Returns a new Validator instance
func New() *Validator {
	return &Validator{JSONErrors: make(map[string]string)}
}

// Returns true if the JSONErrors map is empty
func (v *Validator) ValidJSON() bool {
	return len(v.JSONErrors) == 0
}

// Returns true if the NonFieldErrors slice is empty
func (v *Validator) ValidNonField() bool {
	return len(v.NonFieldErrors) == 0
}

// Returns true if the FieldErrors map is empty
func (v *Validator) ValidField() bool {
	return len(v.FieldErrors) == 0
}

// Adds an error message to the JSONErrors map
func (v *Validator) AddJSONError(key, message string) {
	if _, exists := v.JSONErrors[key]; !exists {
		v.JSONErrors[key] = message
	}
}

// Adds an error message to the JSONErrors map if validation check is not ok
func (v *Validator) CheckJSON(ok bool, key, message string) {
	if !ok {
		v.AddJSONError(key, message)
	}
}

// Adds an error message to the NonFieldErrors slice
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// Adds an error message to the FieldErrors map if map doesn't exist
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}
	
	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// Adds an error message to the FieldErrors map if validation check is not ok
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// Returns true if a value is not an empty string
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// Returns true if a value is more than n characters
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// Returns true if a value contains no more than n characters
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// Returns true if a value is in a list of specified values
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// Returns true if a slice contains no duplicate values
func UniqueValue[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, value := range values {
		if _, exists := uniqueValues[value]; exists {
			return false
		}
		uniqueValues[value] = true
	}
	return true
}

// Returns true if a value contains no invalid characters
func NoInvalidCharacters(value string) bool {
	// Define an array of invalid characters
	var invalidChars = []rune{'[', ']', '{', '}', '\\', '|', '/', '+', '<', '>', '~'}

	for _, char := range value {
		for _, invalidChar := range invalidChars {
			if char == invalidChar {
				return false
			}
		}
	}

	return true
}

// Returns true if a value matches a regular expression
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// TODO: Regex for password validation


