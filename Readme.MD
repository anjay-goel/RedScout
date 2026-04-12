# RedScout - Redis Namespace Monitor

![Screenshot](assets/screenshot.png)

RedScout is a Redis monitoring tool that provides namespace-level insights into your Redis database unlike
traditional tools that focus more on server-level metrics. It breaks down memory usage, key TTLs, and OP/s by logical
namespaces, giving you clear visibility into which parts of your application are driving Redis usage. RedScout can also
handle deeply nested key patterns like `user:<user-id>:messages:unread` or `user:<user-id>:order:<order-id>:status`,
helping you pinpoint hotspots with precision.

## How It Works

RedScout leverages three core Redis commands: `SCAN`, `MEMORY USAGE`, and `MONITOR` to assemble a namespace-level
view of your cache. On top of these, it accepts configurable inputs such as a custom delimiter for splitting key
hierarchies and regular expressions to automatically identify and infer id components within your keys.

## Requirements

- **Redis Version**: 4.0.0 or higher (required for `MEMORY USAGE` command)

## Installation

Download the appropriate binary for your platform from the releases page.

## Usage

```bash
./redscout \
  -h redis.example.com \
  -p 6380 \
  -a "your-password" \
  --tls \
  --monitor-duration 30 \
  --scan-size 10000
```

## Configuration Options

### Connection Settings

| Flag    | Type   | Default   | Description                                              |
|---------|--------|-----------|----------------------------------------------------------|
| `-h`    | string | `localhost` | Redis host address                                       |
| `-p`    | int    | `6379`    | Redis port number                                        |
| `-u`    | string | `default` | Redis username for authentication                        |
| `-a`    | string | _(empty)_ | Redis password for authentication                        |
| `-n`    | int    | `0`       | Redis database number to analyze                         |
| `--tls` | bool   | `false`   | Use TLS for Redis connection (presence enables TLS)      |

### Analysis Settings

| Flag                   | Type   | Default   | Description                                                   |
|------------------------|--------|-----------|---------------------------------------------------------------|
| `--scan-size`          | int    | `5000`    | Number of keys to scan                                        |
| `--monitor-duration`   | int    | `10`      | Duration in seconds to run the `monitor` command              |
| `--refresh-interval`   | int    | `5`       | Interval in seconds between Redis info refreshes              |

### Key Parsing

By default, RedScout splits keys on `:` and uses built-in patterns to detect IDs (UUIDs, Firebase UIDs, hex hashes, etc.). For advanced use cases, you can customize this or switch to auto-inference.

| Flag             | Type   | Default   | Description                                                   |
|------------------|--------|-----------|---------------------------------------------------------------|
| `--delimiter`    | string | `:`       | Delimiter for splitting keys (manual mode)                    |
| `--id-regex`     | string | _defaults_| Space-separated regex patterns to detect IDs (overrides defaults) |
| `--infer-keys`   | bool   | `false`   | Auto-infer key structure using heuristics                     |

**Manual mode** (default): Splits on `--delimiter`. No ID detection unless `--id-regex` is provided. When `--id-regex` is given, matching segments are collapsed into `{id}` groups.

**Auto-infer mode** (`--infer-keys`): Tries multiple delimiters per key (`: . / - _`), classifies each segment as ID or namespace by its character pattern, and merges consecutive namespace parts. Zero config needed — handles mixed key patterns in the same Redis instance.

### Output Settings

| Flag          | Type   | Default         | Description                                |
|---------------|--------|-----------------|--------------------------------------------|
| `--logs-dir`  | string | _OS's temp dir_ | Directory to store temporary analysis logs |

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `N` | Switch to Namespaces tab |
| `L` | Switch to Slow Log tab |
| `B` | Switch to Big Keys tab |
| `H` | Switch to Hot Keys tab |
| `T` | Cycle to next tab |
| `S` | Run additional SCAN (accumulates data) |
| `M` | Run additional MONITOR (accumulates data) |
| `1-8` | Sort current table by column |
| `→` / `Enter` | Drill down into namespace |
| `←` / `Backspace` | Navigate back up |
| `?` | Show/hide keyboard shortcut overlay |
| `Q` | Quit |

## Notes

- Uses Redis [MONITOR](https://redis.io/docs/latest/commands/monitor/) command, so be careful when using in production
  environments
- Results are estimates based on sampling, not exhaustive key scanning
