# TODO

## Features

- [x] Show key value on leaf drill-down — when expanding into a namespace that has no sub-namespaces, show the actual key value instead of an empty table. Needs per-type fetch (GET for strings, LRANGE for lists, HGETALL for hashes, etc.), value truncation for large keys, and consider security implications (PII/tokens).
- [ ] Refresh slow log on S/M key press
- [ ] Auto-infer IDs from key patterns — detect high-cardinality segments as IDs without requiring `--id-regex`. Needs: per-pattern grouping, cardinality analysis per position, handling mixed key structures and varying depths across different services/frameworks.
