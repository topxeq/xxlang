// pkg/stdlib/db.go
// Database module for Xxlang.
// Provides database connection, query, and execution capabilities.
// Supports SQLite, MySQL, PostgreSQL, Oracle, and MSSQL Server.
// All drivers are pure Go implementations (no CGO required).
//
// Driver names:
//   - SQLite:       "sqlite"
//   - MySQL:        "mysql"
//   - PostgreSQL:   "postgres"
//   - Oracle:       "oracle"
//   - MSSQL Server: "mssql"
package stdlib

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "db",
		Exports: map[string]objects.Object{
			// ============================================================
			// Connection Functions
			// ============================================================

			// open opens a database connection.
			// Usage: db = db.open(driverName, dataSourceName)
			// driverName: "sqlite", "mysql", "postgres", "oracle", "mssql"
			// dataSourceName: connection string specific to each driver
			// Example for SQLite:       db.open("sqlite", "./test.db")
			// Example for MySQL:        db.open("mysql", "user:password@tcp(localhost:3306)/dbname")
			// Example for PostgreSQL:   db.open("postgres", "host=localhost port=5432 user=postgres dbname=test sslmode=disable")
			// Example for Oracle:       db.open("oracle", "oracle://user:password@localhost:1521/service")
			// Example for MSSQL Server: db.open("mssql", "server=localhost;user id=sa;password=pass;database=test")
			"open": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("open() takes exactly 2 arguments: driverName and dataSourceName")
				}

				driverName, ok := args[0].(*objects.String)
				if !ok {
					return Error("open() driverName must be a string")
				}

				dataSourceName, ok := args[1].(*objects.String)
				if !ok {
					return Error("open() dataSourceName must be a string")
				}

				db, err := sql.Open(driverName.Value, dataSourceName.Value)
				if err != nil {
					return Error(fmt.Sprintf("open() failed: %s", err.Error()))
				}

				// Test connection
				if err := db.Ping(); err != nil {
					db.Close()
					return Error(fmt.Sprintf("open() ping failed: %s", err.Error()))
				}

				return objects.NewDB(db, driverName.Value, dataSourceName.Value)
			}),

			// openWithoutPing opens a database connection without testing.
			// Useful when the database might not be immediately available.
			// Usage: db = db.openWithoutPing(driverName, dataSourceName)
			"openWithoutPing": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("openWithoutPing() takes exactly 2 arguments: driverName and dataSourceName")
				}

				driverName, ok := args[0].(*objects.String)
				if !ok {
					return Error("openWithoutPing() driverName must be a string")
				}

				dataSourceName, ok := args[1].(*objects.String)
				if !ok {
					return Error("openWithoutPing() dataSourceName must be a string")
				}

				db, err := sql.Open(driverName.Value, dataSourceName.Value)
				if err != nil {
					return Error(fmt.Sprintf("openWithoutPing() failed: %s", err.Error()))
				}

				return objects.NewDB(db, driverName.Value, dataSourceName.Value)
			}),

			// ============================================================
			// Connection Management Functions
			// ============================================================

			// close closes a database connection.
			// Usage: db.close(conn)
			"close": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("close() takes exactly 1 argument")
				}

				switch obj := args[0].(type) {
				case *objects.DB:
					if err := obj.Close(); err != nil {
						return Error(fmt.Sprintf("close() failed: %s", err.Error()))
					}
					return Null()
				case *objects.DBRows:
					if err := obj.Close(); err != nil {
						return Error(fmt.Sprintf("close() failed: %s", err.Error()))
					}
					return Null()
				case *objects.DBStmt:
					if err := obj.Close(); err != nil {
						return Error(fmt.Sprintf("close() failed: %s", err.Error()))
					}
					return Null()
				case *objects.DBTx:
					// Rollback if not committed
					if !obj.IsClosed() {
						obj.Value.Rollback()
					}
					return Null()
				default:
					return Error("close() requires a database object (DB, Rows, Stmt, or Tx)")
				}
			}),

			// ping tests the database connection.
			// Usage: err = db.ping(conn)
			"ping": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("ping() takes exactly 1 argument")
				}

				db, ok := args[0].(*objects.DB)
				if !ok {
					return Error("ping() requires a DB connection")
				}

				if db.IsClosed() {
					return Error("database connection is closed")
				}

				if err := db.Value.Ping(); err != nil {
					return Error(fmt.Sprintf("ping() failed: %s", err.Error()))
				}

				return Null()
			}),

			// setMaxOpenConns sets the maximum number of open connections.
			// Usage: db.setMaxOpenConns(conn, n)
			"setMaxOpenConns": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setMaxOpenConns() takes exactly 2 arguments")
				}

				db, ok := args[0].(*objects.DB)
				if !ok {
					return Error("setMaxOpenConns() first argument must be a DB connection")
				}

				n, ok := args[1].(*objects.Int)
				if !ok {
					return Error("setMaxOpenConns() second argument must be an integer")
				}

				if db.IsClosed() {
					return Error("database connection is closed")
				}

				db.Value.SetMaxOpenConns(int(n.Value))
				return Null()
			}),

			// setMaxIdleConns sets the maximum number of idle connections.
			// Usage: db.setMaxIdleConns(conn, n)
			"setMaxIdleConns": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setMaxIdleConns() takes exactly 2 arguments")
				}

				db, ok := args[0].(*objects.DB)
				if !ok {
					return Error("setMaxIdleConns() first argument must be a DB connection")
				}

				n, ok := args[1].(*objects.Int)
				if !ok {
					return Error("setMaxIdleConns() second argument must be an integer")
				}

				if db.IsClosed() {
					return Error("database connection is closed")
				}

				db.Value.SetMaxIdleConns(int(n.Value))
				return Null()
			}),

			// setConnMaxLifetime sets the maximum lifetime of connections.
			// Usage: db.setConnMaxLifetime(conn, seconds)
			"setConnMaxLifetime": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setConnMaxLifetime() takes exactly 2 arguments")
				}

				db, ok := args[0].(*objects.DB)
				if !ok {
					return Error("setConnMaxLifetime() first argument must be a DB connection")
				}

				seconds, ok := args[1].(*objects.Int)
				if !ok {
					return Error("setConnMaxLifetime() second argument must be an integer (seconds)")
				}

				if db.IsClosed() {
					return Error("database connection is closed")
				}

				db.Value.SetConnMaxLifetime(time.Duration(seconds.Value) * time.Second)
				return Null()
			}),

			// ============================================================
			// Query Functions
			// ============================================================

			// query executes a query that returns rows.
			// Returns an array of maps (each row is a map).
			// Usage: rows = db.query(conn, sql, [args...])
			// Example: rows = db.query(db, "SELECT * FROM users WHERE id = ?", 1)
			"query": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("query() takes at least 2 arguments: connection and SQL query")
				}

				db, ok := args[0].(*objects.DB)
				if !ok {
					return Error("query() first argument must be a DB connection")
				}

				if db.IsClosed() {
					return Error("database connection is closed")
				}

				query, ok := args[1].(*objects.String)
				if !ok {
					return Error("query() second argument must be a SQL string")
				}

				// Convert remaining arguments to interface{}
				queryArgs := make([]interface{}, len(args)-2)
				for i, arg := range args[2:] {
					queryArgs[i] = objectToGoValue(arg)
				}

				rows, err := db.Value.Query(query.Value, queryArgs...)
				if err != nil {
					return Error(fmt.Sprintf("query() failed: %s", err.Error()))
				}
				defer rows.Close()

				result, err := objects.RowsToMaps(rows)
				if err != nil {
					return Error(fmt.Sprintf("query failed: %s", err.Error()))
				}
				// Check if result is an error object
				if errObj, ok := result.(*objects.Error); ok {
					return errObj
				}

				return result
			}),

			// queryArrays executes a query and returns rows as arrays.
			// Each row is an array of values in column order.
			// Usage: rows = db.queryArrays(conn, sql, [args...])
			"queryArrays": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("queryArrays() takes at least 2 arguments: connection and SQL query")
				}

				db, ok := args[0].(*objects.DB)
				if !ok {
					return Error("queryArrays() first argument must be a DB connection")
				}

				if db.IsClosed() {
					return Error("database connection is closed")
				}

				query, ok := args[1].(*objects.String)
				if !ok {
					return Error("queryArrays() second argument must be a SQL string")
				}

				queryArgs := make([]interface{}, len(args)-2)
				for i, arg := range args[2:] {
					queryArgs[i] = objectToGoValue(arg)
				}

				rows, err := db.Value.Query(query.Value, queryArgs...)
				if err != nil {
					return Error(fmt.Sprintf("queryArrays() failed: %s", err.Error()))
				}
				defer rows.Close()

				result, err := objects.RowsToArrays(rows)
				if err != nil {
					return Error(fmt.Sprintf("queryArrays failed: %s", err.Error()))
				}
				// Check if result is an error object
				if errObj, ok := result.(*objects.Error); ok {
					return errObj
				}

				return result
			}),

			// queryRow executes a query that returns at most one row.
			// Returns a map or null if no row found.
			// Usage: row = db.queryRow(conn, sql, [args...])
			"queryRow": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("queryRow() takes at least 2 arguments: connection and SQL query")
				}

				db, ok := args[0].(*objects.DB)
				if !ok {
					return Error("queryRow() first argument must be a DB connection")
				}

				if db.IsClosed() {
					return Error("database connection is closed")
				}

				query, ok := args[1].(*objects.String)
				if !ok {
					return Error("queryRow() second argument must be a SQL string")
				}

				queryArgs := make([]interface{}, len(args)-2)
				for i, arg := range args[2:] {
					queryArgs[i] = objectToGoValue(arg)
				}

				rows, err := db.Value.Query(query.Value, queryArgs...)
				if err != nil {
					return Error(fmt.Sprintf("queryRow() failed: %s", err.Error()))
				}
				defer rows.Close()

				if !rows.Next() {
					return Null()
				}

				columns, err := rows.Columns()
				if err != nil {
					return Error(fmt.Sprintf("queryRow() failed to get columns: %s", err.Error()))
				}

				rowMap, err := objects.ScanRow(rows, columns)
				if err != nil {
					return Error(fmt.Sprintf("queryRow() scan failed: %s", err.Error()))
				}

				// Convert to Xxlang map
				pairs := make(map[objects.HashKey]objects.MapPair)
				for col, val := range rowMap {
					key := String(col)
					pairs[key.HashKey()] = objects.MapPair{
						Key:   key,
						Value: dbValueToObj(val),
					}
				}

				return &objects.Map{Pairs: pairs}
			}),

			// ============================================================
			// Execute Functions
			// ============================================================

			// exec executes a query without returning rows.
			// Returns [lastInsertId, rowsAffected].
			// Usage: result = db.exec(conn, sql, [args...])
			// Example: result = db.exec(db, "INSERT INTO users (name) VALUES (?)", "John")
			"exec": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("exec() takes at least 2 arguments: connection and SQL query")
				}

				db, ok := args[0].(*objects.DB)
				if !ok {
					return Error("exec() first argument must be a DB connection")
				}

				if db.IsClosed() {
					return Error("database connection is closed")
				}

				query, ok := args[1].(*objects.String)
				if !ok {
					return Error("exec() second argument must be a SQL string")
				}

				queryArgs := make([]interface{}, len(args)-2)
				for i, arg := range args[2:] {
					queryArgs[i] = objectToGoValue(arg)
				}

				result, err := db.Value.Exec(query.Value, queryArgs...)
				if err != nil {
					return Error(fmt.Sprintf("exec() failed: %s", err.Error()))
				}

				lastInsertId, _ := result.LastInsertId()
				rowsAffected, _ := result.RowsAffected()

				return Array(Int(lastInsertId), Int(rowsAffected))
			}),

			// ============================================================
			// Transaction Functions
			// ============================================================

			// begin starts a new transaction.
			// Usage: tx = db.begin(conn)
			"begin": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("begin() takes exactly 1 argument")
				}

				db, ok := args[0].(*objects.DB)
				if !ok {
					return Error("begin() requires a DB connection")
				}

				if db.IsClosed() {
					return Error("database connection is closed")
				}

				tx, err := db.Value.Begin()
				if err != nil {
					return Error(fmt.Sprintf("begin() failed: %s", err.Error()))
				}

				return objects.NewDBTx(tx, db)
			}),

			// commit commits a transaction.
			// Usage: db.commit(tx)
			"commit": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("commit() takes exactly 1 argument")
				}

				tx, ok := args[0].(*objects.DBTx)
				if !ok {
					return Error("commit() requires a transaction")
				}

				if tx.IsClosed() {
					return Error("transaction is already closed")
				}

				if err := tx.Value.Commit(); err != nil {
					return Error(fmt.Sprintf("commit() failed: %s", err.Error()))
				}

				tx.MarkClosed()
				return Null()
			}),

			// rollback rolls back a transaction.
			// Usage: db.rollback(tx)
			"rollback": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("rollback() takes exactly 1 argument")
				}

				tx, ok := args[0].(*objects.DBTx)
				if !ok {
					return Error("rollback() requires a transaction")
				}

				if tx.IsClosed() {
					return Error("transaction is already closed")
				}

				if err := tx.Value.Rollback(); err != nil {
					return Error(fmt.Sprintf("rollback() failed: %s", err.Error()))
				}

				tx.MarkClosed()
				return Null()
			}),

			// txExec executes a query within a transaction.
			// Returns [lastInsertId, rowsAffected].
			// Usage: result = db.txExec(tx, sql, [args...])
			"txExec": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("txExec() takes at least 2 arguments: transaction and SQL query")
				}

				tx, ok := args[0].(*objects.DBTx)
				if !ok {
					return Error("txExec() first argument must be a transaction")
				}

				if tx.IsClosed() {
					return Error("transaction is closed")
				}

				query, ok := args[1].(*objects.String)
				if !ok {
					return Error("txExec() second argument must be a SQL string")
				}

				queryArgs := make([]interface{}, len(args)-2)
				for i, arg := range args[2:] {
					queryArgs[i] = objectToGoValue(arg)
				}

				result, err := tx.Value.Exec(query.Value, queryArgs...)
				if err != nil {
					return Error(fmt.Sprintf("txExec() failed: %s", err.Error()))
				}

				lastInsertId, _ := result.LastInsertId()
				rowsAffected, _ := result.RowsAffected()

				return Array(Int(lastInsertId), Int(rowsAffected))
			}),

			// txQuery executes a query within a transaction.
			// Usage: rows = db.txQuery(tx, sql, [args...])
			"txQuery": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("txQuery() takes at least 2 arguments: transaction and SQL query")
				}

				tx, ok := args[0].(*objects.DBTx)
				if !ok {
					return Error("txQuery() first argument must be a transaction")
				}

				if tx.IsClosed() {
					return Error("transaction is closed")
				}

				query, ok := args[1].(*objects.String)
				if !ok {
					return Error("txQuery() second argument must be a SQL string")
				}

				queryArgs := make([]interface{}, len(args)-2)
				for i, arg := range args[2:] {
					queryArgs[i] = objectToGoValue(arg)
				}

				rows, err := tx.Value.Query(query.Value, queryArgs...)
				if err != nil {
					return Error(fmt.Sprintf("txQuery() failed: %s", err.Error()))
				}
				defer rows.Close()

				result, err := objects.RowsToMaps(rows)
				if err != nil {
					return Error(fmt.Sprintf("query failed: %s", err.Error()))
				}
				// Check if result is an error object
				if errObj, ok := result.(*objects.Error); ok {
					return errObj
				}

				return result
			}),

			// ============================================================
			// Prepared Statement Functions
			// ============================================================

			// prepare creates a prepared statement.
			// Usage: stmt = db.prepare(conn, sql)
			"prepare": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("prepare() takes exactly 2 arguments: connection and SQL query")
				}

				db, ok := args[0].(*objects.DB)
				if !ok {
					return Error("prepare() first argument must be a DB connection")
				}

				if db.IsClosed() {
					return Error("database connection is closed")
				}

				query, ok := args[1].(*objects.String)
				if !ok {
					return Error("prepare() second argument must be a SQL string")
				}

				stmt, err := db.Value.Prepare(query.Value)
				if err != nil {
					return Error(fmt.Sprintf("prepare() failed: %s", err.Error()))
				}

				return objects.NewDBStmt(stmt, db)
			}),

			// stmtExec executes a prepared statement.
			// Returns [lastInsertId, rowsAffected].
			// Usage: result = db.stmtExec(stmt, [args...])
			"stmtExec": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("stmtExec() takes at least 1 argument: statement")
				}

				stmt, ok := args[0].(*objects.DBStmt)
				if !ok {
					return Error("stmtExec() first argument must be a statement")
				}

				if stmt.IsClosed() {
					return Error("statement is closed")
				}

				queryArgs := make([]interface{}, len(args)-1)
				for i, arg := range args[1:] {
					queryArgs[i] = objectToGoValue(arg)
				}

				result, err := stmt.Value.Exec(queryArgs...)
				if err != nil {
					return Error(fmt.Sprintf("stmtExec() failed: %s", err.Error()))
				}

				lastInsertId, _ := result.LastInsertId()
				rowsAffected, _ := result.RowsAffected()

				return Array(Int(lastInsertId), Int(rowsAffected))
			}),

			// stmtQuery executes a prepared statement query.
			// Usage: rows = db.stmtQuery(stmt, [args...])
			"stmtQuery": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("stmtQuery() takes at least 1 argument: statement")
				}

				stmt, ok := args[0].(*objects.DBStmt)
				if !ok {
					return Error("stmtQuery() first argument must be a statement")
				}

				if stmt.IsClosed() {
					return Error("statement is closed")
				}

				queryArgs := make([]interface{}, len(args)-1)
				for i, arg := range args[1:] {
					queryArgs[i] = objectToGoValue(arg)
				}

				rows, err := stmt.Value.Query(queryArgs...)
				if err != nil {
					return Error(fmt.Sprintf("stmtQuery() failed: %s", err.Error()))
				}
				defer rows.Close()

				result, err := objects.RowsToMaps(rows)
				if err != nil {
					return Error(fmt.Sprintf("query failed: %s", err.Error()))
				}
				// Check if result is an error object
				if errObj, ok := result.(*objects.Error); ok {
					return errObj
				}

				return result
			}),

			// ============================================================
			// Utility Functions
			// ============================================================

			// drivers returns a list of registered database drivers.
			// Usage: driverList = db.drivers()
			"drivers": BuiltinFunc(func(args ...objects.Object) objects.Object {
				drivers := sql.Drivers()
				elements := make([]objects.Object, len(drivers))
				for i, d := range drivers {
					elements[i] = String(d)
				}
				return Array(elements...)
			}),

			// isConnected checks if a database connection is open.
			// Usage: connected = db.isConnected(conn)
			"isConnected": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isConnected() takes exactly 1 argument")
				}

				db, ok := args[0].(*objects.DB)
				if !ok {
					return Error("isConnected() requires a DB connection")
				}

				return Bool(!db.IsClosed())
			}),

			// stats returns database connection statistics.
			// Usage: stats = db.stats(conn)
			"stats": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("stats() takes exactly 1 argument")
				}

				db, ok := args[0].(*objects.DB)
				if !ok {
					return Error("stats() requires a DB connection")
				}

				return db.GetMember("stats")
			}),

			// columns returns column names from a rows object.
			// Usage: cols = db.columns(rows)
			"columns": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("columns() takes exactly 1 argument")
				}

				rows, ok := args[0].(*objects.DBRows)
				if !ok {
					return Error("columns() requires a Rows object")
				}

				return rows.GetMember("columns")
			}),
		},
	})
}

// dbValueToObj converts a database value to an Xxlang Object.
func dbValueToObj(val interface{}) objects.Object {
	if val == nil {
		return Null()
	}

	switch v := val.(type) {
	case int:
		return Int(int64(v))
	case int64:
		return Int(v)
	case int32:
		return Int(int64(v))
	case int16:
		return Int(int64(v))
	case int8:
		return Int(int64(v))
	case uint:
		return Int(int64(v))
	case uint64:
		return Int(int64(v))
	case uint32:
		return Int(int64(v))
	case uint16:
		return Int(int64(v))
	case uint8:
		return Int(int64(v))
	case float32:
		return Float(float64(v))
	case float64:
		return Float(v)
	case string:
		return String(v)
	case []byte:
		return String(string(v))
	case bool:
		return Bool(v)
	default:
		return String(fmt.Sprintf("%v", v))
	}
}
