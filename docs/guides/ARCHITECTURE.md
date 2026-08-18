# Architecture

NZBHarbor is a single Go service with a static web frontend.

```text
Browser / Sonarr / Radarr
          |
          v
+---------------------------+
| HTTP server :6789         |
| Native API + SAB API      |
+-------------+-------------+
              |
              v
+---------------------------+
| Persistent queue engine   |
| /config/state.json        |
+-------------+-------------+
              |
              v
+---------------------------+
| NZB parser + segment work |
+-------------+-------------+
              |
      +-------+-------+
      v               v
 Primary NNTP      Fill NNTP
      |               |
      +-------+-------+
              v
        yEnc decode
              |
              v
   incomplete segments
              |
              v
      assembled files
              |
              v
       PAR2 -> unar
              |
              v
   /downloads/complete
```

## Persistence

Configuration and state use small JSON files in `/config`; this keeps the first release simple, inspectable, and backup-friendly. NZB files are retained under `/config/nzbs` to support restart recovery.

## Concurrency

The engine runs one job at a time in v0.1 and parallelizes article segments within that job using `max_workers`. Article files are written independently and then assembled in segment-number order.

## Provider fallback

Enabled providers are sorted by ascending priority. An unavailable article/connect/decode failure moves to the next provider.

## Post-processing

After assembly, the Docker image can run `par2 r` when PAR2 files exist and `unar` for RAR sets. v0.1 favors simple deterministic behavior; richer policy/progress support is on the roadmap.

## Next performance step

NNTP connection pooling is the most important performance upgrade. v0.1 creates a connection for each article, which is easy to reason about but adds substantial latency. v0.2 will maintain per-provider reusable pools and honor each provider's connection limit.
