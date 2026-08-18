FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/nzbharbor ./cmd/nzbharbor

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl par2 unar && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/nzbharbor /usr/local/bin/nzbharbor
COPY web /opt/nzbharbor/web
RUN mkdir -p /config /downloads
VOLUME ["/config","/downloads"]
EXPOSE 6789
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/bin/sh","-c","curl -fsS http://127.0.0.1:6789/api/v1/health >/dev/null || exit 1"]
ENTRYPOINT ["/usr/local/bin/nzbharbor"]
