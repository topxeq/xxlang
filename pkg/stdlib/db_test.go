// pkg/stdlib/db_test.go
// Tests for the database module.
package stdlib

import (
	"os"
	"testing"

	// Import database drivers for testing
	_ "github.com/glebarez/go-sqlite" // SQLite3 driver (pure Go)

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/topxeq/xxlang/pkg/objects"
)

// getDBModule returns the db module for testing.
func getDBModule() *Module {
	return Get("db")
}

// TestDBModuleRegistration tests that the db module is registered.
func TestDBModuleRegistration(t *testing.T) {
	module := getDBModule()
	require.NotNil(t, module, "db module should be registered")
	assert.Equal(t, "db", module.Name)
}

// TestDBDrivers tests that database drivers are available.
func TestDBDrivers(t *testing.T) {
	module := getDBModule()
	driversFn := module.Exports["drivers"]
	require.NotNil(t, driversFn, "drivers function should exist")

	builtin, ok := driversFn.(*objects.Builtin)
	require.True(t, ok, "drivers should be a builtin function")

	result := builtin.Fn()
	arr, ok := result.(*objects.Array)
	require.True(t, ok, "drivers should return an array")

	// Check that at least sqlite is available
	found := false
	for _, elem := range arr.Elements {
		if s, ok := elem.(*objects.String); ok && s.Value == "sqlite" {
			found = true
			break
		}
	}
	assert.True(t, found, "sqlite driver should be available")
}

// TestDBOpen tests opening a database connection.
func TestDBOpen(t *testing.T) {
	module := getDBModule()
	// Create a temporary database file
	tmpFile := "test_db_open.sqlite"
	defer os.Remove(tmpFile)

	openFn := module.Exports["open"]
	require.NotNil(t, openFn, "open function should exist")

	builtin, ok := openFn.(*objects.Builtin)
	require.True(t, ok, "open should be a builtin function")

	// Test opening SQLite database
	result := builtin.Fn(String("sqlite"), String(tmpFile))
	db, ok := result.(*objects.DB)
	require.True(t, ok, "open should return a DB object")
	assert.NotNil(t, db.Value)
	assert.Equal(t, "sqlite", db.DriverName)
	assert.False(t, db.IsClosed())

	// Close the database
	closeFn := module.Exports["close"]
	require.NotNil(t, closeFn)
	closeBuiltin := closeFn.(*objects.Builtin)
	closeBuiltin.Fn(db)
	assert.True(t, db.IsClosed())
}

// TestDBExecAndQuery tests executing and querying data.
func TestDBExecAndQuery(t *testing.T) {
	module := getDBModule()
	// Create a temporary database file
	tmpFile := "test_db_exec.sqlite"
	defer os.Remove(tmpFile)

	// Open database
	openFn := module.Exports["open"].(*objects.Builtin)
	result := openFn.Fn(String("sqlite"), String(tmpFile))
	db, ok := result.(*objects.DB)
	require.True(t, ok)
	defer db.Close()

	// Create table
	execFn := module.Exports["exec"].(*objects.Builtin)
	result = execFn.Fn(db, String("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)"))
	m, ok := result.(*objects.OrderedMap)
	require.True(t, ok)
	lastInsertId := m.Get(&objects.String{Value: "lastInsertId"})
	assert.Equal(t, int64(0), lastInsertId.(*objects.Int).Value)
	rowsAffected := m.Get(&objects.String{Value: "rowsAffected"})
	assert.Equal(t, int64(0), rowsAffected.(*objects.Int).Value)

	// Insert data
	result = execFn.Fn(db, String("INSERT INTO users (name, age) VALUES (?, ?)"), String("Alice"), Int(30))
	// Debug: check if result is an error
	if errObj, ok := result.(*objects.Error); ok {
		t.Fatalf("Insert failed: %s", errObj.Message)
	}
	m, ok = result.(*objects.OrderedMap)
	require.True(t, ok, "result should be an OrderedMap, got %T: %s", result, result.Inspect())
	lastInsertId = m.Get(&objects.String{Value: "lastInsertId"})
	assert.True(t, lastInsertId.(*objects.Int).Value > 0)
	rowsAffected = m.Get(&objects.String{Value: "rowsAffected"})
	assert.Equal(t, int64(1), rowsAffected.(*objects.Int).Value)

	// Insert more data
	execFn.Fn(db, String("INSERT INTO users (name, age) VALUES (?, ?)"), String("Bob"), Int(25))

	// Query data
	queryFn := module.Exports["query"].(*objects.Builtin)
	result = queryFn.Fn(db, String("SELECT * FROM users ORDER BY id"))
	rows, ok := result.(*objects.Array)
	require.True(t, ok)
	assert.Equal(t, 2, len(rows.Elements))

	// Check first row
	row1, ok := rows.Elements[0].(*objects.Map)
	require.True(t, ok)
	assert.NotNil(t, row1.Pairs)

	// Query with condition
	result = queryFn.Fn(db, String("SELECT name, age FROM users WHERE age > ?"), Int(26))
	rows, ok = result.(*objects.Array)
	require.True(t, ok)
	assert.Equal(t, 1, len(rows.Elements)) // Only Alice (age 30)

	// Query single row
	queryRowFn := module.Exports["queryRow"].(*objects.Builtin)
	result = queryRowFn.Fn(db, String("SELECT * FROM users WHERE name = ?"), String("Alice"))
	row, ok := result.(*objects.Map)
	require.True(t, ok)
	assert.NotNil(t, row.Pairs)

	// Query non-existent row
	result = queryRowFn.Fn(db, String("SELECT * FROM users WHERE name = ?"), String("Unknown"))
	assert.Equal(t, objects.NULL, result)
}

// TestDBTransaction tests transaction operations.
func TestDBTransaction(t *testing.T) {
	module := getDBModule()
	tmpFile := "test_db_tx.sqlite"
	defer os.Remove(tmpFile)

	// Open database
	openFn := module.Exports["open"].(*objects.Builtin)
	result := openFn.Fn(String("sqlite"), String(tmpFile))
	db, ok := result.(*objects.DB)
	require.True(t, ok)
	defer db.Close()

	// Create table
	execFn := module.Exports["exec"].(*objects.Builtin)
	execFn.Fn(db, String("CREATE TABLE accounts (id INTEGER PRIMARY KEY, balance INTEGER)"))
	execFn.Fn(db, String("INSERT INTO accounts (balance) VALUES (100)"))

	// Begin transaction
	beginFn := module.Exports["begin"].(*objects.Builtin)
	result = beginFn.Fn(db)
	tx, ok := result.(*objects.DBTx)
	require.True(t, ok)
	assert.False(t, tx.IsClosed())

	// Execute within transaction
	txExecFn := module.Exports["txExec"].(*objects.Builtin)
	result = txExecFn.Fn(tx, String("UPDATE accounts SET balance = balance - 50 WHERE id = 1"))
	assert.NotNil(t, result)

	// Commit transaction
	commitFn := module.Exports["commit"].(*objects.Builtin)
	result = commitFn.Fn(tx)
	assert.Equal(t, objects.NULL, result)
	assert.True(t, tx.IsClosed())

	// Verify the change
	queryFn := module.Exports["query"].(*objects.Builtin)
	result = queryFn.Fn(db, String("SELECT balance FROM accounts WHERE id = 1"))
	rows := result.(*objects.Array)
	row := rows.Elements[0].(*objects.Map)
	for _, pair := range row.Pairs {
		if key, ok := pair.Key.(*objects.String); ok && key.Value == "balance" {
			assert.Equal(t, int64(50), pair.Value.(*objects.Int).Value)
			break
		}
	}
}

// TestDBRollback tests transaction rollback.
func TestDBRollback(t *testing.T) {
	module := getDBModule()
	tmpFile := "test_db_rollback.sqlite"
	defer os.Remove(tmpFile)

	// Open database
	openFn := module.Exports["open"].(*objects.Builtin)
	result := openFn.Fn(String("sqlite"), String(tmpFile))
	db, ok := result.(*objects.DB)
	require.True(t, ok)
	defer db.Close()

	// Create table
	execFn := module.Exports["exec"].(*objects.Builtin)
	execFn.Fn(db, String("CREATE TABLE items (id INTEGER PRIMARY KEY, value TEXT)"))
	execFn.Fn(db, String("INSERT INTO items (value) VALUES ('original')"))

	// Begin transaction
	beginFn := module.Exports["begin"].(*objects.Builtin)
	result = beginFn.Fn(db)
	tx, ok := result.(*objects.DBTx)
	require.True(t, ok)

	// Update within transaction
	txExecFn := module.Exports["txExec"].(*objects.Builtin)
	txExecFn.Fn(tx, String("UPDATE items SET value = 'modified' WHERE id = 1"))

	// Rollback transaction
	rollbackFn := module.Exports["rollback"].(*objects.Builtin)
	result = rollbackFn.Fn(tx)
	assert.Equal(t, objects.NULL, result)
	assert.True(t, tx.IsClosed())

	// Verify the original value is preserved
	queryFn := module.Exports["query"].(*objects.Builtin)
	result = queryFn.Fn(db, String("SELECT value FROM items WHERE id = 1"))
	rows := result.(*objects.Array)
	row := rows.Elements[0].(*objects.Map)
	for _, pair := range row.Pairs {
		if key, ok := pair.Key.(*objects.String); ok && key.Value == "value" {
			assert.Equal(t, "original", pair.Value.(*objects.String).Value)
			break
		}
	}
}

// TestDBPreparedStatement tests prepared statements.
func TestDBPreparedStatement(t *testing.T) {
	module := getDBModule()
	tmpFile := "test_db_stmt.sqlite"
	defer os.Remove(tmpFile)

	// Open database
	openFn := module.Exports["open"].(*objects.Builtin)
	result := openFn.Fn(String("sqlite"), String(tmpFile))
	db, ok := result.(*objects.DB)
	require.True(t, ok)
	defer db.Close()

	// Create table
	execFn := module.Exports["exec"].(*objects.Builtin)
	execFn.Fn(db, String("CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT, price REAL)"))

	// Prepare insert statement
	prepareFn := module.Exports["prepare"].(*objects.Builtin)
	result = prepareFn.Fn(db, String("INSERT INTO products (name, price) VALUES (?, ?)"))
	stmt, ok := result.(*objects.DBStmt)
	require.True(t, ok)
	assert.False(t, stmt.IsClosed())

	// Execute prepared statement multiple times
	stmtExecFn := module.Exports["stmtExec"].(*objects.Builtin)
	result = stmtExecFn.Fn(stmt, String("Apple"), Float(1.99))
	assert.NotNil(t, result)

	result = stmtExecFn.Fn(stmt, String("Banana"), Float(0.99))
	assert.NotNil(t, result)

	// Close statement
	closeFn := module.Exports["close"].(*objects.Builtin)
	closeFn.Fn(stmt)
	assert.True(t, stmt.IsClosed())

	// Query to verify
	queryFn := module.Exports["query"].(*objects.Builtin)
	result = queryFn.Fn(db, String("SELECT * FROM products ORDER BY name"))
	rows := result.(*objects.Array)
	assert.Equal(t, 2, len(rows.Elements))
}

// TestDBQueryArrays tests queryArrays function.
func TestDBQueryArrays(t *testing.T) {
	module := getDBModule()
	tmpFile := "test_db_arrays.sqlite"
	defer os.Remove(tmpFile)

	// Open database
	openFn := module.Exports["open"].(*objects.Builtin)
	result := openFn.Fn(String("sqlite"), String(tmpFile))
	db, ok := result.(*objects.DB)
	require.True(t, ok)
	defer db.Close()

	// Create and populate table
	execFn := module.Exports["exec"].(*objects.Builtin)
	execFn.Fn(db, String("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)"))
	execFn.Fn(db, String("INSERT INTO test (name) VALUES ('a'), ('b'), ('c')"))

	// Query as arrays
	queryArraysFn := module.Exports["queryArrays"].(*objects.Builtin)
	result = queryArraysFn.Fn(db, String("SELECT name FROM test ORDER BY id"))
	rows, ok := result.(*objects.Array)
	require.True(t, ok)
	assert.Equal(t, 3, len(rows.Elements))

	// Check each row is an array
	for i, row := range rows.Elements {
		arr, ok := row.(*objects.Array)
		require.True(t, ok, "row %d should be an array", i)
		assert.Equal(t, 1, len(arr.Elements))
	}
}

// TestDBIsConnected tests the isConnected function.
func TestDBIsConnected(t *testing.T) {
	module := getDBModule()
	tmpFile := "test_db_connected.sqlite"
	defer os.Remove(tmpFile)

	// Open database
	openFn := module.Exports["open"].(*objects.Builtin)
	result := openFn.Fn(String("sqlite"), String(tmpFile))
	db, ok := result.(*objects.DB)
	require.True(t, ok)

	// Check connected
	isConnectedFn := module.Exports["isConnected"].(*objects.Builtin)
	result = isConnectedFn.Fn(db)
	assert.Equal(t, objects.TRUE, result)

	// Close database
	db.Close()

	// Check not connected
	result = isConnectedFn.Fn(db)
	assert.Equal(t, objects.FALSE, result)
}

// TestDBSetConnectionPool tests connection pool settings.
func TestDBSetConnectionPool(t *testing.T) {
	module := getDBModule()
	tmpFile := "test_db_pool.sqlite"
	defer os.Remove(tmpFile)

	// Open database
	openFn := module.Exports["open"].(*objects.Builtin)
	result := openFn.Fn(String("sqlite"), String(tmpFile))
	db, ok := result.(*objects.DB)
	require.True(t, ok)
	defer db.Close()

	// Set connection pool settings
	setMaxOpenFn := module.Exports["setMaxOpenConns"].(*objects.Builtin)
	result = setMaxOpenFn.Fn(db, Int(10))
	assert.Equal(t, objects.NULL, result)

	setMaxIdleFn := module.Exports["setMaxIdleConns"].(*objects.Builtin)
	result = setMaxIdleFn.Fn(db, Int(5))
	assert.Equal(t, objects.NULL, result)

	setLifetimeFn := module.Exports["setConnMaxLifetime"].(*objects.Builtin)
	result = setLifetimeFn.Fn(db, Int(3600)) // 1 hour
	assert.Equal(t, objects.NULL, result)
}

// TestDBStats tests the stats function.
func TestDBStats(t *testing.T) {
	module := getDBModule()
	tmpFile := "test_db_stats.sqlite"
	defer os.Remove(tmpFile)

	// Open database
	openFn := module.Exports["open"].(*objects.Builtin)
	result := openFn.Fn(String("sqlite"), String(tmpFile))
	db, ok := result.(*objects.DB)
	require.True(t, ok)
	defer db.Close()

	// Get stats
	statsFn := module.Exports["stats"].(*objects.Builtin)
	result = statsFn.Fn(db)
	stats, ok := result.(*objects.Map)
	require.True(t, ok)
	assert.NotNil(t, stats.Pairs)
}

// TestDBOpenWithoutPing tests opening without ping.
func TestDBOpenWithoutPing(t *testing.T) {
	module := getDBModule()
	tmpFile := "test_db_noping.sqlite"
	defer os.Remove(tmpFile)

	// Open database without ping
	openWithoutPingFn := module.Exports["openWithoutPing"].(*objects.Builtin)
	result := openWithoutPingFn.Fn(String("sqlite"), String(tmpFile))
	db, ok := result.(*objects.DB)
	require.True(t, ok)
	assert.NotNil(t, db.Value)
	defer db.Close()
}

// TestDBPing tests the ping function.
func TestDBPing(t *testing.T) {
	module := getDBModule()
	tmpFile := "test_db_ping.sqlite"
	defer os.Remove(tmpFile)

	// Open database
	openFn := module.Exports["open"].(*objects.Builtin)
	result := openFn.Fn(String("sqlite"), String(tmpFile))
	db, ok := result.(*objects.DB)
	require.True(t, ok)
	defer db.Close()

	// Ping database
	pingFn := module.Exports["ping"].(*objects.Builtin)
	result = pingFn.Fn(db)
	assert.Equal(t, objects.NULL, result)
}

// TestDBGetMember tests DB object member access.
func TestDBGetMember(t *testing.T) {
	module := getDBModule()
	tmpFile := "test_db_member.sqlite"
	defer os.Remove(tmpFile)

	// Open database
	openFn := module.Exports["open"].(*objects.Builtin)
	result := openFn.Fn(String("sqlite"), String(tmpFile))
	db, ok := result.(*objects.DB)
	require.True(t, ok)
	defer db.Close()

	// Test member access
	driver := db.GetMember("driver")
	assert.Equal(t, "sqlite", driver.(*objects.String).Value)

	closed := db.GetMember("closed")
	assert.Equal(t, objects.FALSE, closed)

	stats := db.GetMember("stats")
	assert.NotNil(t, stats)
}
