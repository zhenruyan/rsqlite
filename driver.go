package rsqlite

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"time"
)

// Driver implements the database/sql/driver.Driver interface
type Driver struct{}

// Open returns a new connection to the database
func (d *Driver) Open(dsn string) (driver.Conn, error) {
	return Open(dsn)
}

// Config holds the configuration for the rqlite connection
type Config struct {
	Nodes            []string
	Username         string
	Password         string
	Timeout          time.Duration
	ConsistencyLevel string
}

// ParseDSN parses the data source name
func ParseDSN(dsn string) (*Config, error) {
	cfg := &Config{
		Timeout:          30 * time.Second,
		ConsistencyLevel: "weak",
	}

	// DSN format: rqlite://[username:password@]host1:port1,host2:port2/[?consistency=strong&timeout=30s]
	// But we register as sqlite, so DSN might be: sqlite://host1:port1,host2:port2/[?consistency=strong&timeout=30s]

	if dsn == "" {
		return nil, errors.New("dsn cannot be empty")
	}

	// Remove sqlite:// or rqlite:// prefix if present
	dsn = strings.TrimPrefix(dsn, "sqlite://")
	dsn = strings.TrimPrefix(dsn, "rqlite://")

	// Parse authentication if present
	if strings.Contains(dsn, "@") {
		parts := strings.Split(dsn, "@")
		if len(parts) != 2 {
			return nil, errors.New("invalid dsn format")
		}

		authPart := parts[0]
		dsn = parts[1]

		if strings.Contains(authPart, ":") {
			authParts := strings.Split(authPart, ":")
			cfg.Username = authParts[0]
			cfg.Password = authParts[1]
		} else {
			cfg.Username = authPart
		}
	}

	// Parse parameters if present
	if strings.Contains(dsn, "?") {
		parts := strings.Split(dsn, "?")
		if len(parts) != 2 {
			return nil, errors.New("invalid dsn format")
		}

		dsn = parts[0]
		params := parts[1]

		for _, param := range strings.Split(params, "&") {
			kv := strings.Split(param, "=")
			if len(kv) != 2 {
				continue
			}

			key, value := kv[0], kv[1]
			switch key {
			case "consistency":
				cfg.ConsistencyLevel = value
			case "timeout":
				if timeout, err := time.ParseDuration(value); err == nil {
					cfg.Timeout = timeout
				}
			}
		}
	}

	// Parse nodes
	if dsn == "" {
		return nil, errors.New("no nodes specified")
	}

	nodes := strings.Split(dsn, ",")
	for _, node := range nodes {
		node = strings.TrimSpace(node)
		if node != "" {
			// Add http:// prefix if not present
			if !strings.HasPrefix(node, "http://") && !strings.HasPrefix(node, "https://") {
				node = "http://" + node
			}
			cfg.Nodes = append(cfg.Nodes, node)
		}
	}

	if len(cfg.Nodes) == 0 {
		return nil, errors.New("no valid nodes found")
	}

	return cfg, nil
}

// Open creates a new connection
func Open(dsn string) (driver.Conn, error) {
	cfg, err := ParseDSN(dsn)
	if err != nil {
		return nil, err
	}

	return NewConn(cfg)
}

func init() {
	sql.Register("sqlite", &Driver{})
	sql.Register("rqlite", &Driver{})
}
