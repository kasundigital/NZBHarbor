# API reference

NZBHarbor v0.1 has a native REST API plus a SABnzbd compatibility endpoint.

## Authentication

`GET /api/v1/health` is public. Native management endpoints accept:

```text
X-Api-Key: YOUR_KEY
```

The SAB endpoint uses `apikey=YOUR_KEY`.

## Native endpoints

```text
GET  /api/v1/health
GET  /api/v1/status
GET  /api/v1/jobs
POST /api/v1/jobs                  multipart NZB upload
POST /api/v1/jobs/{id}/pause
POST /api/v1/jobs/{id}/resume
POST /api/v1/jobs/{id}/delete
GET  /api/v1/config
PUT  /api/v1/config
POST /api/v1/test-server
```

Example:

```bash
curl -H 'X-Api-Key: YOUR_KEY' http://localhost:6789/api/v1/status
```

## SAB compatibility

Endpoint:

```text
/api?mode=MODE&output=json&apikey=YOUR_KEY
```

Initial supported modes/actions:

- `version`
- `get_config`
- `fullstatus`
- `addfile`
- `queue`
- queue `pause`, `resume`, `delete`
- `history`
- history `delete`
- `retry`

Queue/history accept a `category` filter. `addfile` accepts category through `cat`.

The compatibility API intentionally implements the Servarr-relevant subset first; it is not yet a byte-for-byte implementation of every SABnzbd API feature.
