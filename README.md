<p align="center"><img src="assets/logo.svg" width="110" alt="NZBHarbor logo"></p>
<h1 align="center">NZBHarbor</h1>
<p align="center"><strong>A Docker-first, open-source NZB/Usenet downloader with a SABnzbd-compatible API.</strong></p>

> **Status:** v0.1 development release. The core download path works, but NZBHarbor is still young. Test it alongside your existing downloader before trusting important automation.

## Why NZBHarbor?

NZBHarbor is being built around three goals: a simple self-hosted UI, useful failure diagnostics, and compatibility with Sonarr/Radarr without requiring Servarr to add a new download-client type.

### Included in v0.1

- Docker-first Go application
- NZB XML parsing and persistent queue/history
- TLS/plain NNTP authentication and article fetching
- yEnc decoding and multipart file assembly
- Multiple Usenet providers with priority/fill fallback
- Resume queue after container restart
- PAR2 repair and RAR extraction in the Docker image
- Responsive web dashboard
- Provider connection test
- API-key authentication
- Native REST API
- SABnzbd-compatible `version`, `get_config`, `fullstatus`, `addfile`, `queue`, `history`, `retry`, pause/resume/delete operations
- `tv`, `movies`, `music`, and `default` categories
- Sonarr/Radarr-compatible queue/history category filtering
- Docker health check
- amd64/arm64 container release workflow
- GitHub Pages documentation workflow

## Docker install

### One-command install / update

The same command installs NZBHarbor the first time and updates it later:

```bash
curl -fsSL https://raw.githubusercontent.com/kasundigital/NZBHarbor/main/install.sh | sudo bash
```

During development this installs the `edge` Docker image. Stable releases will use `latest`.

Default files:

```text
/opt/nzbharbor/docker-compose.yml
/opt/nzbharbor/config/
/opt/nzbharbor/downloads/
```

Open:

```text
http://YOUR-SERVER-IP:6789
```

On first start NZBHarbor generates an API key:

```bash
sudo cat /opt/nzbharbor/config/config.json
```

### Docker Compose

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

Start or update:

```bash
docker compose pull
docker compose up -d --remove-orphans
```

### Docker run

```bash
docker pull ghcr.io/kasundigital/nzbharbor:edge

docker rm -f nzbharbor 2>/dev/null || true

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

### Useful Docker commands

```bash
docker logs -f nzbharbor
docker ps --filter name=nzbharbor
cd /opt/nzbharbor && docker compose restart
cd /opt/nzbharbor && docker compose down
```

## Recommended media-stack paths

For Sonarr/Radarr, make sure both NZBHarbor and the Servarr app can see completed downloads through compatible paths. A shared `/data` layout is recommended in a larger Compose stack, for example:

```yaml
volumes:
  - /srv/media-data:/data
```

Then configure NZBHarbor downloads under `/data/downloads` and give Sonarr/Radarr the same `/data` mount. This reduces remote path mapping problems.

## Sonarr / Radarr

NZBHarbor presents a SABnzbd-compatible API, so in Sonarr/Radarr choose **SABnzbd** as the download client.

When all containers are on the same Docker network:

```text
Host: nzbharbor
Port: 6789
SSL: No (use HTTPS at your reverse proxy if needed)
API Key: <NZBHarbor API key>
Category in Sonarr: tv
Category in Radarr: movies
```

Do **not** use `localhost` from inside the Sonarr/Radarr container; it refers to that container itself.

See [Sonarr & Radarr setup](docs/guides/SONARR-RADARR.md).

## Files and persistence

```text
/config/config.json       Settings and provider credentials
/config/state.json        Persistent jobs/history
/config/nzbs/             Saved NZB metadata
/downloads/incomplete/    Temporary article segments
/downloads/complete/      Completed downloads
```

Back up `/config`. Protect it: `config.json` contains your Usenet password and API key.

## API

Health does not require authentication:

```bash
curl http://localhost:6789/api/v1/health
```

Native API requests use:

```text
X-Api-Key: YOUR_KEY
```

SAB-compatible requests use the usual `apikey` query parameter. See [API reference](docs/guides/API.md).

## Troubleshooting

Start here:

```bash
docker compose ps
docker logs --tail=200 nzbharbor
curl http://127.0.0.1:6789/api/v1/health
```

Common causes are wrong NNTP TLS/port settings, provider authentication, missing articles, inconsistent Docker paths, permissions, or an incorrect API key. The full checklist is in [Troubleshooting](docs/guides/TROUBLESHOOTING.md).

## Security

- Keep port `6789` on your LAN/private Docker network where possible.
- For Internet access, use an authenticated HTTPS reverse proxy or VPN.
- Never publish `/config/config.json`.
- Rotate the API key/provider password if exposed.

## Current limitations

NZBHarbor is still a development release. Connection pooling, per-provider limits, retries, NNTP timeouts, and yEnc CRC/size validation are now part of the stabilization branch, but more work remains around disk-space safety, article-level health tracking, PAR/RAR edge cases, post-processing retry, stress testing, and deeper Servarr compatibility.

See [ROADMAP.md](ROADMAP.md).

## Development

Requires Go 1.23+.

```bash
go test ./...
go vet ./...
go run ./cmd/nzbharbor
```

For local development, point runtime directories somewhere writable:

```bash
NZBHARBOR_CONFIG="$PWD/config" \
NZBHARBOR_DOWNLOADS="$PWD/downloads" \
NZBHARBOR_WEB="$PWD/web" \
go run ./cmd/nzbharbor
```

## License

MIT. See [LICENSE](LICENSE).

---

Built by [Kasun Indika](https://www.kasunindika.com).

---

## ☕ Support this project

This project is free and open source. If it helps you, you can support continued development:

<div align="center">
  <a href="https://buymeacoffee.com/kasundigital" target="_blank">
    <img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me a Coffee" height="50">
  </a>
</div>
