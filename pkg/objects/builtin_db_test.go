// pkg/objects/builtin_db_test.go
// Tests for database builtin functions.
package objects

import (
	"os"
	"testing"

	// Import database driver for testing
	_ "github.com/glebarez/go-sqlite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuiltinFormatSQLValue tests the formatSQLValue function.
func TestBuiltinFormatSQLValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"O'Brien", "O''Brien"},
		{"line1\nline2", "line1\\nline2"},
		{"carriage\rreturn", "carriage\\rreturn"},
		{"mixed'with\nnewline", "mixed''with\\nnewline"},
	}

	for _, tt := range tests {
		result := BuiltinFormatSQLValue.Fn(NewString(tt.input))
		str, ok := result.(*String)
		require.True(t, ok, "expected String, got %T", result)
		assert.Equal(t, tt.expected, str.Value)
	}

	// Test error case: wrong argument type
	result := BuiltinFormatSQLValue.Fn(NewInt(123))
	_, ok := result.(*Error)
	assert.True(t, ok, "expected Error for wrong argument type")

	// Test error case: wrong number of arguments
	result = BuiltinFormatSQLValue.Fn()
	_, ok = result.(*Error)
	assert.True(t, ok, "expected Error for wrong number of arguments")
}

// TestBuiltinDbConnect tests the dbConnect function.
func TestBuiltinDbConnect(t *testing.T) {
	tmpFile := "test_builtin_db_connect.sqlite"
	defer os.Remove(tmpFile)

	// Test successful connection
	result := BuiltinDbConnect.Fn(NewString("sqlite"), NewString(tmpFile))
	db, ok := result.(*DB)
	require.True(t, ok, "expected DB, got %T: %v", result, result)
	require.NotNil(t, db.Value)
	assert.Equal(t, "sqlite", db.DriverName)
	assert.False(t, db.IsClosed())

	// Clean up
	db.Close()

	// Test with sqlite3 alias
	result = BuiltinDbConnect.Fn(NewString("sqlite3"), NewString(tmpFile))
	db, ok = result.(*DB)
	require.True(t, ok)
	db.Close()
}

// TestBuiltinDbClose tests the dbClose function.
func TestBuiltinDbClose(t *testing.T) {
	tmpFile := "test_builtin_db_close.sqlite"
	defer os.Remove(tmpFile)

	// Open connection
	result := BuiltinDbConnect.Fn(NewString("sqlite"), NewString(tmpFile))
	db, ok := result.(*DB)
	require.True(t, ok)

	// Close connection
	result = BuiltinDbClose.Fn(db)
	assert.Equal(t, NULL, result)
	assert.True(t, db.IsClosed())

	// Test error case: wrong argument type
	result = BuiltinDbClose.Fn(NewString("not a db"))
	_, ok = result.(*Error)
	assert.True(t, ok, "expected Error for wrong argument type")
}

// TestBuiltinDbQuery tests the dbQuery function (string-based, Charlang compatible).
func TestBuiltinDbQuery(t *testing.T) {
	tmpFile := "test_builtin_db_query.sqlite"
	defer os.Remove(tmpFile)

	// Connect and setup
	result := BuiltinDbConnect.Fn(NewString("sqlite"), NewString(tmpFile))
	db, ok := result.(*DB)
	require.True(t, ok)
	defer db.Close()

	// Create table
	result = BuiltinDbExec.Fn(db, NewString("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER, salary REAL, active BOOLEAN)"))
	require.IsType(t, &Array{}, result)

	// Insert test data
	result = BuiltinDbExec.Fn(db, NewString("INSERT INTO users (name, age, salary, active) VALUES (?, ?, ?, ?)"),
		NewString("Alice"), NewInt(30), &Float{Value: 50000.50}, TRUE)
	require.IsType(t, &Array{}, result)

	result = BuiltinDbExec.Fn(db, NewString("INSERT INTO users (name, age, salary, active) VALUES (?, ?, ?, ?)"),
		NewString("Bob"), NewInt(25), &Float{Value: 40000.00}, FALSE)
	require.IsType(t, &Array{}, result)

	// Test query all - dbQuery returns strings
	result = BuiltinDbQuery.Fn(db, NewString("SELECT * FROM users ORDER BY id"))
	rows, ok := result.(*Array)
	require.True(t, ok, "expected Array, got %T", result)
	assert.Equal(t, 2, len(rows.Elements))

	// Check first row
	row1, ok := rows.Elements[0].(*Map)
	require.True(t, ok)

	// Verify all values are strings (Charlang behavior)
	nameVal := row1.Pairs[NewString("name").HashKey()].Value
	name, ok := nameVal.(*String)
	require.True(t, ok, "name should be String, got %T", nameVal)
	assert.Equal(t, "Alice", name.Value)

	ageVal := row1.Pairs[NewString("age").HashKey()].Value
	age, ok := ageVal.(*String)
	require.True(t, ok, "age should be String in dbQuery, got %T", ageVal)
	assert.Equal(t, "30", age.Value)

	salaryVal := row1.Pairs[NewString("salary").HashKey()].Value
	salary, ok := salaryVal.(*String)
	require.True(t, ok, "salary should be String in dbQuery, got %T", salaryVal)
	assert.Contains(t, salary.Value, "50000")

	// Note: SQLite stores BOOLEAN as INTEGER (0 or 1)
	activeVal := row1.Pairs[NewString("active").HashKey()].Value
	active, ok := activeVal.(*String)
	require.True(t, ok, "active should be String in dbQuery, got %T", activeVal)
	assert.Equal(t, "1", active.Value)

	// Test query with parameter
	result = BuiltinDbQuery.Fn(db, NewString("SELECT name, age FROM users WHERE age > ?"), NewInt(26))
	rows, ok = result.(*Array)
	require.True(t, ok)
	assert.Equal(t, 1, len(rows.Elements)) // Only Alice
}

// TestBuiltinDbQueryRow tests the dbQueryRow function.
func TestBuiltinDbQueryRowTyped(t *testing.T) {
	tmpFile := "test_builtin_db_query_row.sqlite"
	defer os.Remove(tmpFile)

	// Connect and setup
	result := BuiltinDbConnect.Fn(NewString("sqlite"), NewString(tmpFile))
	db, ok := result.(*DB)
	require.True(t, ok)
	defer db.Close()

	// Create table and insert
	BuiltinDbExec.Fn(db, NewString("CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)"))
	BuiltinDbExec.Fn(db, NewString("INSERT INTO items (name) VALUES (?)"), NewString("item1"))

	// Test query single row
	result = BuiltinDbQueryRowTyped.Fn(db, NewString("SELECT * FROM items WHERE id = ?"), NewInt(1))
	row, ok := result.(*Map)
	require.True(t, ok)
	assert.NotNil(t, row.Pairs)

	// Test query non-existent row
	result = BuiltinDbQueryRowTyped.Fn(db, NewString("SELECT * FROM items WHERE id = ?"), NewInt(999))
	assert.Equal(t, NULL, result)
}

// TestBuiltinDbQueryArrayTyped tests the dbQueryArrayTyped function.
func TestBuiltinDbQueryArrayTyped(t *testing.T) {
	tmpFile := "test_builtin_db_query_array.sqlite"
	defer os.Remove(tmpFile)

	// Connect and setup
	result := BuiltinDbConnect.Fn(NewString("sqlite"), NewString(tmpFile))
	db, ok := result.(*DB)
	require.True(t, ok)
	defer db.Close()

	// Create table and insert
	BuiltinDbExec.Fn(db, NewString("CREATE TABLE numbers (id INTEGER PRIMARY KEY, value INTEGER)"))
	BuiltinDbExec.Fn(db, NewString("INSERT INTO numbers (value) VALUES (1), (2), (3)"))

	// Test query array
	result = BuiltinDbQueryArrayTyped.Fn(db, NewString("SELECT value FROM numbers ORDER BY id"))
	rows, ok := result.(*Array)
	require.True(t, ok, "expected Array, got %T", result)
	assert.Equal(t, 3, len(rows.Elements))

	// Check each row is an array
	for i, row := range rows.Elements {
		arr, ok := row.(*Array)
		require.True(t, ok, "row %d should be an Array, got %T", i, row)
		assert.Equal(t, 1, len(arr.Elements))

		// Verify type is preserved
		val, ok := arr.Elements[0].(*Int)
		require.True(t, ok, "value should be Int, got %T", arr.Elements[0])
		assert.Equal(t, int64(i+1), val.Value)
	}
}

// TestBuiltinDbQueryValueTyped tests the dbQueryValueTyped function.
func TestBuiltinDbQueryValueTyped(t *testing.T) {
	tmpFile := "test_builtin_db_query_value.sqlite"
	defer os.Remove(tmpFile)

	// Connect and setup
	result := BuiltinDbConnect.Fn(NewString("sqlite"), NewString(tmpFile))
	db, ok := result.(*DB)
	require.True(t, ok)
	defer db.Close()

	// Create table and insert
	BuiltinDbExec.Fn(db, NewString("CREATE TABLE metrics (id INTEGER PRIMARY KEY, count INTEGER, ratio REAL, label TEXT)"))
	BuiltinDbExec.Fn(db, NewString("INSERT INTO metrics (count, ratio, label) VALUES (42, 3.14, 'test')"))

	// Test query int
	result = BuiltinDbQueryValueTyped.Fn(db, NewString("SELECT count FROM metrics WHERE id = 1"))
	count, ok := result.(*Int)
	require.True(t, ok, "expected Int, got %T", result)
	assert.Equal(t, int64(42), count.Value)

	// Test query float
	result = BuiltinDbQueryValueTyped.Fn(db, NewString("SELECT ratio FROM metrics WHERE id = 1"))
	ratio, ok := result.(*Float)
	require.True(t, ok, "expected Float, got %T", result)
	assert.InDelta(t, 3.14, ratio.Value, 0.01)

	// Test query string
	result = BuiltinDbQueryValueTyped.Fn(db, NewString("SELECT label FROM metrics WHERE id = 1"))
	label, ok := result.(*String)
	require.True(t, ok, "expected String, got %T", result)
	assert.Equal(t, "test", label.Value)

	// Test query count
	result = BuiltinDbQueryValueTyped.Fn(db, NewString("SELECT COUNT(*) FROM metrics"))
	count, ok = result.(*Int)
	require.True(t, ok)
	assert.Equal(t, int64(1), count.Value)

	// Test query non-existent
	result = BuiltinDbQueryValueTyped.Fn(db, NewString("SELECT count FROM metrics WHERE id = 999"))
	assert.Equal(t, NULL, result)
}

// TestBuiltinDbExec tests the dbExec function.
func TestBuiltinDbExec(t *testing.T) {
	tmpFile := "test_builtin_db_exec.sqlite"
	defer os.Remove(tmpFile)

	// Connect
	result := BuiltinDbConnect.Fn(NewString("sqlite"), NewString(tmpFile))
	db, ok := result.(*DB)
	require.True(t, ok)
	defer db.Close()

	// Test CREATE
	result = BuiltinDbExec.Fn(db, NewString("CREATE TABLE test_exec (id INTEGER PRIMARY KEY, value TEXT)"))
	arr, ok := result.(*Array)
	require.True(t, ok)
	// For CREATE, lastInsertId is 0
	assert.Equal(t, int64(0), arr.Elements[0].(*Int).Value)

	// Test INSERT
	result = BuiltinDbExec.Fn(db, NewString("INSERT INTO test_exec (value) VALUES (?)"), NewString("hello"))
	arr, ok = result.(*Array)
	require.True(t, ok)
	assert.True(t, arr.Elements[0].(*Int).Value > 0) // lastInsertId should be > 0
	assert.Equal(t, int64(1), arr.Elements[1].(*Int).Value) // rowsAffected = 1

	// Test UPDATE
	result = BuiltinDbExec.Fn(db, NewString("UPDATE test_exec SET value = ? WHERE id = ?"),
		NewString("world"), NewInt(1))
	arr, ok = result.(*Array)
	require.True(t, ok)
	assert.Equal(t, int64(1), arr.Elements[1].(*Int).Value) // rowsAffected = 1

	// Test DELETE
	BuiltinDbExec.Fn(db, NewString("INSERT INTO test_exec (value) VALUES (?)"), NewString("delete_me"))
	result = BuiltinDbExec.Fn(db, NewString("DELETE FROM test_exec WHERE value = ?"), NewString("delete_me"))
	arr, ok = result.(*Array)
	require.True(t, ok)
	assert.Equal(t, int64(1), arr.Elements[1].(*Int).Value) // rowsAffected = 1
}

// TestDbTypePreservation tests that native types are preserved.
func TestDbTypePreservation(t *testing.T) {
	tmpFile := "test_builtin_db_types.sqlite"
	defer os.Remove(tmpFile)

	// Connect
	result := BuiltinDbConnect.Fn(NewString("sqlite"), NewString(tmpFile))
	db, ok := result.(*DB)
	require.True(t, ok)
	defer db.Close()

	// Create table with various types
	BuiltinDbExec.Fn(db, NewString(`
		CREATE TABLE types_test (
			id INTEGER PRIMARY KEY,
			int_val INTEGER,
			float_val REAL,
			text_val TEXT,
			bool_val INTEGER
		)
	`))

	// Insert with various types
	BuiltinDbExec.Fn(db, NewString(`
		INSERT INTO types_test (int_val, float_val, text_val, bool_val)
		VALUES (?, ?, ?, ?)
	`), NewInt(123), &Float{Value: 45.67}, NewString("hello"), TRUE)

	// Query and verify types - use Typed version
	result = BuiltinDbQueryTyped.Fn(db, NewString("SELECT * FROM types_test"))
	rows, ok := result.(*Array)
	require.True(t, ok)
	require.Equal(t, 1, len(rows.Elements))

	row := rows.Elements[0].(*Map)

	// Check int type
	intVal := row.Pairs[NewString("int_val").HashKey()].Value
	_, isInt := intVal.(*Int)
	assert.True(t, isInt, "int_val should be Int, got %T", intVal)

	// Check float type
	floatVal := row.Pairs[NewString("float_val").HashKey()].Value
	_, isFloat := floatVal.(*Float)
	assert.True(t, isFloat, "float_val should be Float, got %T", floatVal)

	// Check string type
	textVal := row.Pairs[NewString("text_val").HashKey()].Value
	_, isString := textVal.(*String)
	assert.True(t, isString, "text_val should be String, got %T", textVal)

	// Check bool type (SQLite stores as 0/1)
	boolVal := row.Pairs[NewString("bool_val").HashKey()].Value
	_, isInt = boolVal.(*Int)
	assert.True(t, isInt, "bool_val should be Int in SQLite, got %T", boolVal)
}

// TestDbNullHandling tests that NULL values are handled correctly.
func TestDbNullHandling(t *testing.T) {
	tmpFile := "test_builtin_db_null.sqlite"
	defer os.Remove(tmpFile)

	// Connect
	result := BuiltinDbConnect.Fn(NewString("sqlite"), NewString(tmpFile))
	db, ok := result.(*DB)
	require.True(t, ok)
	defer db.Close()

	// Create table with nullable columns
	BuiltinDbExec.Fn(db, NewString("CREATE TABLE null_test (id INTEGER PRIMARY KEY, value TEXT)"))
	BuiltinDbExec.Fn(db, NewString("INSERT INTO null_test (value) VALUES (NULL)"))

	// Query null value - use Typed version for NULL handling
	result = BuiltinDbQueryTyped.Fn(db, NewString("SELECT * FROM null_test"))
	rows, ok := result.(*Array)
	require.True(t, ok)
	require.Equal(t, 1, len(rows.Elements))

	row := rows.Elements[0].(*Map)
	value := row.Pairs[NewString("value").HashKey()].Value
	assert.Equal(t, NULL, value, "NULL database value should be NULL object")
}

// TestDbQueryStringVersion tests that string-based functions convert all values to strings.
func TestDbQueryStringVersion(t *testing.T) {
	tmpFile := "test_builtin_db_string_version.sqlite"
	defer os.Remove(tmpFile)

	// Connect
	result := BuiltinDbConnect.Fn(NewString("sqlite"), NewString(tmpFile))
	db, ok := result.(*DB)
	require.True(t, ok)
	defer db.Close()

	// Create table and insert
	BuiltinDbExec.Fn(db, NewString("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, age INTEGER, salary REAL)"))
	BuiltinDbExec.Fn(db, NewString("INSERT INTO test (name, age, salary) VALUES (?, ?, ?)"), NewString("Alice"), NewInt(30), &Float{Value: 50000.50})

	// Test dbQuery returns strings
	rows := BuiltinDbQuery.Fn(db, NewString("SELECT * FROM test"))
	arr, ok := rows.(*Array)
	require.True(t, ok)
	require.Equal(t, 1, len(arr.Elements))

	row := arr.Elements[0].(*Map)
	name := row.Pairs[NewString("name").HashKey()].Value
	_, isString := name.(*String)
	assert.True(t, isString, "name should be String, got %T", name)

	age := row.Pairs[NewString("age").HashKey()].Value
	_, isString = age.(*String)
	assert.True(t, isString, "age should be String in string version, got %T", age)

	salary := row.Pairs[NewString("salary").HashKey()].Value
	_, isString = salary.(*String)
	assert.True(t, isString, "salary should be String in string version, got %T", salary)

	// Test dbQueryRecs returns 2D array with header
	recs := BuiltinDbQueryRecs.Fn(db, NewString("SELECT name, age FROM test"))
	recsArr, ok := recs.(*Array)
	require.True(t, ok)
	require.Equal(t, 2, len(recsArr.Elements)) // header + 1 data row

	// First row is header
	header := recsArr.Elements[0].(*Array)
	assert.Equal(t, "name", header.Elements[0].(*String).Value)
	assert.Equal(t, "age", header.Elements[1].(*String).Value)

	// Test dbQueryString
	nameStr := BuiltinDbQueryString.Fn(db, NewString("SELECT name FROM test WHERE id = 1"))
	_, isString = nameStr.(*String)
	assert.True(t, isString)

	// Test dbQueryCount
	count := BuiltinDbQueryCount.Fn(db, NewString("SELECT COUNT(*) FROM test"))
	_, isInt := count.(*Int)
	assert.True(t, isInt)

	// Test dbQueryFloat
	sum := BuiltinDbQueryFloat.Fn(db, NewString("SELECT SUM(salary) FROM test"))
	_, isFloat := sum.(*Float)
	assert.True(t, isFloat)
}
