package mailutils

import (
	"errors"
	"regexp"
	"strings"
)

// email regex
var (
	emailRegex        = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	teamEmailRegex    = ".+@((dwarvesv\\.com)|(d\\.foundation))"
	oldTeamEmailRegex = ".+@((dwarvesv\\.com))"
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
	return err == nil
}

func IsDwarvesMail(mail string) bool {
	regex, _ := regexp.Compile(teamEmailRegex)
	return regex.MatchString(mail)
}

func IsOldDwarvesMail(mail string) bool {
	regex, _ := regexp.Compile(oldTeamEmailRegex)
	return regex.MatchString(mail)
}

func TransformToNewDomainEmail(email string) string {
	if IsOldDwarvesMail(email) {
		mailParts := strings.Split(email, "@")
		return mailParts[0] + "@d.foundation"
	}

	return email
}
