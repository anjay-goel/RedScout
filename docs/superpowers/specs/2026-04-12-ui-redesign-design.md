# RedScout UI Redesign — Design Spec

**Date:** 2026-04-12
**Status:** Approved
**Scope:** Full UI visual overhaul of the tview-based TUI

## Summary

Redesign RedScout's terminal UI from the current teal/cyan theme to a GitHub Dark aesthetic. Restructure the layout to
reduce vertical overhead, improve visual hierarchy, and provide contextual keyboard hints instead of a dedicated
shortcuts bar.

## Decisions

| Decision     | Choice                                 | Rationale                                               |
|--------------|----------------------------------------|---------------------------------------------------------|
| Theme        | GitHub Dark                            | Clean, modern, familiar to developers                   |
| Header       | 3 panels (System, Performance, Memory) | Scan state is tool state, not Redis info                |
| Scan state   | Title bar, right-aligned               | Compact, always visible, zero extra rows                |
| Tabs         | 4 tabs with underline active indicator | Keep separate: Namespaces, Slow Log, Big Keys, Hot Keys |
| Big/Hot Keys | Enriched columns                       | Add Type, TTL, Namespace, Command columns               |
| Shortcuts    | Contextual inline hints + ? overlay    | No dedicated bottom bar, hints where relevant           |

## Color System

GitHub Dark palette mapped to tcell RGB colors.

| Role           | Hex       | Go constant name | Usage                                  |
|----------------|-----------|------------------|----------------------------------------|
| Background     | `#0d1117` | `ColorBg`        | All backgrounds                        |
| Surface        | `#161b22` | `ColorSurface`   | Alternating rows, panel fill           |
| Border         | `#21262d` | `ColorBorder`    | Panel borders, dividers, table lines   |
| Text primary   | `#c9d1d9` | `ColorText`      | Data values, key names                 |
| Text muted     | `#484f58` | `ColorMuted`     | Labels, section titles, secondary info |
| Text secondary | `#8b949e` | `ColorSecondary` | Units, descriptions, column headers    |
| Accent blue    | `#58a6ff` | `ColorBlue`      | Active tab, branding, memory values    |
| Accent orange  | `#f0883e` | `ColorOrange`    | Key counts, hotkeys, sort hints        |
| Accent green   | `#3fb950` | `ColorGreen`     | Hit rate, status dot, positive states  |
| Accent red     | `#f85149` | `ColorRed`       | Errors, warnings, delete ops           |

All colors use `tcell.NewRGBColor(r, g, b)`. Define as constants in a new `lib/ui/theme.go` file.

## Layout Structure

```
Row 0:    [Title Bar]     RedScout (left)  |  scan state + ? help (right)
Row 1-3:  [3 Panels]      SYSTEM  |  PERFORMANCE  |  MEMORY
Row 4:    [Tab Bar]        Tabs (left)  |  S scan  M monitor (right)
Row 5+:   [Content]        Active table view (fills remaining space)
```

**Total header overhead:** ~5 rows (title + panels + tabs)
**Current overhead:** ~12 rows (4 panels + tabs + shortcuts)
**Savings:** ~7 extra data rows visible

### Title Bar (new component)

Single row, no border.

- **Left:** `RedScout` in `ColorBlue`, bold
- **Right:** `{scannedKeys} scanned · {monitorDuration} monitored · ? help` in `ColorMuted`, with `?` in `ColorOrange`

### Header Panels (3 panels, restyled)

Three equal-width panels in a horizontal flex. Each has:

- 1px border in `ColorBorder`, rounded corners (tview box drawing)
- Uppercase label in `ColorMuted` with letter-spacing (e.g., `SYSTEM`)
- Compact content (2-3 lines)

**System panel:**

```
SYSTEM
Redis v7.4.3 · Linux · 30d up · 121 clients
```

All on one content line. "Redis" in `ColorBlue` bold, rest in `ColorMuted`.

**Performance panel:**

```
PERFORMANCE
11.1M keys │ 12.9K/s │ 93.8% hit │ ttl 3h 3m
```

Key numbers (`11.1M`, `12.9K`) in `ColorOrange` bold. `93.8%` in `ColorGreen`. Units in `ColorSecondary`.

**Memory panel:**

```
MEMORY
████░░░░░░ 6.5G · no limit · noeviction
```

Progress bar: filled portion in `ColorOrange`, empty in `ColorBorder`. Value in `ColorOrange` bold. When maxmemory=0,
show `{used} · no limit`. When CPU not reported, show `cpu: n/a` in `ColorMuted`.

### Tab Bar (restyled)

Single row with bottom border in `ColorBorder`.

- **Left:** Tab names. Active tab: `ColorBlue` text with 2-char underline in `ColorBlue`. Inactive: `ColorSecondary`.
- **Right:** Contextual action hints. `S` and `M` in `ColorOrange`, labels in `ColorMuted`. Changes per active tab if
  needed.

### Content Area

Fills remaining vertical space. Contains the active table view.

**Alternating row colors:** Even rows on `ColorBg`, odd rows on `ColorSurface`.

**Column headers:** `ColorSecondary` text, uppercase. Sort key numbers in `ColorOrange` next to sortable columns (e.g.,
`~KEYS 1  ~MEMORY 2`).

## Table Specifications

### Namespace Table

**Breadcrumb row** (only shown for this tab):

- Left: `/ root › prefix › subprefix` in `ColorOrange`
- Right: `→ drill  ← back` hints, arrows in `ColorOrange`, labels in `ColorMuted`

**Columns:**

| Column    | Align | Width                         | Color            |
|-----------|-------|-------------------------------|------------------|
| Namespace | left  | `MaxKeyDisplayLen` (64) fixed | `ColorText`      |
| ~Keys     | right | 12                            | `ColorOrange`    |
| ~Memory   | right | 12                            | `ColorBlue`      |
| Avg TTL   | right | 12                            | `ColorSecondary` |
| % TTL     | right | 8                             | `ColorSecondary` |
| GET/s     | right | 10                            | `ColorSecondary` |
| SET/s     | right | 10                            | `ColorSecondary` |
| DEL/s     | right | 10                            | `ColorSecondary` |
| OPS/s     | right | 10                            | `ColorSecondary` |

**Removed:** Types column (low value, saves horizontal space).

**Sort hints in headers:** `~KEYS 1`, `~MEMORY 2`, `AVG TTL 3`, `% TTL 4`, `GET/s 5`, `SET/s 6`, `DEL/s 7`, `OPS/s 8`

### Big Keys Table (enriched)

| Column    | Align                           | Color            |
|-----------|---------------------------------|------------------|
| Key       | left (fixed `MaxKeyDisplayLen`) | `ColorText`      |
| Size      | right                           | `ColorOrange`    |
| Type      | right                           | `ColorSecondary` |
| TTL       | right                           | `ColorGreen`     |
| Namespace | right                           | `ColorMuted`     |

**Sort hints:** `SIZE 1`, `TYPE 2`, `TTL 3`

Data source: Already available in scan log (`key memory ttl type`). Namespace derived from key parser.

### Hot Keys Table (enriched)

| Column    | Align                           | Color            |
|-----------|---------------------------------|------------------|
| Key       | left (fixed `MaxKeyDisplayLen`) | `ColorText`      |
| Ops/s     | right                           | `ColorBlue`      |
| Command   | right                           | `ColorSecondary` |
| Namespace | right                           | `ColorMuted`     |

**Sort hints:** `OPS/s 1`, `CMD 2`

Data source: Command already tracked in monitor log. Namespace derived from key parser. Need to track top command per
key in `ComputeHotKeysFromMonitorLog`.

### Slow Log Table

| Column    | Align                                  | Color            |
|-----------|----------------------------------------|------------------|
| ID        | right                                  | `ColorText`      |
| Timestamp | left                                   | `ColorSecondary` |
| Duration  | right                                  | `ColorOrange`    |
| Command   | left                                   | `ColorGreen`     |
| Arguments | left (truncated to `MaxKeyDisplayLen`) | `ColorMuted`     |

**Sort hints:** `ID 1`, `TIME 2`, `DUR 3`, `CMD 4`

## Keyboard Shortcut System

### Inline Contextual Hints

Hints are placed where they're relevant:

- **Title bar:** `? help` at far right
- **Tab bar right side:** `S scan  M monitor` (action hints)
- **Column headers:** Sort numbers `1-8` next to column names
- **Breadcrumb row:** `→ drill  ← back` (namespace tab only)

All hint keys rendered in `ColorOrange`, labels in `ColorMuted`.

### ? Help Overlay

Pressing `?` shows a centered modal overlay (tview Modal or custom Flex):

- Background: `ColorSurface` with `ColorBorder` border
- Title: "Keyboard Shortcuts" in `ColorBlue`
- Subtitle: "press ? or Esc to close" in `ColorMuted`
- 3 columns:

**Navigation:** N=Namespaces, L=Slow Log, B=Big Keys, H=Hot Keys, T=Next tab
**Actions:** S=Run SCAN, M=Run MONITOR, →/Enter=Drill down, ←/Bksp=Level up, Q=Quit
**Sort:** 1=Keys, 2=Memory, 3=Avg TTL, 4-8=other columns (context-dependent)

Dismiss with `?` or `Esc`. Implemented as a tview Pages overlay.

## Loading & Disclaimer Screens

### Disclaimer Screen

Same content, GitHub Dark colors:

- Title: "DISCLAIMER" in `ColorRed`
- Body text in `ColorSecondary`
- MONITOR warning in `ColorOrange`
- Y/N: `ColorGreen` / `ColorRed`
- Background: `ColorBg`

### Loading Screen

- Spinner: `ColorOrange`
- Status text: `ColorText`
- Progress bar: Fill in `ColorOrange`, empty track in `ColorBorder`
- Percentage: `ColorSecondary`

### Error Screen

- "ERROR" title in `ColorRed`
- Message in `ColorText`
- R/Q: `ColorGreen` / `ColorRed`

## New Files

| File                           | Purpose                                            |
|--------------------------------|----------------------------------------------------|
| `lib/ui/theme.go`              | Color constants, helper functions for themed cells |
| `lib/ui/views/titlebar.go`     | Title bar component                                |
| `lib/ui/views/help_overlay.go` | ? help overlay modal                               |

## Modified Files

| File                                         | Changes                                                                  |
|----------------------------------------------|--------------------------------------------------------------------------|
| `lib/ui/ui.go`                               | New layout assembly, ? key handler, remove shortcuts bar                 |
| `lib/ui/views/header.go`                     | 3 panels, new styling, remove scan state panel                           |
| `lib/ui/views/body.go`                       | Remove shortcuts TextView, add action hints to tab bar                   |
| `lib/ui/views/components/namespace_table.go` | New colors, sort hints in headers, alternating rows, remove Types column |
| `lib/ui/views/components/bigkeys_table.go`   | Add Type/TTL/Namespace columns, new colors                               |
| `lib/ui/views/components/hotkeys_table.go`   | Add Command/Namespace columns, new colors                                |
| `lib/ui/views/components/slowlog_table.go`   | New colors, sort hints                                                   |
| `lib/ui/views/components/progress_bar.go`    | GitHub Dark colors for fill/track                                        |
| `lib/scanner/analytics.go`                   | Track top command per hot key, extract namespace for big/hot keys        |
| `models/special_keys.go`                     | Add fields: Type, TTL, Namespace to BigKey; Command, Namespace to HotKey |

## Out of Scope

- Responsive layout for narrow terminals (future)
- Mouse support (future)
- Config file for custom color themes (future)
- Export/save functionality (future)
