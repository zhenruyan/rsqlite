package rsqlite

import (
	"context"
	"database/sql/driver"
)

// Stmt implements the database/sql/driver.Stmt interface
type Stmt struct {
	conn  *Conn
	query string
}

// Close implements the database/sql/driver.Stmt interface
func (s *Stmt) Close() error {
	// Nothing to close for prepared statements in rqlite
	return nil
}

// NumInput implements the database/sql/driver.Stmt interface
func (s *Stmt) NumInput() int {
	// Return -1 to indicate that we don't know the number of parameters
	return -1
}

// Exec implements the database/sql/driver.Stmt interface
func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.ExecContext(context.Background(), convertToNamedValues(args))
}

// ExecContext implements the database/sql/driver.StmtExecContext interface
func (s *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return s.conn.ExecContext(ctx, s.query, args)
}

// Query implements the database/sql/driver.Stmt interface
func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.QueryContext(context.Background(), convertToNamedValues(args))
}

// QueryContext implements the database/sql/driver.StmtQueryContext interface
func (s *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return s.conn.QueryContext(ctx, s.query, args)
}

// convertToNamedValues converts []driver.Value to []driver.NamedValue
func convertToNamedValues(values []driver.Value) []driver.NamedValue {
	namedValues := make([]driver.NamedValue, len(values))
	for i, v := range values {
		namedValues[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   v,
		}
	}
	return namedValues
}
