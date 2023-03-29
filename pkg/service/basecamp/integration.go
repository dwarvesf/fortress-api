package basecamp

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/thoas/go-funk"
	"gorm.io/datatypes"
)

// BasecampExpenseData --
type BasecampExpenseData struct {
	Reason          string
	Amount          int
	CurrencyType    string
	CreatorEmail    string
	InvoiceImageURL string
	MetaData        datatypes.JSON
	BasecampID      int
}

const (
	defaultCurrencyType = "VND"
	thousandUnit        = 1000
	millionUnit         = 1000000
	amountPat           = "(\\d+(k|tr|m)\\d+|\\d+(k|tr|m)|\\d+)"
)

func getAmountStr(s string) string {
	c, _ := regexp.Compile(amountPat)
	return c.FindString(s)
}

// func getReason(s string) string {
// 	amount := getAmountStr(s)
// 	s = strings.Replace(s, amount, "", 1)
// 	return strings.TrimSpace(strings.Replace(s, "for", "", 1))
// }

// ExtractBasecampExpenseAmount --
func ExtractBasecampExpenseAmount(source string) int {
	return getAmount(strings.Replace(source, ".", "", -1))
}

func getAmount(source string) int {
	s := getAmountStr(source)
	if len(s) == 0 {
		return 0
	}

	switch {
	case isThousand(s):
		return thousand(s)
	case isMillion(s):
		return million(s)
	default:
		a, _ := strconv.Atoi(s)
		return a
	}
}

func isThousand(s string) bool {
	return funk.Contains(s, "k")
}

func thousand(s string) int {
	a := strings.Index(s, "k")
	if len(s[a+1:]) > 3 {
		return 0
	}
	prefix, _ := strconv.Atoi(s[0:a])
	suffix, _ := strconv.Atoi(s[a+1:])
	return prefix*thousandUnit + int(float64(suffix)/math.Pow10(len(s[a+1:])-1)*100)
}

func isMillion(s string) bool {
	return funk.Contains(s, "tr") || funk.Contains(s, "m")
}

func million(s string) int {
	newStr := strings.Replace(s, "tr", "m", -1)
	i := strings.Index(newStr, "m")
	if len(newStr[i+1:]) > 6 {
		return 0
	}
	pref, _ := strconv.Atoi(newStr[0:i])
	suf, _ := strconv.Atoi(newStr[i+1:])
	return (pref * millionUnit) + int(float64(suf)/math.Pow10(len(newStr[i+1:])-1)*thousandUnit*100)
}
