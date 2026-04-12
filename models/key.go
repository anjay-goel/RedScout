package models

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const PatternPlaceholder = "{id}"

type Key []string

func (k Key) Pop() (Key, error) {
	if len(k) == 0 {
		return nil, fmt.Errorf("key is empty")
	}

	if len(k) == 1 {
		return Key{}, nil
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

// KeyInferFunc parses a raw key string into segments with IDs replaced
// by {id} and consecutive namespaces merged.
type KeyInferFunc func(key string) ([]string, string)

type KeyParser struct {
	delimiter  string
	idPatterns []*regexp.Regexp
	inferFunc  KeyInferFunc
}

func NewKeyParser(delimiter string, idPatterns []*regexp.Regexp) *KeyParser {
	return &KeyParser{
		delimiter:  delimiter,
		idPatterns: idPatterns,
	}
}

// SetInferFunc enables auto-inference mode. When set, NewKey with
// inferIds=true will use this function instead of delimiter+regex.
func (kp *KeyParser) SetInferFunc(fn KeyInferFunc) {
	kp.inferFunc = fn
}

// IsID determines if a key segment looks like an identifier using heuristics.
//
// Rules:
//
//	UUID (36 chars, 8-4-4-4-12 hex)          → always ID
//	Pure numeric (any length)                 → always ID
//	>= 6 chars with >= 50% digits            → ID
//	8-11 chars with mixed case + digits       → ID
//	8-11 chars with >= 40% digits             → ID
//	12-19 chars with mixed case + digits      → ID
//	12-19 chars with >= 30% digits            → ID
//	20+ chars with any digit                  → ID
//	< 6 chars or no digits                    → not ID
func IsID(segment string) bool {
	// UUID: 8-4-4-4-12 hex with dashes
	if len(segment) == 36 && segment[8] == '-' && segment[13] == '-' && segment[18] == '-' && segment[23] == '-' {
		return true
	}

	// Count character classes
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

	// Pure numeric → always ID (any length)
	if digits > 0 && digits == len(segment) {
		return true
	}

	total := upper + lower + digits
	if total == 0 || digits == 0 {
		return false
	}

	digitRatio := float64(digits) / float64(total)

	// >= 6 chars with >= 50% digits → ID (e.g., "a1b2c3", "abc123")
	if len(segment) >= 6 && digitRatio >= 0.5 {
		return true
	}

	if len(segment) < 8 {
		return false
	}

	hasMixedCase := upper > 0 && lower > 0

	// 20+ chars: any digit is enough
	if len(segment) >= 20 {
		return true
	}

	// 12-19 chars: mixed case + digits, or >= 30% digits
	if len(segment) >= 12 {
		return hasMixedCase || digitRatio >= 0.3
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

// NewKey parses a raw Redis key string into segments.
//
// When inferIds=true:
//   - Auto mode (inferFunc set): delegates to the heuristic function
//   - Manual mode: splits on delimiter, replaces regex-matched segments with {id}
//
// When inferIds=false: splits on delimiter only, no ID replacement.
func (kp *KeyParser) NewKey(s string, inferIds bool) Key {
	if inferIds && kp.inferFunc != nil {
		parts, _ := kp.inferFunc(s)
		return parts
	}

	if !strings.Contains(s, kp.delimiter) {
		return []string{s}
	}

	parts := strings.Split(s, kp.delimiter)
	if inferIds {
		for i, part := range parts {
			if kp.matchesPattern(part) {
				parts[i] = PatternPlaceholder
			}
		}
	}

	return parts
}

// IsA checks if key k matches the given prefix pattern.
// {id} in the prefix matches any value at that position.
func (kp *KeyParser) IsA(k Key, prefix Key) bool {
	if len(prefix) > len(k) {
		return false
	}

	for i := 0; i < len(prefix); i++ {
		if prefix[i] == PatternPlaceholder {
			continue // prefix {id} matches any key value at this position
		}
		if k[i] == PatternPlaceholder {
			// key has {id} at this position — only match if prefix also has {id}
			return false
		}
		if k[i] != prefix[i] {
			return false
		}
	}

	return true
}

// Namespace extracts the next namespace segment from a key relative to a prefix.
func (kp *KeyParser) Namespace(k Key, prefix Key) (string, error) {
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

	return k[len(prefix)], nil
}

// Append adds a segment to a key.
func (kp *KeyParser) Append(k Key, part string) (Key, error) {
	if part == "" {
		return k, fmt.Errorf("key is empty")
	}

	if len(k) == 0 {
		return Key{part}, nil
	}

	newKey := make(Key, len(k)+1)
	copy(newKey, k)
	newKey[len(k)] = part
	return newKey, nil
}
