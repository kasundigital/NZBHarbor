package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kasundigital/NZBHarbor/internal/config"
	"github.com/kasundigital/NZBHarbor/internal/downloader"
	"github.com/kasundigital/NZBHarbor/internal/server"
	"github.com/kasundigital/NZBHarbor/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	st, err := store.New(cfg.ConfigDir)
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	engine := downloader.New(cfg, st)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go engine.Run(ctx)

	srv := server.New(cfg, st, engine)
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("server: %v", err)
	}
}
