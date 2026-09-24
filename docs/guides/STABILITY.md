# NZBHarbor Stability Plan

NZBHarbor is intentionally Docker-first. Native desktop packages are not part of the stable-release requirement.

## What "stable" means

A stable NZBHarbor release must reliably download, verify, repair, extract, resume, and expose correct queue/history state to automation clients. Feature count alone is not the goal.

## Tier 1: release blockers

- Reusable NNTP connection pools with per-provider limits
- Retry transient article failures and fall back to fill providers
- Detect missing articles separately from transport failures
- yEnc size and CRC validation
- Persistent article/job state that survives container restarts
- Atomic config/state writes and recovery from interrupted writes
- Disk-space preflight and low-space handling
- PAR2 verification/repair with explicit failure states
- Archive extraction with timeout/cancel handling
- Password-protected archive handling
- Retry post-processing without re-downloading completed data
- Correct pause/resume/delete behavior during active work
- Correct graceful shutdown and restart recovery
- End-to-end Sonarr/Radarr add, queue, history, retry and import tests
- Large-NZB and 24-hour stress tests
- Docker amd64/arm64 smoke tests
- Upgrade/migration tests

## Tier 2: operational quality

- Rolling speed and ETA
- Per-provider success/missing/failure/retry statistics
- Download health and critical-health indicators
- Queue priority and reordering
- Global pause/resume
- Speed limit
- Structured logs
- Redacted support bundle
- Prometheus metrics
- Config backup/restore
- User-defined categories and paths
- Queue/history pagination and search
- Duplicate detection

## Tier 3: optional ecosystem features

These are useful but are not required for the first stable Docker release:

- RSS
- Watched folder
- Scheduled speed changes
- Notification integrations
- Custom pre/post-processing scripts
- Native Windows/macOS/Linux packages

## Compatibility target

NZBHarbor should keep its native API while providing the SABnzbd API surface required by Sonarr, Radarr, Lidarr and Whisparr. Compatibility should be verified with automated fixtures rather than by advertising a SAB version number alone.

## Release policy

- v0.x: development / release candidates
- v1.0.0: first stable Docker release
- stable releases only after CI, integration tests, stress tests and migration tests pass
- `latest` should point only to a stable release
- development images should use a separate tag such as `edge` or `rc`
