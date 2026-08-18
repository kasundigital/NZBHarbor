# Sonarr and Radarr integration

NZBHarbor implements the SABnzbd API surface required for its initial Servarr integration. You therefore select **SABnzbd** in Sonarr/Radarr.

## Same Docker network

Use:

```text
Name: NZBHarbor
Host: nzbharbor
Port: 6789
Use SSL: No
API Key: <from NZBHarbor config/UI>
Category: tv      (Sonarr)
Category: movies  (Radarr)
```

If your containers are in different Compose projects, place them on a shared external Docker network or use a reachable server hostname/IP.

## Download paths

The applications must agree on where completed files live. Prefer a shared mount such as:

```text
Host: /srv/data
Container: /data
```

for NZBHarbor, Sonarr, and Radarr. Then use `/data/downloads` consistently.

## What Sonarr/Radarr can do

The initial compatibility layer supports adding NZBs, polling queue/history, reading completed directory/category configuration, retrying failed jobs, and deleting queue/history entries.

## Test fails

1. Verify `curl http://nzbharbor:6789/api/v1/health` from the relevant Docker network.
2. Confirm the API key.
3. Do not use `localhost` unless NZBHarbor actually runs in the same network namespace.
4. Check `docker logs -f nzbharbor` while pressing **Test** in Sonarr/Radarr.
