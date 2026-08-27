package calderadb

import (
	"context"
	"errors"
	"sync"
	"time"
)

type PoolConfig struct {
	MaxSize      int
	Timeout      time.Duration
	ClientConfig ClientConfig
}

type Pool struct {
	config  ClientConfig
	clients chan *Client
	mu      sync.Mutex
	closed  bool
	created int
}

func NewPool(config PoolConfig) (*Pool, error) {
	if config.MaxSize <= 0 {
		config.MaxSize = 10
	}
	if config.Timeout > 0 && config.ClientConfig.Timeout == 0 {
		config.ClientConfig.Timeout = config.Timeout
	}
	pool := &Pool{
		config:  config.ClientConfig,
		clients: make(chan *Client, config.MaxSize),
	}
	for i := 0; i < config.MaxSize; i++ {
		client, err := NewClient(config.ClientConfig)
		if err != nil {
			pool.Close()
			return nil, err
		}
		pool.clients <- client
		pool.created++
	}
	return pool, nil
}

func (p *Pool) Get(ctx context.Context) (*Client, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case client, ok := <-p.clients:
		if !ok {
			return nil, ErrPoolExhausted
		}
		return client, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *Pool) Put(client *Client) error {
	if client == nil {
		return errors.New("nil client")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return client.Close()
	}
	p.clients <- client
	return nil
}

func (p *Pool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	close(p.clients)
	p.mu.Unlock()

	var firstErr error
	for client := range p.clients {
		if err := client.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
