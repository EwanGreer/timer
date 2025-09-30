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
