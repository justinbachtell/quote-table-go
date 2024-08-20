package validator

import "unicode"

// Checks if the quote meets all the required criteria
func ValidateQuote(v *Validator, quote string) {
    v.CheckField(NotBlank(quote), "quote", "The quote field cannot be blank.")
    v.CheckField(MaxChars(quote, 19000), "quote", "The quote field is too long (max. 19,000 characters).")
    v.CheckField(NoInvalidCharacters(quote), "quote", "The quote field contains invalid characters.")
}

// Checks if the author meets all the required criteria
func ValidateAuthor(v *Validator, author string) {
    v.CheckField(NotBlank(author), "author", "The author field cannot be blank.")
    v.CheckField(MaxChars(author, 100), "author", "The author field is too long (max. 100 characters).")
    v.CheckField(NoInvalidCharacters(author), "author", "The author field contains invalid characters.")
}

// Checks if the input string contains only valid characters
func ValidateCharacters(s string) bool {
    for _, r := range s {
        if !unicode.IsLetter(r) && !unicode.IsNumber(r) && !unicode.IsPunct(r) && !unicode.IsSpace(r) {
            return false
        }
    }
    return true
}