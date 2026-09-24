<p align="center">
  <img src="assets/logo.svg" width="120" alt="NZBHarbor logo">
</p>

<h1 align="center">NZBHarbor</h1>

<p align="center">
  <strong>A Docker-first, open-source NZB / Usenet downloader with a SABnzbd-compatible API.</strong>
</p>

<p align="center">
  <a href="https://github.com/kasundigital/NZBHarbor/actions/workflows/ci.yml"><img src="https://github.com/kasundigital/NZBHarbor/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/kasundigital/NZBHarbor/pkgs/container/nzbharbor"><img src="https://img.shields.io/badge/Docker-GHCR-2496ED?logo=docker&logoColor=white" alt="Docker"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="MIT License"></a>
  <img src="https://img.shields.io/badge/amd64-supported-success" alt="amd64">
  <img src="https://img.shields.io/badge/arm64-supported-success" alt="arm64">
</p>

> [!IMPORTANT]
> **Current status: v0.2 development / stabilization.**
> NZBHarbor is usable for testing, but it is not yet the v1.0 stable release. The current Docker image is tagged `edge`. Do not replace a trusted production downloader until NZBHarbor has completed the stable-release test gate.

## What is NZBHarbor?

NZBHarbor is a lightweight self-hosted NZB downloader built in Go.

The project focuses on:

- Docker-first installation and updates
- SABnzbd API compatibility for Sonarr, Radarr and other Servarr apps
- Fast reusable NNTP connections
- Multiple Usenet providers with priority and fill-server fallback
- Persistent queue and history
- Restart-safe downloads
- PAR2 repair and archive extraction
- Useful failure information instead of silent failures
- A simple responsive web interface
- Easy self-hosting without a large dependency stack

## Features

| Area | Current support |
| --- | --- |
| Docker | Docker Compose, Docker Run, amd64, arm64 |
| NZB | XML parsing, multipart assembly |
| NNTP | TLS/plain connections, authentication, reusable connection pools |
| Providers | Multiple servers, priority order, per-server connection limits |
| Reliability | Transient retries, fill-provider fallback, socket timeouts |
| Integrity | yEnc size validation, CRC32 / pCRC32 validation |
| Queue | Persistent queue/history, pause, resume, delete, retry |
| Recovery | Resume after container restart using saved segment files |
| Post-processing | PAR2 repair and RAR extraction |
| Automation | SABnzbd-compatible API for Servarr |
| Categories | `default`, `tv`, `movies`, `music` |
| Security | API-key protected native API |
| UI | Responsive web dashboard and provider connection test |
| Health | Docker health check endpoint |
| Development | Go tests, race detector, vet and Docker build CI |

## Install or update with one command

The same command is used for the **first installation and future updates**:

```bash
curl -fsSL https://raw.githubusercontent.com/kasundigital/NZBHarbor/main/install.sh | sudo bash
```

The installer:

- Creates `/opt/nzbharbor`
- Creates persistent config and download directories
- Writes the Docker Compose file
- Pulls the newest NZBHarbor image
- Starts or updates the container
- Keeps your existing settings and downloads

Default paths:

```text
/opt/nzbharbor/docker-compose.yml
/opt/nzbharbor/config/
/opt/nzbharbor/downloads/
```

Open the web interface:

```text
http://YOUR-SERVER-IP:6789
```

### Get the API key

NZBHarbor creates an API key on the first start:

```bash
sudo cat /opt/nzbharbor/config/config.json
```

Use that key in the web UI and when connecting Sonarr/Radarr.

## Docker Compose

```yaml
services:
  nzbharbor:
    image: ghcr.io/kasundigital/nzbharbor:edge
    container_name: nzbharbor
    restart: unless-stopped
    init: true
    stop_grace_period: 30s

    ports:
      - "6789:6789"

    environment:
      NZBHARBOR_CONFIG: /config
      NZBHARBOR_DOWNLOADS: /downloads

    volumes:
      - ./config:/config
      - ./downloads:/downloads
```

Start:

```bash
docker compose up -d
```

Update:

```bash
docker compose pull
docker compose up -d --remove-orphans
```

## Docker Run

Create persistent directories:

```bash
sudo mkdir -p /opt/nzbharbor/config /opt/nzbharbor/downloads
```

Run:

```bash
docker run -d \
  --name nzbharbor \
  --restart unless-stopped \
  --init \
  -p 6789:6789 \
  -e NZBHARBOR_CONFIG=/config \
  -e NZBHARBOR_DOWNLOADS=/downloads \
  -v /opt/nzbharbor/config:/config \
  -v /opt/nzbharbor/downloads:/downloads \
  ghcr.io/kasundigital/nzbharbor:edge
```

Update a Docker Run installation:

```bash
docker pull ghcr.io/kasundigital/nzbharbor:edge
docker rm -f nzbharbor

docker run -d \
  --name nzbharbor \
  --restart unless-stopped \
  --init \
  -p 6789:6789 \
  -e NZBHARBOR_CONFIG=/config \
  -e NZBHARBOR_DOWNLOADS=/downloads \
  -v /opt/nzbharbor/config:/config \
  -v /opt/nzbharbor/downloads:/downloads \
  ghcr.io/kasundigital/nzbharbor:edge
```

Your settings and downloads remain in `/opt/nzbharbor`, so recreating the container does not remove them.

## Useful Docker commands

```bash
# Container status
docker ps --filter name=nzbharbor

# Follow logs
docker logs -f nzbharbor

# Last 200 log lines
docker logs --tail=200 nzbharbor

# Restart
cd /opt/nzbharbor && docker compose restart

# Stop
cd /opt/nzbharbor && docker compose down

# Health check
curl http://127.0.0.1:6789/api/v1/health
```

## Add a Usenet provider

Open the NZBHarbor web UI and go to **Servers**.

Configure:

```text
Name
Host
Port
TLS
Username
Password
Connections
Priority
Enabled
```

Use **Test** to verify the connection and authentication before downloading.

Lower priority numbers are tried first. Other enabled providers can be used as fallback/fill servers when an article is missing or a provider fails.

## Sonarr / Radarr

NZBHarbor exposes a SABnzbd-compatible API.

In Sonarr or Radarr:

1. Open **Settings > Download Clients**
2. Add **SABnzbd**
3. Use the NZBHarbor host and port
4. Enter the NZBHarbor API key
5. Set the matching category

Typical Docker setup:

```text
Host: nzbharbor
Port: 6789
SSL: No
API Key: YOUR_NZBHARBOR_API_KEY

Sonarr category: tv
Radarr category: movies
```

> [!NOTE]
> Do not use `localhost` from inside a Sonarr/Radarr container. Inside Docker, `localhost` means that container itself. Use the NZBHarbor container/service name when both applications share a Docker network.

See [Sonarr & Radarr setup](docs/guides/SONARR-RADARR.md).

## Recommended media paths

For a Servarr stack, shared Docker paths make importing easier.

Example:

```yaml
volumes:
  - /srv/media-data:/data
```

A practical layout is:

```text
/data/downloads/
/data/media/movies/
/data/media/tv/
```

Give NZBHarbor and the relevant Servarr containers the same `/data` mount whenever possible.

This avoids unnecessary Remote Path Mappings.

## Persistent files

Inside the container:

```text
/config/config.json       NZBHarbor settings and provider credentials
/config/state.json        Persistent queue and history
/config/nzbs/             Saved NZB metadata

/downloads/incomplete/    Temporary downloaded article segments
/downloads/complete/      Completed downloads
```

With the one-command installer these are stored on the host at:

```text
/opt/nzbharbor/config/
/opt/nzbharbor/downloads/
```

### Backup

The most important directory to back up is:

```text
/opt/nzbharbor/config
```

It contains settings, queue/history state, NZB metadata and credentials.

## API

### Health

Authentication is not required:

```bash
curl http://localhost:6789/api/v1/health
```

### Native API

Use:

```text
X-Api-Key: YOUR_KEY
```

### SABnzbd-compatible API

SAB-compatible requests use the standard:

```text
apikey=YOUR_KEY
```

Current compatibility includes:

- `version`
- `get_config`
- `fullstatus`
- `addfile`
- `queue`
- `history`
- `retry`
- queue pause/resume/delete operations
- category filtering

See [API reference](docs/guides/API.md).

## Reliability work included in v0.2

The stabilization work adds several changes that are important for real Usenet workloads:

- NNTP connection reuse instead of reconnecting for every article
- Configured per-provider connection limits
- Retry of transient article failures
- Provider fallback
- NNTP connection and article timeouts
- Download cancellation that interrupts network I/O
- yEnc decoded-size checking
- yEnc CRC32 and pCRC32 validation
- Safer persistent state writes
- Restart recovery
- Cancellable PAR2 repair
- Cancellable RAR extraction
- Post-processing timeout protection
- Race-detector CI

## Work still required before v1.0

The current development release should not yet be treated as a full SABnzbd/NZBGet replacement.

Important remaining work includes:

- Article-level success, retry, missing and failure statistics
- Download health / critical-health calculations
- Disk-space preflight and low-space protection
- Post-processing retry without downloading again
- More PAR2/RAR edge-case testing
- Password-protected archive support
- Better rolling download speed and ETA
- Queue priorities and reordering
- Global pause/resume and bandwidth limiting
- User-defined categories and directories
- Duplicate detection
- Deeper Sonarr/Radarr/Lidarr/Whisparr compatibility tests
- Large-NZB stress testing
- Restart testing during active download and post-processing
- Config/state schema migration testing
- Redacted support bundle and richer diagnostics

See [ROADMAP.md](ROADMAP.md) and [Stable release criteria](docs/guides/STABILITY.md).

## Release channels

| Tag | Purpose |
| --- | --- |
| `edge` | Current development/stabilization builds |
| `latest` | Reserved for stable releases |
| `x.y.z` | Versioned stable release |

Until v1.0 is ready, the installer uses `edge`.

## Troubleshooting

Start with:

```bash
docker ps --filter name=nzbharbor
docker logs --tail=200 nzbharbor
curl http://127.0.0.1:6789/api/v1/health
```

Common problems include:

- Wrong NNTP host or port
- TLS setting does not match the provider port
- Incorrect provider username/password
- Provider connection limit exceeded
- Missing Usenet articles
- Wrong Docker volume paths
- Sonarr/Radarr using `localhost`
- File permissions
- Incorrect API key
- Insufficient disk space
- PAR2 or archive extraction errors

See the full [Troubleshooting guide](docs/guides/TROUBLESHOOTING.md).

## Security

- Keep port `6789` on your LAN or private Docker network when possible.
- Use an authenticated HTTPS reverse proxy or VPN for remote access.
- Do not expose `/config/config.json`.
- `config.json` contains your API key and Usenet provider credentials.
- Rotate credentials if the config or API key is exposed.
- Back up `/config` securely.

## Development

Requires Go 1.23+.

```bash
go test -race ./...
go vet ./...
docker build -t nzbharbor:test .
```

Run locally:

```bash
NZBHARBOR_CONFIG="$PWD/config" \
NZBHARBOR_DOWNLOADS="$PWD/downloads" \
NZBHARBOR_WEB="$PWD/web" \
go run ./cmd/nzbharbor
```

## Documentation

- [Installation guide](docs/guides/INSTALL.md)
- [Sonarr & Radarr setup](docs/guides/SONARR-RADARR.md)
- [API reference](docs/guides/API.md)
- [Architecture](docs/guides/ARCHITECTURE.md)
- [Troubleshooting](docs/guides/TROUBLESHOOTING.md)
- [Stable release criteria](docs/guides/STABILITY.md)
- [Roadmap](ROADMAP.md)
- [Security policy](SECURITY.md)
- [Contributing](CONTRIBUTING.md)

## Contributing

Bug reports, compatibility reports and pull requests are welcome.

When reporting a download problem, include:

- NZBHarbor version/image tag
- Docker architecture
- Relevant sanitized logs
- Number of configured providers
- Whether the problem occurs before, during or after download
- Whether retrying succeeds

Do not include provider passwords or your NZBHarbor API key.

## License

NZBHarbor is released under the [MIT License](LICENSE).

---

<p align="center">
  Built by <a href="https://www.kasunindika.com">Kasun Indika</a>
</p>

<p align="center">
  If NZBHarbor helps you, you can support continued development.
</p>

<p align="center">
  <a href="https://buymeacoffee.com/kasundigital">
    <img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me a Coffee" height="50">
  </a>
</p>
