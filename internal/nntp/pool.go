package nntp

import (
	"context"
	"fmt"
	"sync"

	"github.com/kasundigital/NZBHarbor/internal/config"
)

type Pool struct {
	server config.NewsServer
	slots  chan struct{}
	idle   chan *Client
	mu     sync.Mutex
	closed bool
}

func NewPool(server config.NewsServer) *Pool {
	connections := server.Connections
	if connections < 1 {
		connections = 1
	}
	return &Pool{
		server: server,
		slots:  make(chan struct{}, connections),
		idle:   make(chan *Client, connections),
	}
}

func (p *Pool) Body(ctx context.Context, messageID string) ([]byte, error) {
	select {
	case p.slots <- struct{}{}:
		defer func() { <-p.slots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, fmt.Errorf("NNTP pool is closed")
	}
	p.mu.Unlock()

	var client *Client
	select {
	case client = <-p.idle:
	default:
		var err error
		client, err = Dial(p.server)
		if err != nil {
			return nil, err
		}
	}

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			client.interrupt()
		case <-done:
		}
	}()

	body, err := client.Body(messageID)
	close(done)
	if err != nil || ctx.Err() != nil {
		_ = client.Close()
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()
	if closed {
		_ = client.Close()
		return body, nil
	}

	select {
	case p.idle <- client:
	default:
		_ = client.Close()
	}
	return body, nil
}

func (p *Pool) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()

	for {
		select {
		case client := <-p.idle:
			_ = client.Close()
		default:
			return
		}
	}
}
