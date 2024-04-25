package stringutils

import (
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
