## v0.12.0 (2026-08-27)

### Feat

- stop accepts prefix ID matches

## v0.11.0 (2026-08-27)

### BREAKING CHANGE

- timer builds only on macOS; Linux and Windows no longer compile.

### Feat

- drop Windows and Linux support

## v0.10.0 (2026-08-24)

### Feat

- add stop command for detached timers
- add registry stop that signals a timer and removes its file

## v0.9.0 (2026-08-24)

### Feat

- add ps command listing running timers
- detached timer records itself in the registry and logs errors
- add running dir and log path helpers
- read registry with staleness cleanup via process start times
- add registry write and remove for running timers
- add --detach flag to run timers in the background

### Fix

- always print ps table header, even with no running timers
- restore slog default in detach test and add pid to log lines
- keep registry files when process checks unsupported

## v0.8.0 (2026-08-23)

### Feat

- load timer art from art dir with embedded fallback
- add --name flag for named timers

## v0.7.1 (2026-08-23)

### Fix

- fire completion notification once

## v0.7.0 (2026-08-23)

### Feat

- desktop notifications
- config: make better (auto-create XDG config file)

### Docs

- readme updates

### Chore

- rm copyright comments
- update license

## 0.4.0 (2025-06-26)

### Feat

- readme

## 0.3.0 (2025-06-24)

### Feat

- add text color when timer is less than 1 min

## 0.2.0 (2025-06-24)

### Feat

- update install name

## 0.1.0 (2025-06-24)

### Feat

- make work
- more cleanup + init tea
- remove list command
- test fang
- init better
- init deps pattern
- init config
- init viper
- init commands
- init mise
- init cobra

## v0.6.0 (2025-08-02)

### Feat

- **start**: allow program exit on esc
- accept deadline

## v0.5.0 (2025-06-28)

### Feat

- stopwatch
- readme
- add text color when timer is less than 1 min
- update install name
- make work
- more cleanup + init tea
- remove list command
- test fang
- init better
- init deps pattern
- init config
- init viper
- init commands
- init mise
- init cobra

### Fix

- ux plus docs
- help docs
