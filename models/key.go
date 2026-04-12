package models

import (
	"fmt"
	"regexp"
	"strings"
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

func (kp *KeyParser) matchesPattern(part string) bool {
	for _, regex := range kp.idPatterns {
		if regex.MatchString(part) {
			return true
		}
	}
	return false
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
		if prefix[i] == PatternPlaceholder || k[i] == PatternPlaceholder {
			continue
		}
		if k[i] != prefix[i] {
			return false
		}
	}

	return true
}

// Namespace extracts the next namespace segment from a key relative to a prefix.
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

	return k[len(prefix)], nil
}

// Append adds a segment to a key.
func (kp *KeyParser) Append(k Key, part string, inferIds bool) (Key, error) {
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
