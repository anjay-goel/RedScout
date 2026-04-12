package models

// KeyValueInfo holds fetched key details for the value viewer
type KeyValueInfo struct {
	Key    string
	Type   string
	Value  string // string value or summary for non-string types
	Size   int64  // memory usage in bytes
	TTL    int64  // TTL in seconds, -1 = no expiry, -2 = key not found
	Length int64  // element count for lists/sets/hashes/zsets, string length for strings
}
