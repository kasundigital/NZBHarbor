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

## v0.2 — performance and integrity
- [ ] Reusable NNTP connection pools per provider
- [ ] Honor configured per-server connection limits
- [ ] yEnc CRC32/pCRC validation
- [ ] Better missing-article/retry accounting
- [ ] Pause global queue and speed limit
- [ ] Better ETA/speed sampling
- [ ] Queue reordering and priorities
- [ ] Extended automated NNTP protocol tests

## v0.3 — post-processing
- [ ] Detailed PAR2 verification/repair progress
- [ ] More archive layouts and password handling
- [ ] Cleanup policies
- [ ] Custom post-processing hooks
- [ ] Disk-space preflight

## v0.4 — ecosystem
- [ ] Deeper SAB API compatibility tests against Sonarr/Radarr/Lidarr/Whisparr
- [ ] Prometheus metrics
- [ ] Webhooks / Telegram / Discord notifications
- [ ] Unraid template
- [ ] TrueNAS packaging

## v1.0
- [ ] Stable migration/config format
- [ ] Windows binary/package
- [ ] macOS binary/package
- [ ] Linux native packages
- [ ] Complete operational diagnostics and support bundle
