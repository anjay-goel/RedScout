# RedScout - Development Guide

## Project Overview

Redis namespace-level monitoring TUI built with Go + tview. Scans keys, monitors operations, and provides breakdown by namespace hierarchy.

## Tech Stack

- **Language:** Go
- **TUI Framework:** tview (github.com/rivo/tview) + tcell
- **Redis Client:** go-redis/v9
- **Styling:** GitHub Dark theme via tcell RGB colors (lib/ui/theme/)

## Project Structure

```
lib/
  cli.go                    # CLI flag parsing
  constants.go              # Scan batch sizes
  redis.go                  # Redis client factory
  scanner/
    scanner.go              # Scanner lifecycle (Start, Close)
    commands.go             # Redis commands (ScanMemory, MonitorOps, FetchRedisInfo)
    analytics.go            # Data processing (namespace stats, big/hot keys)
  ui/
    ui.go                   # Root AppUI, screen management, input handling
    theme/theme.go          # GitHub Dark color constants (tcell)
    views/
      titlebar.go           # Title bar (branding + scan stats + help)
      header.go             # 3-panel header (System, Performance, Resources)
      body.go               # Tab bar + content area management
      help_overlay.go       # ? help modal
      components/
        namespace_table.go  # Namespace breakdown table
        bigkeys_table.go    # Big keys table (Key, Size, Type, TTL)
        hotkeys_table.go    # Hot keys table (Key, Ops/s, Command)
        slowlog_table.go    # Slow log table
        progress_bar.go     # Progress bar renderer
models/
  state.go                  # Application state + update channel
  config.go                 # CLI config struct
  redis_info.go             # Redis INFO parser
  namespace.go              # Namespace metrics + estimation
  ops_map.go                # Redis command → GET/SET/DEL categorization
  special_keys.go           # BigKey, HotKey structs + min-heaps
  key.go                    # Key parsing, namespace extraction
  slowlog_list.go           # SlowLog wrapper
```

## Build & Run

```bash
go build ./...
go run main.go -h localhost -p 6379
go test ./...
```

## Key Architecture Decisions

- **Scanner ↔ UI communication:** Scanner sends state updates via `models.State.Updates` channel. UI listens in `stateUpdateListener` goroutine.
- **Namespace estimation:** Keys are sampled (not full scan), metrics extrapolated based on sample ratio.
- **Multi-key commands:** MGET/MSET/DEL etc. are split into individual key operations in the monitor parser.
- **Ops categorization:** All Redis commands mapped to GET/SET/DEL/EVAL in `models/ops_map.go`.
- **Theme:** All colors defined as tcell RGB vars in `lib/ui/theme/theme.go`. UI uses tview dynamic color tags (`[#hex]`).

## Conventions

- tview color markup: `[#rrggbb]text[-]` for colored text, `[#rrggbb::b]text[-::-]` for bold
- Shortcut hints use orange (`#f0883e`) in square brackets: `[N]`, `[1]`, `[?]`
- Tables use fixed-width key columns (`utils.MaxKeyDisplayLen`) to prevent layout shift
- Alternating row backgrounds: `ColorBg` / `ColorSurface`
