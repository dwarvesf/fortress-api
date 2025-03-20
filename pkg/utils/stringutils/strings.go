package stringutils

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/constant"
)

func ExtractPattern(str string, pattern string) []string {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(str, -1)
	var result []string
	for _, match := range matches {
		result = append(result, match[1])
	}

	return result
}

func ExtractEmailPattern(str string) []string {
	re := regexp.MustCompile(constant.RegexPatternEmail)
	matches := re.FindAllStringSubmatch(str, -1)
	var result []string
	for _, match := range matches {
		result = append(result, match[0])
	}

	return result
}

func ExtractNumber(str string) []string {
	re := regexp.MustCompile(constant.RegexPatternNumber)
	matches := re.FindAllStringSubmatch(str, -1)
	var result []string
	for _, match := range matches {
		result = append(result, match[0])
	}

	return result
}

func FormatString(str string) string {
	// Replace spaces with a single space
	re := regexp.MustCompile(`\s+`)
	formattedStr := re.ReplaceAllString(str, " ")

	// Remove spaces after the "#" symbol
	formattedStr = strings.ReplaceAll(formattedStr, "# ", "#")

	return formattedStr
}

// FloatToString convert float to big int string with given decimal
// Ignore negative float
// example: FloatToString("0.000000000000000001", 18) => "1"
func FloatToString(s string, decimal int64) string {
	c, _ := strconv.ParseFloat(s, 64)
	if c < 0 {
		return "0"
	}
	bigval := new(big.Float)
	bigval.SetFloat64(c)

	d := new(big.Float)
	d.SetInt(big.NewInt(int64(math.Pow(10, float64(decimal)))))
	bigval.Mul(bigval, d)

	r := new(big.Int)
	bigval.Int(r) // store converted number in r
	return r.String()
}

func Shorten(s string) string {
	// avoid oor error
	if len(s) < 12 {
		return s
	}

	// already shortened -> return
	if strings.Contains(s, "..") {
		return s
	}

	// shorten
	// e.g. Shorten("0x7dff46370e9ea5f0bad3c4e29711ad50062ea7a4) = "0x7d..a7a4"
	return s[:5] + ".." + s[len(s)-5:]
}
