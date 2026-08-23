# Timer

A simple timer.

## Installation

`go install github.com/EwanGreer/timer@$version`

### From Source

Clone the repo. Then run `mise run build && cp ./bin/timer ~/go/bin/timer`

## Configuration

Timer automatically creates a configuration file on first run at:
- `$XDG_CONFIG_HOME/timer/config.toml` (if `XDG_CONFIG_HOME` is set)
- `~/.config/timer/config.toml` (default)

You can also specify a custom config file location with the `-c` flag:
```bash
timer -c /path/to/config.toml 5m
```

You can name a timer with the `-n` flag. The name appears in the completion notification:
```bash
timer -n Tea 5m
```

## Detached timers

Run a timer in the background with the `-d` flag. The prompt returns
immediately and the completion notification still fires:
```bash
timer -d -n Tea 5m
timer "Tea" started — will notify on completion
```

Detached timers do not take over the terminal, and survive closing it.
They cannot be cancelled — close a detached timer's notification and wait
for it, or leave it out. `-d` is not supported on Windows.

## Custom art

The timer reads its display art from files at runtime, so you can customise it
without rebuilding. Put art files in the `art` directory next to your config:

- `$XDG_CONFIG_HOME/timer/art/` (if `XDG_CONFIG_HOME` is set)
- `~/.config/timer/art/` (default)

| File         | Used for                          |
| ------------ | --------------------------------- |
| `done.txt`   | The DONE art shown on completion  |
| `0.txt`–`9.txt` | The big digit glyphs           |
| `colon.txt`  | The `:` glyph in the clock        |

Digit glyphs must be exactly 5 rows, with at least one non-blank row; a
trailing newline after the last row is fine. Rows are padded to the widest
glyph, so uneven widths still line up.

Missing files silently fall back to the built-in art (embedded in the binary
at build time). Files that exist but cannot be read or parsed also fall back,
with a warning printed on startup.

