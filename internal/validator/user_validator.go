package validator

import "regexp"

// EmailRX is a regular expression for validating email addresses
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// PhoneRX is a regular expression for validating phone numbers with dashes
var PhoneRX = regexp.MustCompile("^[0-9]{3}-[0-9]{3}-[0-9]{4}$")

// ValidateSignupForm validates the user signup form
func ValidateSignupForm(v *Validator, name, email, password string) {
    ValidateName(v, name)
    ValidateEmail(v, email)
    ValidatePassword(v, password)
}

// ValidateLoginForm validates the user login form
func ValidateLoginForm(v *Validator, email, password string) {
    ValidateEmail(v, email)
    ValidatePassword(v, password)
}

// ValidateProfileForm validates the user profile form
func ValidateProfileForm(v *Validator, name, email, phone string) {
    ValidateName(v, name)
    ValidateEmail(v, email)
    ValidatePhone(v, phone)
}

// ValidateChangePasswordForm validates the change password form
func ValidateChangePasswordForm(v *Validator, currentPassword, newPassword, confirmPassword string) {
    v.CheckField(NotBlank(currentPassword), "currentPassword", "This field cannot be blank")
    v.CheckField(NotBlank(newPassword), "newPassword", "This field cannot be blank")
    v.CheckField(MinChars(newPassword, 8), "newPassword", "This field must be at least 8 characters long")
    v.CheckField(NotBlank(confirmPassword), "confirmPassword", "This field cannot be blank")
    v.CheckField(newPassword == confirmPassword, "confirmPassword", "Passwords do not match")
}

// ValidateName validates the user's name
func ValidateName(v *Validator, name string) {
    v.CheckField(NotBlank(name), "name", "The name field cannot be blank.")
    v.CheckField(MaxChars(name, 100), "name", "The name field is too long (max. 100 characters).")
    v.CheckField(MinChars(name, 2), "name", "The name field is too short (min. 2 characters).")
    v.CheckField(NoInvalidCharacters(name), "name", "The name field contains invalid characters.")
}

// ValidateEmail validates the user's email
func ValidateEmail(v *Validator, email string) {
    v.CheckField(NotBlank(email), "email", "The email field cannot be blank.")
    v.CheckField(MaxChars(email, 255), "email", "The email field is too long (max. 255 characters).")
    v.CheckField(MinChars(email, 5), "email", "The email field is too short (min. 5 characters).")
    v.CheckField(Matches(email, EmailRX), "email", "The email field is not a valid email address.")
}

// ValidatePhone validates the user's phone
func ValidatePhone(v *Validator, phone string) {
    v.CheckField(NotBlank(phone), "phone", "The phone field cannot be blank.")
    v.CheckField(MaxChars(phone, 15), "phone", "The phone field is too long (max. 15 characters).")
    v.CheckField(MinChars(phone, 10), "phone", "The phone field is too short (min. 10 characters).")
    v.CheckField(Matches(phone, PhoneRX), "phone", "The phone field is not a valid phone number.")
}

// ValidatePassword validates the user's password
func ValidatePassword(v *Validator, password string) {
    v.CheckField(NotBlank(password), "password", "The password field cannot be blank.")
    v.CheckField(MinChars(password, 8), "password", "The password field is too short (min. 8 characters).")
    v.CheckField(MaxChars(password, 70), "password", "The password field is too long (max. 70 characters).")
    v.CheckField(NoInvalidCharacters(password), "password", "The password field contains invalid characters.")
}
