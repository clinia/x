package stringsx

import (
	"strings"
)

func SingleLine(str string) string {
	str = strings.ReplaceAll(str, "\n", "")
	return strings.ReplaceAll(str, "\t", "")
}
