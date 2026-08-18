# Troubleshooting NZBHarbor

## First checks

```bash
docker compose ps
docker logs --tail=200 nzbharbor
curl -v http://127.0.0.1:6789/api/v1/health
```

Expected health response includes `"ok":true`.

## Web UI opens but says API key is wrong

With the default Compose setup:

```bash
grep api_key config/config.json
```

Clear an old browser key by removing `nzbharbor_api_key` from browser local storage or use a private window and enter the current key.

## Provider test fails

Check the provider's exact NNTP hostname, TLS setting and port, username/password, subscription status, DNS, and outbound firewall. A typical secure NNTP port is 563, but provider values differ.

Watch logs while testing:

```bash
docker logs -f nzbharbor
```

## Authentication succeeds but articles fail

An article can be missing from one provider because of retention/removal. Add a fill provider with a higher numeric priority. NZBHarbor falls back through enabled providers.

## Download stays failed with `no Usenet servers configured`

No provider has been saved. Add one under **Servers**, test it, save configuration, and retry the job.

## Sonarr/Radarr cannot connect

If all apps are containers, `localhost:6789` inside Sonarr is not NZBHarbor. Use the NZBHarbor service/container DNS name on a shared Docker network:

```text
nzbharbor:6789
```

Test network resolution from the Sonarr container if it contains curl/wget, or inspect both networks with `docker inspect`.

## Sonarr/Radarr downloads but cannot import

This is usually a path-mapping problem rather than NNTP. Give NZBHarbor and Servarr applications the same host data directory mounted at the same container path where possible. Check ownership and read/write permission as well.

## PAR2 repair fails

The Docker image includes `par2`. A repair can still fail if not enough recovery blocks exist. History records the post-processing error. Try another release/provider when recovery data is insufficient.

## RAR extraction fails

The image uses `unar`. Confirm the archive is complete and not encrypted with an unsupported/missing password. v0.1 archive handling is intentionally basic and will be expanded.

## Permission denied

Check the host directories:

```bash
ls -ld config downloads
```

The Docker process must be able to create files in both mounted paths. Also check SELinux/AppArmor/container policy on platforms where applicable.

## Port 6789 already in use

Change only the host-side mapping, for example:

```yaml
ports:
  - "6790:6789"
```

Then browse to port 6790. Containers on the same Docker network can still use port 6789 internally.

## Container keeps restarting

```bash
docker inspect nzbharbor --format '{{json .State}}'
docker logs --tail=300 nzbharbor
```

Look for malformed `config.json`, filesystem permissions, or a port/listen error.

## Collecting a useful bug report

Include NZBHarbor commit/version, host OS, architecture, Docker/Compose versions, sanitized logs, whether native upload or Sonarr/Radarr submitted the job, and the exact stage that failed. Never post API keys, Usenet credentials, or copyrighted NZB contents.
