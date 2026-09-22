# Detached Timer Robustness — Design

Date: 2026-09-21

Status: Draft (derived from a failure-mode review of the current detach flow)

## Goal

Make `timer -d` trustworthy: the confirmation must mean the timer really
started, the child must run the duration the user entered, and the
notification must fire when the wall clock reaches the deadline.

## Problems in the current behaviour

1. **Deadline race.** The parent parses the input and computes a duration,
   then spawns the child with the raw args; the child re-parses them. A
   `HH:MM` deadline that crosses the minute boundary between the two
   parses rolls over to +24h in the child.
2. **Unverified confirmation.** The parent prints "started" as soon as
   spawn succeeds. Child failures (arg re-parse, registry write,
   notification) are invisible: stdio is /dev/null and log errors may go
   nowhere. `beeep`/osascript also fails outright in headless sessions.
3. **Syntactic flag stripping.** Deleting every `-d` token from the argv
   corrupts commands where `-d` is a value, e.g. `timer -n -d 5m` (a timer
   named "-d").
4. **Clock mismatch.** The child sleeps on the monotonic clock, which
   stops during system sleep; the registry computes remaining from the
   wall clock. After a wake, `ps` shows 0:00 while the child is still
   sleeping.
5. **Weak marker.** Any non-empty `TIMER_DETACHED_CHILD` in the
   environment makes a foreground `timer` run headless, and a duplicate
   env key can leave the child running the TUI against /dev/null.
6. **Truncated durations.** The registry stores whole seconds, so
   sub-second timers disagree with the sleep.
7. **Misnamed start command.** `startCmd.Use` names the subcommand
   `timer`, so `timer start 5m` is rejected; the Run bodies of root and
   start are duplicated.

## Remediation

1. The parent passes the computed duration and name to the child through
   the environment (`TIMER_DETACHED_DURATION_NS`, `TIMER_DETACHED_NAME`).
   The child never re-parses user input.
2. The marker `TIMER_DETACHED_CHILD` counts only when it equals exactly
   `1`; all `TIMER_DETACHED_*` keys are stripped from the inherited
   environment before the payload is appended, so no stale marker leaks
   in.
3. The parent builds the child argv: a `0s` placeholder positional (to
   satisfy `ExactArgs(1)`) plus `-c <path>` when a custom config is set.
   The syntactic flag stripper is deleted.
4. Spawn returns the child pid; the parent polls the registry for
   `<pid>.json` for up to 2 s before confirming. On timeout it prints a
   failure line naming the timer log instead of claiming success.
5. The child waits in ticks of at most 1 s and notifies as soon as the
   wall clock reaches the deadline.
6. The registry stores the duration in nanoseconds (`duration_ns`),
   keeping `duration_seconds` (truncated) so older binaries can still
   read new files and vice versa.
7. `startCmd` is named `start`, and root and start share one `runStart`
   body.

## Out of scope

- Cancellation and resume of detached timers.
- Notification retry or a delivery guarantee.
- Platform support beyond darwin (unchanged: `cmd/detach_darwin.go`).
