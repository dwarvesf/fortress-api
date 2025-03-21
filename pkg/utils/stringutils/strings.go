package stringutils

import (
	"math/big"
	"regexp"
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

// ConvertToFullDecimal Convert a float value to its full decimal representation with specified decimal places
func ConvertToFullDecimal(value float64, decimalPlaces int) string {
	// Handle special cases
	if value == 0 {
		return "0"
	}

	// Create a high-precision big.Float to hold the original value
	valueAsBigFloat := new(big.Float).SetPrec(256).SetFloat64(value)

	// Create the multiplier as a big.Int (10^decimalPlaces)
	multiplier := new(big.Int).Exp(
		big.NewInt(10),
		big.NewInt(int64(decimalPlaces)),
		nil,
	)

	// Convert multiplier to big.Float for the multiplication
	multiplierAsBigFloat := new(big.Float).SetPrec(256).SetInt(multiplier)

	// Perform the multiplication with high precision
	result := new(big.Float).SetPrec(256).Mul(valueAsBigFloat, multiplierAsBigFloat)

	// Convert to big.Int (this will automatically round/truncate any remaining decimals)
	var resultAsBigInt big.Int
	result.Int(&resultAsBigInt)

	// Return the formatted string without scientific notation
	return resultAsBigInt.String()
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
