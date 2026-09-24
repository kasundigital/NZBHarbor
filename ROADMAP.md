# NZBHarbor Roadmap

## v0.1 — working foundation
- [x] Docker-first Go service
- [x] NZB parser
- [x] NNTP TLS/authentication
- [x] yEnc decode
- [x] Parallel segment workers
- [x] Multi-provider fallback
- [x] Persistent queue/history
- [x] Web dashboard and provider test
- [x] PAR2 + RAR post-processing
- [x] Initial SABnzbd compatibility for Sonarr/Radarr
- [x] CI, multi-arch release workflow, documentation site

## v0.2 — stability foundation
- [x] Reusable NNTP connection pools per provider
- [x] Honor configured per-server connection limits
- [x] Retry transient article failures before provider fallback
- [x] yEnc size and CRC32/pCRC validation
- [x] NNTP socket deadlines and cancellation
- [x] Atomic/synced persistent state writes
- [x] Cancellable PAR2/RAR post-processing with timeout protection
- [x] Race-enabled Go CI
- [ ] Persist article-level failure/retry accounting
- [ ] Download health / critical-health calculation
- [ ] Disk-space preflight and low-space pause
- [ ] Better ETA and rolling speed sampling
- [ ] Global pause/resume and speed limit
- [ ] Queue reordering and priorities
- [ ] Password-protected archive support
- [ ] Post-processing retry without re-downloading
- [ ] Integration tests using a mock NNTP server

## v0.3 — automation compatibility
- [ ] SABnzbd API compatibility test matrix for Sonarr/Radarr/Lidarr/Whisparr
- [ ] Add NZB by URL
- [ ] Queue/history pagination and search
- [ ] User-defined categories and category directories
- [ ] Duplicate detection and duplicate keys
- [ ] Per-job priority and pause state
- [ ] Detailed failure reason codes for Servarr
- [ ] Completed Download Handling regression tests
- [ ] API versioning and compatibility fixtures

## v0.4 — operations and diagnostics
- [ ] Per-provider statistics: success, missing, failed, retried articles
- [ ] Download health and repairability indicators
- [ ] Structured logs with job/provider/article context
- [ ] Prometheus metrics
- [ ] Webhooks / Telegram / Discord notifications
- [ ] Support bundle with redacted config, logs and runtime status
- [ ] Config backup/restore and schema migrations
- [ ] Graceful database/state recovery tests
- [ ] Unraid template
- [ ] TrueNAS packaging

## v1.0 stable-release gate
NZBHarbor should not be called stable until these are true:

- [ ] No known data-loss or queue-corruption bugs
- [ ] Successful restart/resume testing during download and post-processing
- [ ] Missing-article and provider-failover testing with large NZBs
- [ ] PAR2 repair and RAR extraction test corpus passes
- [ ] Sonarr and Radarr download/import flows pass end-to-end
- [ ] At least 24-hour sustained download stress test passes
- [ ] Docker amd64 and arm64 images pass smoke tests
- [ ] Upgrade/migration test from the previous release passes
- [ ] Security review completed for API auth, file paths and uploaded NZBs
- [ ] Stable config/state schema with documented migration policy
- [ ] Recovery documentation and support bundle available

## Later / optional
These are useful, but they are not required to make the Docker downloader stable:

- [ ] RSS reader
- [ ] Watched folder
- [ ] Native Windows package
- [ ] Native macOS package
- [ ] Native Linux packages
