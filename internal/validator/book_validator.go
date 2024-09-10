package validator

import "regexp"

// ISBNRegex is a regular expression for validating ISBN numbers
var ISBNRegex = regexp.MustCompile("^[0-9]{13}$")

// ValidateBook validates the book form
func ValidateBook(v *Validator, title string, publishYear int, calendarTime string, isbn string, source string) {
    ValidateTitle(v, title)
    ValidatePublishYear(v, publishYear)
    ValidateCalendarTime(v, calendarTime)
    ValidateISBN(v, isbn)
    ValidateSource(v, source)
}

// PermittedInt checks if a value is within a range
func PermittedInt(value int, min, max int) bool {
    return value >= min && value <= max
}

// PermittedValues checks if a value is in a list of permitted values
func PermittedValues(value string, permittedValues ...string) bool {
    for _, permittedValue := range permittedValues {
        if value == permittedValue {
            return true
        }
    }
    return false
}

// ValidateTitle validates the book's title
func ValidateTitle(v *Validator, title string) {
    v.CheckField(NotBlank(title), "title", "The title field cannot be blank")
    v.CheckField(MaxChars(title, 200), "title", "The title field cannot be more than 200 characters long")
    v.CheckField(NoInvalidCharacters(title), "title", "The title field contains invalid characters")
}

// ValidatePublishYear validates the book's publish year
func ValidatePublishYear(v *Validator, publishYear int) {
    v.CheckField(PermittedInt(publishYear, 1, 3000), "publish_year", "This field must be between 1 and 9999")
}

// ValidateCalendarTime validates the book's calendar time
func ValidateCalendarTime(v *Validator, calendarTime string) {
    v.CheckField(PermittedValues(calendarTime, "A.D.", "B.C."), "calendar_time", "This field must be either A.D. or B.C.")
}

// ValidateISBN validates the book's ISBN
func ValidateISBN(v *Validator, isbn string) {
    v.CheckField(NotBlank(isbn), "isbn", "The ISBN field cannot be blank")
    v.CheckField(Matches(isbn, ISBNRegex), "isbn", "This field must be a valid ISBN")
}

// ValidateSource validates the book's source
func ValidateSource(v *Validator, source string) {
    v.CheckField(MaxChars(source, 500), "source", "The source field cannot be more than 500 characters long")
    v.CheckField(NoInvalidCharacters(source), "source", "The source field contains invalid characters")
}