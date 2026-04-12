package scanner

import (
	"redscout/models"
	"strings"
)

var candidateDelimiters = []string{":", ".", "/", "-", "_", "#", "|"}

var idPlaceholder = models.PatternPlaceholder

// InferKeyTemplate takes a raw key string and returns a parsed Key
// with IDs replaced by {id} and consecutive namespace segments merged.
//
// It tries delimiters in priority order, splits the key, classifies each
// segment as ID or namespace, then merges consecutive namespace parts.
//
// Examples:
//
//	"user:abc123def:messages" → ["user", "{id}", "messages"]
//	"Yvu0y2ZE-mandate-registration-result" → ["{id}", "mandate-registration-result"]
//	"cache.getUser.xK92mNpQr7abc.result" → ["cache.getUser", "{id}", "result"]
//	"reel-user-property-cache:cache:abc123def:PREFERRED_LANGUAGE" → ["reel-user-property-cache", "cache", "{id}", "PREFERRED_LANGUAGE"]
func InferKeyTemplate(key string) ([]string, string) {
	// Try delimiters in priority order
	for idx, delim := range candidateDelimiters {
		if !strings.Contains(key, delim) {
			continue
		}

		parts := strings.Split(key, delim)
		if len(parts) < 2 {
			continue
		}

		// Classify each segment
		hasID := false
		classified := make([]struct {
			value string
			isID  bool
		}, len(parts))

		for i, part := range parts {
			id := models.IsID(part)
			if id {
				hasID = true
				classified[i] = struct {
					value string
					isID  bool
				}{idPlaceholder, true}
			} else {
				classified[i] = struct {
					value string
					isID  bool
				}{part, false}
			}
		}

		// For high-priority delimiters (: . /), accept even without IDs
		// For low-priority delimiters (- _ # |), require at least one ID
		isPrimaryDelim := idx < 3
		if !hasID && !isPrimaryDelim {
			continue
		}

		// If no IDs found, just return the split parts as-is
		if !hasID {
			return parts, delim
		}

		// Merge consecutive namespace segments
		var merged []string
		i := 0
		for i < len(classified) {
			if classified[i].isID {
				merged = append(merged, idPlaceholder)
				i++
			} else {
				j := i
				for j < len(classified) && !classified[j].isID {
					j++
				}
				nsParts := make([]string, j-i)
				for k := i; k < j; k++ {
					nsParts[k-i] = classified[k].value
				}
				merged = append(merged, strings.Join(nsParts, delim))
				i = j
			}
		}

		return merged, delim
	}

	// No delimiter matched
	if models.IsID(key) {
		return []string{idPlaceholder}, ""
	}
	return []string{key}, ""
}
