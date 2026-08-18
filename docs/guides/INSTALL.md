# Install NZBHarbor with Docker

## Requirements

- Docker Engine with Docker Compose v2
- A Usenet provider account
- A writable configuration directory
- Enough disk space for incomplete and completed downloads

## Install from source

```bash
git clone https://github.com/kasundigital/NZBHarbor.git
cd NZBHarbor
docker compose up -d --build
```

Open `http://SERVER-IP:6789`.

## First-run API key

NZBHarbor generates an API key and saves it in `./config/config.json` with the default Compose file.

```bash
cat config/config.json
```

Paste `api_key` into the web UI when prompted.

## Add a provider

Open **Servers → Add server**. Enter the hostname, port, TLS setting, username, password, and priority. Port 563 with TLS is common, but use the values supplied by your provider. Press **Test**, then save.

Use priority `0` for the primary provider. A fill provider could use `10`; NZBHarbor tries lower numbers first.

## Update

```bash
git pull
docker compose up -d --build
```

## Backup

Stop the app or take a consistent filesystem snapshot, then back up the `config` directory. It contains configuration, job state, and saved NZB metadata.
