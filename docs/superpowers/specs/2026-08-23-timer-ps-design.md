# Timer Registry and `timer ps` — Design

Date: 2026-08-23

Status: Approved (sections 1–3 reviewed in chat; this document records the
agreed design)

## Goal

Add a `timer ps` command that lists running detached timers, backed by a
registry that will later support cancelling them. Today a detached timer
leaves no trace: the child process waits and exits. This design adds the
trace.

## 1. Registry

### Location

State files live in a `running/` directory next to the config:

- `$XDG_CONFIG_HOME/timer/running/` when `XDG_CONFIG_HOME` is set
- `~/.config/timer/running/` otherwise
- a custom `-c` path derives it from the config file's directory

A new `getRunningDir()` in `cmd/root.go` mirrors the existing
`getArtDir()`.

### Files

One file per timer, named `<pid>.json`, containing:

```json
{"name":"Tea","duration_seconds":300,"started_at":"2026-08-23T14:32:05+01:00"}
```

- `name` — may be empty (unnamed timer)
- `duration_seconds` — the full duration parsed at start, as integer
  seconds
- `started_at` — RFC 3339 wall-clock time when the file is written

JSON is internal runtime state, not user config; the stdlib encoder is
used rather than adding a TOML writer dependency.

### Lifecycle

- The **detached child** writes its own file when it starts (pid from
  `os.Getpid()`, start time from `time.Now()`) and deletes it when the
  countdown ends. This replaces the current bare `runDetached` call with a
  wrapper in `cmd/detach.go`: write → run → remove. The child writes its
  own file so there is no race with the parent process.
- A crash, `kill -9`, or reboot leaves a stale file. `Read` removes stale
  files as it lists (see below). No other cleanup exists.
- Remove-on-completion is best-effort: if removal fails, the error is
  logged and the leftover file is cleaned by the next `ps`.

### Staleness check

For each file, `Read`:

1. Checks the pid is alive (signal 0).
2. Compares the process start time against the recorded `started_at`
   (±2 s tolerance; recorded time is slightly later than process start).
   - macOS: `golang.org/x/sys/unix` sysctl KERN_PROC_PID
   - Linux: `/proc/<pid>/stat` starttime field plus btime from
     `/proc/stat`
3. Either check failing, or an unparseable file, deletes the file and
   excludes the timer from the listing.

The start-time check makes a reused pid unable to resurrect a stale timer,
which matters once cancel (`timer rm <id>`) sends signals based on these
entries. `golang.org/x/sys` is already an indirect dependency
(v0.32.0); this usage promotes it to direct.

### Package layout

New `internal/registry` package:

- `Entry{Name string, Duration time.Duration, StartedAt time.Time}`
- `Write(dir string, e Entry) error` — writes `<pid>.json`
- `Remove(dir string, pid int) error` — deletes the file
- `Read(dir string) ([]Timer, error)` — cleans stale files, returns live
  timers (`Timer{Pid int, Name string, Duration time.Duration, StartedAt time.Time, Remaining time.Duration}`); a missing
  directory returns an empty list, no error
- `proc_unix.go` / `proc_windows.go` (build-tagged) — pid liveness and
  process start time, injectable as package vars so tests can stub them

## 2. `timer ps` command

New `cmd/ps.go`:

- `timer ps` — reads the registry, prints a table to stdout, exit 0.
- Columns: `ID` (pid), `NAME` (`-` when unnamed), `REMAINING`
  (recomputed: `started_at + duration − now`, formatted `3m24s`,
  `1h02m10s`), `STARTED` (`HH:MM` local time).
- Oldest first; column-aligned with stdlib `tabwriter`.
- Empty registry or missing `running/` dir: prints nothing, exit 0
  (script-friendly).
- Errors (unreadable dir, not supported on Windows): message to stderr,
  exit 1.
- The command is thin: it calls the registry through a stub-able package
  var (the same injection pattern as `spawnDetached`), so tests can
  fixture the table output.
- On Windows the command errors "not supported", mirroring `--detach`.

Example:

```
$ timer ps
ID    NAME    REMAINING   STARTED
2491  Tea     3m24s       14:32
2930  -       12m07s      14:31
```

## 3. Error handling and logging

- **Detached child:** on start, opens `$XDG_CONFIG_HOME/timer/timer.log`
  (next to config, derived like the running dir; `getLogPath()`) in
  append mode and repoints the existing slog default to JSON output on
  that file. Registry write/remove failures and notification failures are
  logged there. If the log file cannot be opened, the timer runs without
  logging (there is nowhere left to report that failure).
- **Foreground commands** (`ps`): errors print to stderr and exit 1; they
  do not touch the log file.
- Log entries are slog JSON lines appended to one shared file; each line
  carries the timer's pid where available.

## 4. Testing

TDD throughout, same style as the detach work.

`internal/registry` (real files in a temp dir, stubbed proc checks):

- Write/Read roundtrip returns the written timer with correct pid and
  remaining.
- Dead pid → file removed, timer absent from listing.
- Start-time mismatch (±2 s) → file removed, timer absent.
- Unparseable file → removed, no error.
- Missing dir → empty list, no error.
- Remove deletes the file.
- Proc check stubs are package vars (`procAlive`, `procStartedAt`), like
  the `notify` stub.

`cmd`:

- `ps` table from fixed entries: ordering (oldest first), `-` for unnamed,
  remaining formatting (`3m24s`, `1h02m10s`), STARTED as local HH:MM.
- Empty registry → empty stdout, exit 0.
- Read error → stderr, exit 1 (via cobra's error return).
- Detached-child wrapper: registry write and remove called around
  `runDetached`; a write/remove failure is logged — tested by capturing
  slog into a buffer and forcing failure with an unwritable dir.
- Windows variant of the ps command compiles (build tags).

Smoke test with the real binary:

1. `timer -d -n Tea 30s` → confirmation line; `timer ps` shows one row
   with remaining ≤ 30s.
2. After completion the row is gone.
3. `timer -d -n Ghost 30s`, then `kill -9` the child pid → `timer ps`
   shows nothing and removes the stale file.
4. A registry write failure (unwritable running dir) leaves a line in
   `timer.log` and the timer still completes.

## 5. Documentation

Extend the "Detached timers" README section with a `timer ps` example and
mention of `timer.log`.

## Out of scope

- Cancelling timers (`timer rm`/`timer stop`) — a later feature; the
  registry is shaped for it (pid and start-time verification).
- Listing foreground/TUI timers — only detached timers have files.
- Log rotation — the log file grows; acceptable for now.
