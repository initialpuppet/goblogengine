// Package slug provides functionality for sanitising strings so that they can
// be used in URIs.
//
// Original taken from https://github.com/russross/blackfriday.
package slug

import (
	"unicode"
)

// Make takes a string and converts it to a lower-case, hyphen-separated,
// alphanumeric string for use in URIs.
func Make(text string) string {
	return string(slugify([]byte(text)))
}

// Validate returns true if the supplied string is a lower-case, hyphen-
// separated, alphanumeric string.
// TODO
func Valid(text string) bool {
	return true
}

// isletter returns true if a character is a Latin letter
func isletter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isalnum returns true if a character is a Latin letter or a European digit
func isalnum(c byte) bool {
	return (c >= '0' && c <= '9') || isletter(c)
}

// slugify creates a url-safe slug
func slugify(in []byte) []byte {
	if len(in) == 0 {
		return in
	}
	out := make([]byte, 0, len(in))
	sym := false

	for _, ch := range in {
		if isalnum(ch) {
			sym = false
			// one byte == one rune here thanks to isalnum
			lch := byte(unicode.ToLower(rune(ch)))
			out = append(out, lch)
		} else if sym {
			continue
		} else {
			out = append(out, '-')
			sym = true
		}
	}
	var a, b int
	var ch byte
	for a, ch = range out {
		if ch != '-' {
			break
		}
	}
	for b = len(out) - 1; b > 0; b-- {
		if out[b] != '-' {
			break
		}
	}
	return out[a : b+1]
}
