package utils

import "strings"

func RemoveEmptyString(in []string) []string {
	out := make([]string, 0)
	for _, status := range in {
		if RemoveAllSpace(status) != "" {
			out = append(out, status)
		}
	}

	return out
}

func RemoveAllSpace(str string) string  {
	return strings.ReplaceAll(str, " ", "")
}
