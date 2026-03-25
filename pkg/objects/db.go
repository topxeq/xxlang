// pkg/objects/db.go
// Database object types for Xxlang.
// Provides database connection, transaction, rows, and statement objects.
package objects

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// DB represents a database connection.
// It wraps the standard sql.DB and provides methods for querying and executing SQL.
type DB struct {
	// Value is the underlying Go sql.DB connection
	Value *sql.DB
	// DriverName is the name of the database driver (e.g., "sqlite3", "mysql")
	DriverName string
	// DataSourceName is the connection string
	DataSourceName string
	// mu protects concurrent access
	mu sync.Mutex
	// closed indicates if the connection has been closed
	closed bool
}

// Type returns the object type.
func (db *DB) Type() ObjectType { return DBType }

// TypeTag returns the type tag for fast type checking.
func (db *DB) TypeTag() TypeTag { return TagDB }

// Inspect returns a string representation of the database connection.
func (db *DB) Inspect() string {
	if db.Value == nil {
		return "[db nil]"
	}
	return fmt.Sprintf("[db %s]", db.DriverName)
}

// ToBool converts the DB to a boolean (true if connected).
func (db *DB) ToBool() *Bool {
	db.mu.Lock()
	defer db.mu.Unlock()
	return &Bool{Value: db.Value != nil && !db.closed}
}

// HashKey returns a hash key for the DB.
func (db *DB) HashKey() HashKey {
	return HashKey{Type: DBType, Value: 0}
}

// IsClosed returns whether the connection is closed.
func (db *DB) IsClosed() bool {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.closed
}

// Close closes the database connection.
func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.closed || db.Value == nil {
		return nil
	}
	db.closed = true
	return db.Value.Close()
}

// GetMember returns a member by name for script access.
func (db *DB) GetMember(name string) Object {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.Value == nil || db.closed {
		return &Error{Message: "database connection is closed"}
	}

	switch name {
	case "driver":
		return NewString(db.DriverName)
	case "closed":
		return &Bool{Value: db.closed}
	case "stats":
		return db.getStats()
	}

	return NULL
}

// getStats returns database connection statistics.
func (db *DB) getStats() Object {
	stats := db.Value.Stats()
	pairs := make(map[HashKey]MapPair)

	pairs[NewString("maxOpenConnections").HashKey()] = MapPair{
		Key:   NewString("maxOpenConnections"),
		Value: NewInt(int64(stats.MaxOpenConnections)),
	}
	pairs[NewString("openConnections").HashKey()] = MapPair{
		Key:   NewString("openConnections"),
		Value: NewInt(int64(stats.OpenConnections)),
	}
	pairs[NewString("inUse").HashKey()] = MapPair{
		Key:   NewString("inUse"),
		Value: NewInt(int64(stats.InUse)),
	}
	pairs[NewString("idle").HashKey()] = MapPair{
		Key:   NewString("idle"),
		Value: NewInt(int64(stats.Idle)),
	}
	pairs[NewString("waitCount").HashKey()] = MapPair{
		Key:   NewString("waitCount"),
		Value: NewInt(stats.WaitCount),
	}
	pairs[NewString("waitDuration").HashKey()] = MapPair{
		Key:   NewString("waitDuration"),
		Value: NewInt(stats.WaitDuration.Milliseconds()),
	}

	return NewMap(pairs)
}

// NewDB creates a new DB object from an sql.DB connection.
func NewDB(db *sql.DB, driverName, dataSourceName string) *DB {
	return &DB{
		Value:          db,
		DriverName:     driverName,
		DataSourceName: dataSourceName,
		closed:         false,
	}
}

// DBTx represents a database transaction.
// It provides methods for transactional operations.
type DBTx struct {
	// Value is the underlying Go sql.Tx
	Value *sql.Tx
	// DB is the parent database connection
	DB *DB
	// mu protects concurrent access
	mu sync.Mutex
	// closed indicates if the transaction has been committed or rolled back
	closed bool
}

// Type returns the object type.
func (tx *DBTx) Type() ObjectType { return DBTxType }

// TypeTag returns the type tag for fast type checking.
func (tx *DBTx) TypeTag() TypeTag { return TagDBTx }

// Inspect returns a string representation of the transaction.
func (tx *DBTx) Inspect() string {
	return "[db_tx]"
}

// ToBool converts the transaction to a boolean (true if active).
func (tx *DBTx) ToBool() *Bool {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	return &Bool{Value: tx.Value != nil && !tx.closed}
}

// HashKey returns a hash key for the transaction.
func (tx *DBTx) HashKey() HashKey {
	return HashKey{Type: DBTxType, Value: 0}
}

// IsClosed returns whether the transaction is closed.
func (tx *DBTx) IsClosed() bool {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	return tx.closed
}

// MarkClosed marks the transaction as closed.
func (tx *DBTx) MarkClosed() {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	tx.closed = true
}

// GetMember returns a member by name for script access.
func (tx *DBTx) GetMember(name string) Object {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.Value == nil || tx.closed {
		return &Error{Message: "transaction is closed"}
	}

	switch name {
	case "closed":
		return &Bool{Value: tx.closed}
	}

	return NULL
}

// NewDBTx creates a new DBTx object from an sql.Tx.
func NewDBTx(tx *sql.Tx, db *DB) *DBTx {
	return &DBTx{
		Value:  tx,
		DB:     db,
		closed: false,
	}
}

// DBRows represents database query results.
// It wraps sql.Rows and provides iteration methods.
type DBRows struct {
	// Value is the underlying Go sql.Rows
	Value *sql.Rows
	// mu protects concurrent access
	mu sync.Mutex
	// closed indicates if the rows have been closed
	closed bool
	// columns caches the column names
	columns []string
	// columnTypes caches the column types
	columnTypes []*sql.ColumnType
}

// Type returns the object type.
func (r *DBRows) Type() ObjectType { return DBRowsType }

// TypeTag returns the type tag for fast type checking.
func (r *DBRows) TypeTag() TypeTag { return TagDBRows }

// Inspect returns a string representation of the rows.
func (r *DBRows) Inspect() string {
	return "[db_rows]"
}

// ToBool converts the rows to a boolean (true if not closed).
func (r *DBRows) ToBool() *Bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return &Bool{Value: r.Value != nil && !r.closed}
}

// HashKey returns a hash key for the rows.
func (r *DBRows) HashKey() HashKey {
	return HashKey{Type: DBRowsType, Value: 0}
}

// IsClosed returns whether the rows are closed.
func (r *DBRows) IsClosed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.closed
}

// Close closes the rows.
func (r *DBRows) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed || r.Value == nil {
		return nil
	}
	r.closed = true
	return r.Value.Close()
}

// Columns returns the column names.
func (r *DBRows) Columns() ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.columns != nil {
		return r.columns, nil
	}

	if r.Value == nil || r.closed {
		return nil, fmt.Errorf("rows are closed")
	}

	cols, err := r.Value.Columns()
	if err != nil {
		return nil, err
	}
	r.columns = cols
	return cols, nil
}

// GetMember returns a member by name for script access.
func (r *DBRows) GetMember(name string) Object {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Value == nil || r.closed {
		return &Error{Message: "rows are closed"}
	}

	switch name {
	case "closed":
		return &Bool{Value: r.closed}
	case "columns":
		return r.getColumns()
	}

	return NULL
}

// getColumns returns column names as an array.
func (r *DBRows) getColumns() Object {
	if r.columns == nil {
		cols, err := r.Value.Columns()
		if err != nil {
			return &Error{Message: fmt.Sprintf("failed to get columns: %v", err)}
		}
		r.columns = cols
	}

	elements := make([]Object, len(r.columns))
	for i, col := range r.columns {
		elements[i] = NewString(col)
	}
	return NewArray(elements)
}

// NewDBRows creates a new DBRows object from sql.Rows.
func NewDBRows(rows *sql.Rows) *DBRows {
	return &DBRows{
		Value:  rows,
		closed: false,
	}
}

// DBStmt represents a prepared statement.
// It wraps sql.Stmt and provides execution methods.
type DBStmt struct {
	// Value is the underlying Go sql.Stmt
	Value *sql.Stmt
	// DB is the parent database connection
	DB *DB
	// mu protects concurrent access
	mu sync.Mutex
	// closed indicates if the statement has been closed
	closed bool
}

// Type returns the object type.
func (s *DBStmt) Type() ObjectType { return DBStmtType }

// TypeTag returns the type tag for fast type checking.
func (s *DBStmt) TypeTag() TypeTag { return TagDBStmt }

// Inspect returns a string representation of the statement.
func (s *DBStmt) Inspect() string {
	return "[db_stmt]"
}

// ToBool converts the statement to a boolean (true if not closed).
func (s *DBStmt) ToBool() *Bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &Bool{Value: s.Value != nil && !s.closed}
}

// HashKey returns a hash key for the statement.
func (s *DBStmt) HashKey() HashKey {
	return HashKey{Type: DBStmtType, Value: 0}
}

// IsClosed returns whether the statement is closed.
func (s *DBStmt) IsClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

// Close closes the statement.
func (s *DBStmt) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.Value == nil {
		return nil
	}
	s.closed = true
	return s.Value.Close()
}

// GetMember returns a member by name for script access.
func (s *DBStmt) GetMember(name string) Object {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Value == nil || s.closed {
		return &Error{Message: "statement is closed"}
	}

	switch name {
	case "closed":
		return &Bool{Value: s.closed}
	}

	return NULL
}

// NewDBStmt creates a new DBStmt object from sql.Stmt.
func NewDBStmt(stmt *sql.Stmt, db *DB) *DBStmt {
	return &DBStmt{
		Value:  stmt,
		DB:     db,
		closed: false,
	}
}

// ScanRow scans a single row into a map.
// This is a helper function used by the db module.
func ScanRow(rows *sql.Rows, columns []string) (map[string]interface{}, error) {
	// Create a slice of interface{} to hold the values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Scan the row
	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	// Convert to map
	result := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		// Handle null values
		if val == nil {
			result[col] = nil
			continue
		}
		// Handle []byte as string
		if b, ok := val.([]byte); ok {
			result[col] = string(b)
		} else {
			result[col] = val
		}
	}

	return result, nil
}

// dbValueToObject converts a Go value to an Xxlang Object.
// This is a helper function for converting database values.
func dbValueToObject(val interface{}) Object {
	if val == nil {
		return NULL
	}

	switch v := val.(type) {
	case int:
		return NewInt(int64(v))
	case int64:
		return NewInt(v)
	case int32:
		return NewInt(int64(v))
	case int16:
		return NewInt(int64(v))
	case int8:
		return NewInt(int64(v))
	case uint:
		return NewInt(int64(v))
	case uint64:
		return NewInt(int64(v))
	case uint32:
		return NewInt(int64(v))
	case uint16:
		return NewInt(int64(v))
	case uint8:
		return NewInt(int64(v))
	case float32:
		return &Float{Value: float64(v)}
	case float64:
		return &Float{Value: v}
	case string:
		return NewString(v)
	case []byte:
		return NewString(string(v))
	case bool:
		return &Bool{Value: v}
	case time.Time:
		return NewString(v.Format(time.RFC3339))
	case nil:
		return NULL
	default:
		return NewString(fmt.Sprintf("%v", v))
	}
}

// RowsToArrays converts sql.Rows to an array of arrays.
// Each row is an array of values in column order.
func RowsToArrays(rows *sql.Rows) (Object, error) {
	columns, err := rows.Columns()
	if err != nil {
		return &Error{Message: fmt.Sprintf("failed to get columns: %v", err)}, nil
	}

	var result []Object

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return &Error{Message: fmt.Sprintf("scan failed: %v", err)}, nil
		}

		rowArray := make([]Object, len(columns))
		for i, val := range values {
			rowArray[i] = dbValueToObject(val)
		}
		result = append(result, NewArray(rowArray))
	}

	if err := rows.Err(); err != nil {
		return &Error{Message: fmt.Sprintf("rows error: %v", err)}, nil
	}

	return NewArray(result), nil
}

// RowsToMaps converts sql.Rows to an array of maps.
// Each row is a map with column names as keys.
func RowsToMaps(rows *sql.Rows) (Object, error) {
	columns, err := rows.Columns()
	if err != nil {
		return &Error{Message: fmt.Sprintf("failed to get columns: %v", err)}, nil
	}

	var result []Object

	for rows.Next() {
		rowMap, err := ScanRow(rows, columns)
		if err != nil {
			return &Error{Message: fmt.Sprintf("scan failed: %v", err)}, nil
		}

		pairs := make(map[HashKey]MapPair)
		for col, val := range rowMap {
			key := NewString(col)
			pairs[key.HashKey()] = MapPair{
				Key:   key,
				Value: dbValueToObject(val),
			}
		}
		result = append(result, NewMap(pairs))
	}

	if err := rows.Err(); err != nil {
		return &Error{Message: fmt.Sprintf("rows error: %v", err)}, nil
	}

	return NewArray(result), nil
}
