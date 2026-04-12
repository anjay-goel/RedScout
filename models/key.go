package models

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const patternPlaceholder = "{id}"

type Key []string

func (k Key) Pop() (Key, error) {
	if len(k) == 0 {
		return nil, fmt.Errorf("key is empty")
	}

	if len(k) == 1 {
		return Key{}, nil // return empty key if only one part exists
	}

	newKey := make(Key, len(k)-1)
	copy(newKey, k[:len(k)-1])
	return newKey, nil
}

func (k Key) IsEmpty() bool {
	return len(k) == 0 || (len(k) == 1 && k[0] == "")
}

func (k Key) String() string {
	return strings.Join(k, ":")
}

type KeyParser struct {
	delimiter  string
	idPatterns []*regexp.Regexp
}

func NewKeyParser(Delimiter string, IDPattern []*regexp.Regexp) *KeyParser {
	return &KeyParser{
		delimiter:  Delimiter,
		idPatterns: IDPattern,
	}
}

// IsID determines if a key segment looks like an identifier using heuristics.
//
// Rules (staged by length — longer segments need weaker signals):
//
//	UUID (36 chars, 8-4-4-4-12 hex)       → always ID
//	Pure numeric (any length)              → always ID
//	>= 6 chars with >= 50% digits         → ID
//	8-11 chars with mixed case + digits    → ID
//	8-11 chars with >= 40% digits          → ID
//	12-19 chars with mixed case + digits   → ID
//	12-19 chars with >= 30% digits         → ID
//	20+ chars with any digit               → ID
func IsID(segment string) bool {
	if len(segment) == 36 && segment[8] == '-' && segment[13] == '-' && segment[18] == '-' && segment[23] == '-' {
		return true
	}

	upper := 0
	lower := 0
	digits := 0
	for _, r := range segment {
		switch {
		case unicode.IsUpper(r):
			upper++
		case unicode.IsLower(r):
			lower++
		case unicode.IsDigit(r):
			digits++
		}
	}

	if digits > 0 && digits == len(segment) {
		return true
	}

	total := upper + lower + digits
	if total == 0 {
		return false
	}

	hasMixedCase := upper > 0 && lower > 0

	// 20+ chars: any digit OR mixed case → ID
	if len(segment) >= 20 {
		return digits > 0 || hasMixedCase
	}

	// 16-19 chars: mixed case alone is suspicious enough
	// e.g., "SHOgPULACeRqGfhcBNV" — clearly not a word
	if len(segment) >= 16 && hasMixedCase {
		return true
	}

	// Below here, require digits
	if digits == 0 {
		return false
	}

	digitRatio := float64(digits) / float64(total)

	// 6+ chars with >= 50% digits
	if len(segment) >= 6 && digitRatio >= 0.5 {
		return true
	}

	if len(segment) < 8 {
		return false
	}

	// 12-15 chars: mixed case + digits, or >= 30% digits
	if len(segment) >= 12 {
		return (hasMixedCase && digits > 0) || digitRatio >= 0.3
	}

	// 8-11 chars: mixed case + digits, or >= 40% digits
	return (hasMixedCase && digits > 0) || digitRatio >= 0.4
}

func (kp *KeyParser) matchesPattern(part string) bool {
	if len(kp.idPatterns) > 0 {
		for _, regex := range kp.idPatterns {
			if regex.MatchString(part) {
				return true
			}
		}
		return false
	}
	return IsID(part)
}

func (kp *KeyParser) NewKey(s string, inferIds bool) Key {
	if !strings.Contains(s, kp.delimiter) {
		return []string{s}
	}

	parts := strings.Split(s, kp.delimiter)
	if inferIds {
		for i, part := range parts {
			if kp.matchesPattern(part) {
				parts[i] = patternPlaceholder
			}
		}
	}

	return parts
}

func (kp *KeyParser) IsA(k Key, prefix Key) bool {
	if len(prefix) > len(k) {
		return false
	}

	for i := 0; i < len(prefix); i++ {
		if prefix[i] == patternPlaceholder {
			if !kp.matchesPattern(k[i]) {
				return false
			}
		} else if k[i] != prefix[i] {
			return false
		}
	}

	return true
}

func (kp *KeyParser) Namespace(k Key, prefix Key, inferIds bool) (string, error) {
	if len(k) == 0 {
		return "", fmt.Errorf("key is empty")
	}

	if len(prefix) == 0 {
		return k[0], nil
	}

	if !kp.IsA(k, prefix) {
		return "", fmt.Errorf("key %s is not a child of prefix %s", strings.Join(k, ":"), strings.Join(prefix, ":"))
	}

	if len(k) == len(prefix) {
		return "", fmt.Errorf("key %s is exactly the same as prefix %s", strings.Join(k, ":"), strings.Join(prefix, ":"))
	}

	namespace := k[len(prefix)]

	if inferIds && kp.matchesPattern(namespace) {
		namespace = patternPlaceholder
	}
	return namespace, nil
}

func (kp *KeyParser) Append(k Key, part string, inferIds bool) (Key, error) {
	if part == "" {
		return k, fmt.Errorf("key is empty")
	}
	if inferIds && kp.matchesPattern(part) {
		part = patternPlaceholder
	}

	if len(k) == 0 {
		return Key{part}, nil
	}

	newKey := make(Key, len(k)+1)
	copy(newKey, k)
	newKey[len(k)] = part
	return newKey, nil
}
