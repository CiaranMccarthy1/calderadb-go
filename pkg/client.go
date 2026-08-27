package calderadb

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// Client represents a CalderaDB client
type Client struct {
	addr         string
	conn         net.Conn
	reader       *bufio.Reader
	mu           sync.Mutex
	timeout      time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
}

// ClientConfig configures a CalderaDB client
type ClientConfig struct {
	Address      string
	Timeout      time.Duration // Default: 5s
	ReadTimeout  time.Duration // Default: 5s
	WriteTimeout time.Duration // Default: 5s
}

// NewClient creates a new CalderaDB client
func NewClient(config ClientConfig) (*Client, error) {
	if config.Address == "" {
		config.Address = "localhost:9090"
	}
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 5 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 5 * time.Second
	}

	conn, err := net.DialTimeout("tcp", config.Address, config.Timeout)
	if err != nil {
		return nil, &Error{Op: "connect", Err: err}
	}

	return &Client{
		addr:         config.Address,
		conn:         conn,
		reader:       bufio.NewReader(conn),
		timeout:      config.Timeout,
		readTimeout:  config.ReadTimeout,
		writeTimeout: config.WriteTimeout,
	}, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Set adds or updates a document
func (c *Client) Set(ctx context.Context, collection, key string, value interface{}) error {
	if ctx == nil {
		ctx = context.Background()
	}

	data, err := json.Marshal(value)
	if err != nil {
		return newErrorWithKey("set", key, err)
	}

	cmd := fmt.Sprintf("SET %s %s %s", collection, key, data)
	return c.sendCommand(ctx, cmd)
}

// Get retrieves a document by key
func (c *Client) Get(ctx context.Context, collection, key string) (map[string]interface{}, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := fmt.Sprintf("GET %s %s", collection, key)
	resp, err := c.sendCommandWithResponse(ctx, cmd)
	if err != nil {
		if strings.Contains(err.Error(), "NOT_FOUND") {
			return nil, newErrorWithKey("get", key, ErrKeyNotFound)
		}
		return nil, newErrorWithKey("get", key, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return nil, newErrorWithKey("get", key, err)
	}
	return result, nil
}

// Delete removes a document
func (c *Client) Delete(ctx context.Context, collection, key string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := fmt.Sprintf("DEL %s %s", collection, key)
	return c.sendCommand(ctx, cmd)
}

// Stats retrieves database statistics
func (c *Client) Stats(ctx context.Context) (*Stats, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	resp, err := c.sendCommandWithResponse(ctx, "STATS")
	if err != nil {
		return nil, newError("stats", err)
	}

	var stats Stats
	if err := json.Unmarshal([]byte(resp), &stats); err != nil {
		return nil, newError("stats", err)
	}
	return &stats, nil
}

// FindByPrefix finds documents with keys starting with a prefix
func (c *Client) FindByPrefix(ctx context.Context, prefix string) ([]TierDocument, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := fmt.Sprintf("FIND %s", prefix)
	resp, err := c.sendCommandWithResponse(ctx, cmd)
	if err != nil {
		return nil, newError("find", err)
	}

	var results []TierDocument
	if err := json.Unmarshal([]byte(resp), &results); err != nil {
		return nil, newError("find", err)
	}
	return results, nil
}

// GetHotTier retrieves all documents in the hot tier for a collection
func (c *Client) GetHotTier(ctx context.Context, collection string) ([]TierDocument, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := fmt.Sprintf("GET_HOT %s", collection)
	resp, err := c.sendCommandWithResponse(ctx, cmd)
	if err != nil {
		return nil, newError("get_hot", err)
	}

	var results []TierDocument
	if err := json.Unmarshal([]byte(resp), &results); err != nil {
		return nil, newError("get_hot", err)
	}
	return results, nil
}

// GetColdTier retrieves all documents in the cold tier for a collection
func (c *Client) GetColdTier(ctx context.Context, collection string) ([]TierDocument, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := fmt.Sprintf("GET_COLD %s", collection)
	resp, err := c.sendCommandWithResponse(ctx, cmd)
	if err != nil {
		return nil, newError("get_cold", err)
	}

	var results []TierDocument
	if err := json.Unmarshal([]byte(resp), &results); err != nil {
		return nil, newError("get_cold", err)
	}
	return results, nil
}

// GetAllHot retrieves all documents in the hot tier across all collections
func (c *Client) GetAllHot(ctx context.Context) ([]TierDocument, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	resp, err := c.sendCommandWithResponse(ctx, "GET_ALL_HOT")
	if err != nil {
		return nil, newError("get_all_hot", err)
	}

	var results []TierDocument
	if err := json.Unmarshal([]byte(resp), &results); err != nil {
		return nil, newError("get_all_hot", err)
	}
	return results, nil
}

// GetAllCold retrieves all documents in the cold tier across all collections
func (c *Client) GetAllCold(ctx context.Context) ([]TierDocument, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	resp, err := c.sendCommandWithResponse(ctx, "GET_ALL_COLD")
	if err != nil {
		return nil, newError("get_all_cold", err)
	}

	var results []TierDocument
	if err := json.Unmarshal([]byte(resp), &results); err != nil {
		return nil, newError("get_all_cold", err)
	}
	return results, nil
}

// CreateCollection creates a new collection
func (c *Client) CreateCollection(ctx context.Context, name string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := fmt.Sprintf("CREATE %s", name)
	return c.sendCommand(ctx, cmd)
}

// DropCollection drops a collection
func (c *Client) DropCollection(ctx context.Context, name string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := fmt.Sprintf("DROP %s", name)
	return c.sendCommand(ctx, cmd)
}

// ListCollections lists all collections
func (c *Client) ListCollections(ctx context.Context) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	resp, err := c.sendCommandWithResponse(ctx, "LIST")
	if err != nil {
		return nil, newError("list", err)
	}

	var collections []string
	if err := json.Unmarshal([]byte(resp), &collections); err != nil {
		return nil, newError("list", err)
	}
	return collections, nil
}

// sendCommand sends a command without expecting a response body
func (c *Client) sendCommand(ctx context.Context, cmd string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
		return newError(cmd, err)
	}

	if _, err := c.conn.Write([]byte(cmd + "\n")); err != nil {
		return newError(cmd, err)
	}

	if err := c.conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
		return newError(cmd, err)
	}

	resp, err := c.reader.ReadString('\n')
	if err != nil {
		return newError(cmd, err)
	}
	resp = strings.TrimSpace(resp)

	if strings.HasPrefix(resp, "-ERR") {
		return newError(cmd, errors.New(strings.TrimPrefix(resp, "-ERR ")))
	}

	if strings.HasPrefix(resp, "-NOT_FOUND") {
		return newError(cmd, ErrKeyNotFound)
	}

	if strings.HasPrefix(resp, "+OK") {
		return nil
	}

	return newError(cmd, fmt.Errorf("unexpected response: %s", resp))
}

// sendCommandWithResponse sends a command and returns the response body
func (c *Client) sendCommandWithResponse(ctx context.Context, cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
		return "", newError(cmd, err)
	}

	if _, err := c.conn.Write([]byte(cmd + "\n")); err != nil {
		return "", newError(cmd, err)
	}

	if err := c.conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
		return "", newError(cmd, err)
	}

	resp, err := c.reader.ReadString('\n')
	if err != nil {
		return "", newError(cmd, err)
	}
	resp = strings.TrimSpace(resp)

	if strings.HasPrefix(resp, "-ERR") {
		return "", newError(cmd, errors.New(strings.TrimPrefix(resp, "-ERR ")))
	}

	if strings.HasPrefix(resp, "-NOT_FOUND") {
		return "", newError(cmd, ErrKeyNotFound)
	}

	if strings.HasPrefix(resp, "+") {
		return strings.TrimPrefix(resp, "+"), nil
	}

	if strings.HasPrefix(resp, "[") || strings.HasPrefix(resp, "{") {
		return resp, nil
	}

	return "", newError(cmd, fmt.Errorf("unexpected response: %s", resp))
}
