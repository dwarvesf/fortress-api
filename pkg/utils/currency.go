package utils

import (
	"fmt"
)

func FormatCurrencyAmount(n int) string {
	milPart := n / 1000000
	thouPart := (n % 1000000) / 1000
	hunPart := n % 1000
	var res string
	if milPart != 0 {
		res = fmt.Sprintf("%d", milPart)
		if thouPart == 0 {
			res = fmt.Sprintf("%v,000", res)
		}
	}
	if thouPart != 0 {
		if res == "" {
			res = fmt.Sprintf("%d", thouPart)
		} else {
			res = fmt.Sprintf("%v,%03d", res, thouPart)
		}
	}
	if hunPart != 0 {
		if res == "" {
			res = fmt.Sprintf("%d", hunPart)
		} else {
			res = fmt.Sprintf("%v.%03d", res, hunPart)
		}
	}

	if milPart != 0 || thouPart != 0 {
		return fmt.Sprintf("%vk", res)
	}
	return res
}
