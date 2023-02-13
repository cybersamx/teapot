package common

import (
	"fmt"
	"net/url"
	"strings"
)

func RuneToString(r rune) string {
	if r == rune(0) {
		return ""
	}

	return string(r)
}

// MaskPassword masks the password component of an input.
func MaskPassword(text string) string {
	u, err := url.Parse(text)
	if err != nil {
		return text
	}

	if u.User != nil {
		pwd, set := u.User.Password()
		if set {
			return strings.Replace(u.String(), fmt.Sprintf("%s@", pwd), "***@", 1)
		}
	}

	return u.String()
}
