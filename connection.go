package rsqlite

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"sync"

	"github.com/rqlite/gorqlite"
)

// Conn implements the database/sql/driver.Conn interface
type Conn struct {
	cfg            *Config
	client         *gorqlite.Connection
	mu             sync.RWMutex
	closed         bool
	clusterManager *ClusterManager
}

// NewConn creates a new connection
func NewConn(cfg *Config) (*Conn, error) {
	conn := &Conn{
		cfg:            cfg,
		clusterManager: NewClusterManager(cfg.Nodes),
	}

	err := conn.connect()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// connect establishes connection to rqlite cluster
func (c *Conn) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return errors.New("connection is closed")
	}

	// Discover leader first
	ctx := context.Background()
	err := c.clusterManager.DiscoverLeader(ctx)
	if err != nil {
		// If discovery fails, try connecting to original nodes
		return c.connectToAnyNode()
	}

	// Try to connect to the leader
	leader := c.clusterManager.SelectBestNode(c.cfg.ConsistencyLevel)
	if leader != "" {
		client, err := c.createClient(leader)
		if err == nil {
			c.client = client
			return nil
		}
	}

	// Fallback to connecting to any available node
	return c.connectToAnyNode()
}

// connectToAnyNode tries to connect to any available node
func (c *Conn) connectToAnyNode() error {
	nodes := c.clusterManager.GetAllNodes()
	if len(nodes) == 0 {
		nodes = c.cfg.Nodes
	}

	var lastErr error
	for _, node := range nodes {
		client, err := c.createClient(node)
		if err != nil {
			lastErr = err
			continue
		}

		c.client = client
		return nil
	}

	if lastErr != nil {
		return fmt.Errorf("failed to connect to any node: %w", lastErr)
	}

	return errors.New("no nodes available")
}

// createClient creates a new rqlite client for the given node
func (c *Conn) createClient(node string) (*gorqlite.Connection, error) {
	client, err := gorqlite.Open(node)
	if err != nil {
		return nil, err
	}

	// Set authentication if provided
	// Note: gorqlite handles authentication differently
	// This would need to be implemented based on the actual gorqlite API

	// Set consistency level
	switch c.cfg.ConsistencyLevel {
	case "strong":
		client.SetConsistencyLevel(gorqlite.ConsistencyLevelStrong)
	case "weak":
		client.SetConsistencyLevel(gorqlite.ConsistencyLevelWeak)
	case "none":
		client.SetConsistencyLevel(gorqlite.ConsistencyLevelNone)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.Timeout)
	defer cancel()

	_, err = client.QueryOneContext(ctx, "SELECT 1")
	if err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

// reconnect attempts to reconnect to the cluster
func (c *Conn) reconnect() error {
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}

	return c.connect()
}

// Prepare implements the database/sql/driver.Conn interface
func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

// PrepareContext implements the database/sql/driver.ConnPrepareContext interface
func (c *Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, errors.New("connection is closed")
	}

	return &Stmt{
		conn:  c,
		query: query,
	}, nil
}

// Close implements the database/sql/driver.Conn interface
func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}

	return nil
}

// Begin implements the database/sql/driver.Conn interface
func (c *Conn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

// BeginTx implements the database/sql/driver.ConnBeginTx interface
func (c *Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	// rqlite doesn't support transactions in the traditional sense
	// We'll return a no-op transaction
	return &Tx{conn: c}, nil
}

// ExecContext implements the database/sql/driver.ExecerContext interface
func (c *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	if client == nil {
		return nil, errors.New("connection is closed")
	}

	// Convert named values to interface slice
	values := make([]interface{}, len(args))
	for i, arg := range args {
		values[i] = arg.Value
	}

	// Retry logic for leader changes
	for attempts := 0; attempts < 3; attempts++ {
		result, err := client.WriteOneParameterized(gorqlite.ParameterizedStatement{
			Query:     query,
			Arguments: values,
		})
		if err != nil {
			// If it's a leader change error, try to reconnect
			if attempts < 2 {
				c.mu.Lock()
				reconnectErr := c.reconnect()
				client = c.client
				c.mu.Unlock()
				if reconnectErr != nil {
					return nil, reconnectErr
				}
				continue
			}
			return nil, err
		}

		return &Result{
			lastInsertID: result.LastInsertID,
			rowsAffected: result.RowsAffected,
		}, nil
	}

	return nil, errors.New("max retry attempts exceeded")
}

// QueryContext implements the database/sql/driver.QueryerContext interface
func (c *Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	if client == nil {
		return nil, errors.New("connection is closed")
	}

	// Convert named values to interface slice
	values := make([]interface{}, len(args))
	for i, arg := range args {
		values[i] = arg.Value
	}

	// Retry logic for leader changes
	for attempts := 0; attempts < 3; attempts++ {
		result, err := client.QueryOneParameterized(gorqlite.ParameterizedStatement{
			Query:     query,
			Arguments: values,
		})
		if err != nil {
			// If it's a leader change error, try to reconnect
			if attempts < 2 {
				c.mu.Lock()
				reconnectErr := c.reconnect()
				client = c.client
				c.mu.Unlock()
				if reconnectErr != nil {
					return nil, reconnectErr
				}
				continue
			}
			return nil, err
		}

		return &Rows{
			result: &result,
			closed: false,
		}, nil
	}

	return nil, errors.New("max retry attempts exceeded")
}

// Ping implements the database/sql/driver.Pinger interface
func (c *Conn) Ping(ctx context.Context) error {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	if client == nil {
		return errors.New("connection is closed")
	}

	_, err := client.QueryOneContext(ctx, "SELECT 1")
	if err != nil {
		// Try to reconnect
		c.mu.Lock()
		reconnectErr := c.reconnect()
		c.mu.Unlock()
		if reconnectErr != nil {
			return reconnectErr
		}
		return nil
	}

	return nil
}

// Tx implements the database/sql/driver.Tx interface
type Tx struct {
	conn *Conn
}

// Commit implements the database/sql/driver.Tx interface
func (tx *Tx) Commit() error {
	// rqlite doesn't support transactions, so this is a no-op
	return nil
}

// Rollback implements the database/sql/driver.Tx interface
func (tx *Tx) Rollback() error {
	// rqlite doesn't support transactions, so this is a no-op
	return nil
}
