package basecamp

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

const (
	defaultCurrencyType = "VND"
	thousandUnit        = 1000
	millionUnit         = 1000000
	amountPat           = "(\\d+(k|tr|m)\\d+|\\d+(k|tr|m)|\\d+)"
)

func ExtractBasecampExpenseAmount(source string) int {
	// TODO: has not support decimal part (e.g. 10.5 USD) yet
	return getAmount(strings.Replace(source, ".", "", -1))
}

func getAmountStr(s string) string {
	c, _ := regexp.Compile(amountPat)
	return c.FindString(s)
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
	return containsAny(s, "k")
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
	return containsAny(s, "tr", "m")
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

func BuildFailedComment(s string) *model.Comment {
	return &model.Comment{Content: fmt.Sprintf(`<img width="17" class="thread-entry__icon" src="https://3.basecamp-static.com/assets/icons/thread_events/uncompleted-6066b80e80b6463243d7773fa67373b62e2a7d159ba12a17c94b1e18b30a5770.svg"><div><em>%s</em></div>`, s)}
}

func BuildCompletedComment(s string) *model.Comment {
	return &model.Comment{Content: fmt.Sprintf(`<img width="17" class="thread-entry__icon" src="https://3.basecamp-static.com/assets/icons/thread_events/completed-12705cf5fc372d800bba74c8133d705dc43a12c939a8477099749e2ef056e739.svg"><div><em>%s</em></div>`, s)}
}

func BasecampMention(sgID string) string {
	return fmt.Sprintf(`<bc-attachment sgid="%s" content-type="application/vnd.basecamp.mention"></bc-attachment>`, sgID)
}

func containsAny(s string, sub ...string) bool {
	for i := range sub {
		if strings.Contains(s, sub[i]) {
			return true
		}
	}
	return false
}
