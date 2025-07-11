package rsqlite

import (
	"database/sql/driver"
	"io"
	"reflect"
	"time"

	"github.com/rqlite/gorqlite"
)

// Result implements the database/sql/driver.Result interface
type Result struct {
	lastInsertID int64
	rowsAffected int64
}

// LastInsertId implements the database/sql/driver.Result interface
func (r *Result) LastInsertId() (int64, error) {
	return r.lastInsertID, nil
}

// RowsAffected implements the database/sql/driver.Result interface
func (r *Result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

// Rows implements the database/sql/driver.Rows interface
type Rows struct {
	result *gorqlite.QueryResult
	closed bool
}

// Columns implements the database/sql/driver.Rows interface
func (r *Rows) Columns() []string {
	if r.result == nil {
		return nil
	}
	return r.result.Columns()
}

// Close implements the database/sql/driver.Rows interface
func (r *Rows) Close() error {
	r.closed = true
	return nil
}

// Next implements the database/sql/driver.Rows interface
func (r *Rows) Next(dest []driver.Value) error {
	if r.closed || r.result == nil {
		return io.EOF
	}

	if !r.result.Next() {
		return io.EOF
	}

	// Create temporary slice to hold scanned values
	values := make([]interface{}, len(dest))

	// Scan the current row into our temporary slice
	if err := r.result.Scan(values...); err != nil {
		return err
	}

	// Convert each value to driver.Value
	for i, val := range values {
		dest[i] = convertValue(val)
	}

	return nil
}

// convertValue converts rqlite values to driver values
func convertValue(val interface{}) driver.Value {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case bool:
		return v
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case string:
		return v
	case []byte:
		return v
	case time.Time:
		return v
	default:
		// For other types, try to convert to string
		return convertToString(v)
	}
}

// convertToString converts unknown types to string
func convertToString(val interface{}) string {
	if val == nil {
		return ""
	}

	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return string(rune(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return string(rune(v.Uint()))
	case reflect.Float32, reflect.Float64:
		return string(rune(v.Float()))
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	default:
		return v.String()
	}
}
