package mailutils

import (
	"errors"
	"regexp"
	"strings"
)

// email regex
var (
	emailRegex    = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	ErrInvaidMail = errors.New("invalid email format")
)

// Regex : validate regex
func Regex(regex, sample string) error {
	c, err := regexp.Compile(regex)
	if err != nil {
		return err
	}
	if !c.MatchString(sample) {
		return errors.New("invalid input")
	}

	return nil
}

// Email validate
func Email(email string) bool {
	err := Regex(emailRegex, email)
	if err != nil {
		return false
	}

	return true
}

func IsDwarvesMail(mail string) bool {
	return strings.Contains(mail, "@dwarvesv.com") || strings.Contains(mail, "@d.foundation")
}
