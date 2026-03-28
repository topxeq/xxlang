// pkg/objects/builtin_db.go
// Database builtin functions for Xxlang.
// Provides two versions of database functions:
// 1. String-based functions (Charlang compatible): dbQuery, dbQueryRecs, dbQueryMap, etc.
// 2. Typed functions (preserve native types): dbQueryTyped, dbQueryRowTyped, etc.
package objects

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ============================================================
// Database Builtin Functions
// ============================================================
//
// Two versions are provided:
//
// **String-based (Charlang compatible)**:
//   - formatSQLValue, dbConnect, dbClose, dbQuery, dbQueryOrdered, dbQueryRecs,
//     dbQueryMap, dbQueryMapArray, dbQueryCount, dbQueryFloat, dbQueryString, dbExec
//   - All values converted to strings
//   - NULL values become empty strings
//
// **Typed (preserve native types)**:
//   - dbQueryTyped, dbQueryRowTyped, dbQueryArrayTyped, dbQueryValueTyped
//   - Native data types preserved (int, float, bool, string, time)
//   - NULL values become null (undefined)

// ============================================================
// Helper Functions
// ============================================================

// dbValueToString converts a database value to a string.
// NULL values become empty strings (Charlang behavior).
func dbValueToString(val interface{}) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int16:
		return fmt.Sprintf("%d", v)
	case int8:
		return fmt.Sprintf("%d", v)
	case uint:
		return fmt.Sprintf("%d", v)
	case uint64:
		return fmt.Sprintf("%d", v)
	case uint32:
		return fmt.Sprintf("%d", v)
	case uint16:
		return fmt.Sprintf("%d", v)
	case uint8:
		return fmt.Sprintf("%d", v)
	case float32:
		// Clean up float representation
		s := fmt.Sprintf("%v", v)
		return cleanFloatString(s)
	case float64:
		s := fmt.Sprintf("%v", v)
		return cleanFloatString(s)
	case string:
		return v
	case []byte:
		return string(v)
	case bool:
		if v {
			return "1"
		}
		return "0"
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	case sql.NullString:
		if !v.Valid {
			return ""
		}
		return v.String
	case sql.NullInt64:
		if !v.Valid {
			return ""
		}
		return fmt.Sprintf("%d", v.Int64)
	case sql.NullFloat64:
		if !v.Valid {
			return ""
		}
		return fmt.Sprintf("%v", v.Float64)
	case sql.NullBool:
		if !v.Valid {
			return ""
		}
		if v.Bool {
			return "1"
		}
		return "0"
	case sql.NullTime:
		if !v.Valid {
			return ""
		}
		return v.Time.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// cleanFloatString removes trailing zeros from float string representation.
func cleanFloatString(s string) string {
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		if strings.HasSuffix(s, ".") {
			s = strings.TrimSuffix(s, ".")
		}
	}
	return s
}

// dbValueToObjectPreserveType converts a database value to an Xxlang Object,
// preserving the native data type. NULL values become undefined.
func dbValueToObjectPreserveType(val interface{}) Object {
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
		if v > uint64(1<<63-1) {
			return NewString(fmt.Sprintf("%d", v))
		}
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
	case sql.NullString:
		if !v.Valid {
			return NULL
		}
		return NewString(v.String)
	case sql.NullInt64:
		if !v.Valid {
			return NULL
		}
		return NewInt(v.Int64)
	case sql.NullFloat64:
		if !v.Valid {
			return NULL
		}
		return &Float{Value: v.Float64}
	case sql.NullBool:
		if !v.Valid {
			return NULL
		}
		return &Bool{Value: v.Bool}
	case sql.NullTime:
		if !v.Valid {
			return NULL
		}
		return NewString(v.Time.Format(time.RFC3339))
	default:
		return NewString(fmt.Sprintf("%v", v))
	}
}

// objectToGoValue converts an Xxlang Object to a Go value for database parameters.
func objectToGoValue(obj Object) interface{} {
	if obj == nil || obj == NULL {
		return nil
	}

	switch v := obj.(type) {
	case *Int:
		return v.Value
	case *Float:
		return v.Value
	case *String:
		return v.Value
	case *Bool:
		return v.Value
	case *Bytes:
		return v.Value
	default:
		return obj.Inspect()
	}
}

// ============================================================
// Shared Functions (formatSQLValue, dbConnect, dbClose, dbExec)
// ============================================================

// BuiltinFormatSQLValue escapes special characters in a string for safe SQL usage.
var BuiltinFormatSQLValue = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for formatSQLValue. got=%d, want=1", len(args))
		}

		str, ok := args[0].(*String)
		if !ok {
			return newError("argument to 'formatSQLValue' must be STRING, got %s", args[0].Type())
		}

		result := strings.Replace(str.Value, "\r", "\\r", -1)
		result = strings.Replace(result, "\n", "\\n", -1)
		result = strings.Replace(result, "'", "''", -1)

		return NewString(result)
	},
}

// BuiltinDbConnect connects to a database and returns a DB object.
var BuiltinDbConnect = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for dbConnect. got=%d, want=2", len(args))
		}

		driver, ok := args[0].(*String)
		if !ok {
			return newError("first argument to 'dbConnect' must be STRING (driver), got %s", args[0].Type())
		}

		connStr, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbConnect' must be STRING (connection string), got %s", args[1].Type())
		}

		driverName := driver.Value

		// Normalize driver names
		switch driverName {
		case "sqlite3":
			driverName = "sqlite"
		case "pg", "postgresql":
			driverName = "postgres"
		case "sqlserver", "mssqlserver":
			driverName = "mssql"
		}

		// Open database connection
		db, err := sql.Open(driverName, connStr.Value)
		if err != nil {
			return newError("failed to open database: %v", err)
		}

		// Verify connection with ping
		if err := db.Ping(); err != nil {
			db.Close()
			return newError("failed to connect to database: %v", err)
		}

		return NewDB(db, driverName, connStr.Value)
	},
}

// BuiltinDbClose closes a database connection.
var BuiltinDbClose = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for dbClose. got=%d, want=1", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("argument to 'dbClose' must be DB, got %s", args[0].Type())
		}

		if err := db.Close(); err != nil {
			return newError("failed to close database: %v", err)
		}

		return NULL
	},
}

// BuiltinDbExec executes a SQL statement (INSERT, UPDATE, DELETE, etc.)
// Returns: [lastInsertId, rowsAffected]
var BuiltinDbExec = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for dbExec. got=%d, want>=2", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("first argument to 'dbExec' must be DB, got %s", args[0].Type())
		}

		sqlStmt, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbExec' must be STRING, got %s", args[1].Type())
		}

		execArgs := make([]interface{}, len(args)-2)
		for i, arg := range args[2:] {
			execArgs[i] = objectToGoValue(arg)
		}

		result, err := db.Value.Exec(sqlStmt.Value, execArgs...)
		if err != nil {
			return newError("exec failed: %v", err)
		}

		lastID, _ := result.LastInsertId()
		rowsAffected, _ := result.RowsAffected()

		return NewArray([]Object{
			NewInt(lastID),
			NewInt(rowsAffected),
		})
	},
}

// ============================================================
// String-based Query Functions (Charlang Compatible)
// ============================================================

// BuiltinDbQuery executes a SQL query and returns results as an array of maps.
// All values are converted to strings (Charlang behavior).
var BuiltinDbQuery = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for dbQuery. got=%d, want>=2", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("first argument to 'dbQuery' must be DB, got %s", args[0].Type())
		}

		query, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbQuery' must be STRING, got %s", args[1].Type())
		}

		queryArgs := make([]interface{}, len(args)-2)
		for i, arg := range args[2:] {
			queryArgs[i] = objectToGoValue(arg)
		}

		rows, err := db.Value.Query(query.Value, queryArgs...)
		if err != nil {
			return newError("query failed: %v", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return newError("failed to get columns: %v", err)
		}

		var result []Object

		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return newError("scan failed: %v", err)
			}

			pairs := make(map[HashKey]MapPair)
			for i, col := range columns {
				key := NewString(col)
				pairs[key.HashKey()] = MapPair{
					Key:   key,
					Value: NewString(dbValueToString(values[i])),
				}
			}
			result = append(result, NewMap(pairs))
		}

		if err := rows.Err(); err != nil {
			return newError("rows error: %v", err)
		}

		return NewArray(result)
	},
}

// BuiltinDbQueryOrdered executes a SQL query and returns results as an array of OrderedMaps.
// Column order is preserved in each OrderedMap. All values are converted to strings.
var BuiltinDbQueryOrdered = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for dbQueryOrdered. got=%d, want>=2", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("first argument to 'dbQueryOrdered' must be DB, got %s", args[0].Type())
		}

		query, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbQueryOrdered' must be STRING, got %s", args[1].Type())
		}

		queryArgs := make([]interface{}, len(args)-2)
		for i, arg := range args[2:] {
			queryArgs[i] = objectToGoValue(arg)
		}

		rows, err := db.Value.Query(query.Value, queryArgs...)
		if err != nil {
			return newError("query failed: %v", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return newError("failed to get columns: %v", err)
		}

		var result []Object

		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return newError("scan failed: %v", err)
			}

			// Create ordered map with pairs in column order
			om := NewOrderedMap()
			for i, col := range columns {
				om.Set(NewString(col), NewString(dbValueToString(values[i])))
			}
			result = append(result, om)
		}

		if err := rows.Err(); err != nil {
			return newError("rows error: %v", err)
		}

		return NewArray(result)
	},
}

// BuiltinDbQueryRecs executes a SQL query and returns results as a 2D string array.
// The first row contains column names (Charlang behavior).
var BuiltinDbQueryRecs = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for dbQueryRecs. got=%d, want>=2", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("first argument to 'dbQueryRecs' must be DB, got %s", args[0].Type())
		}

		query, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbQueryRecs' must be STRING, got %s", args[1].Type())
		}

		queryArgs := make([]interface{}, len(args)-2)
		for i, arg := range args[2:] {
			queryArgs[i] = objectToGoValue(arg)
		}

		rows, err := db.Value.Query(query.Value, queryArgs...)
		if err != nil {
			return newError("query failed: %v", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return newError("failed to get columns: %v", err)
		}

		var result []Object

		// First row is column names
		headerRow := make([]Object, len(columns))
		for i, col := range columns {
			headerRow[i] = NewString(col)
		}
		result = append(result, NewArray(headerRow))

		// Data rows
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return newError("scan failed: %v", err)
			}

			rowArray := make([]Object, len(columns))
			for i, val := range values {
				rowArray[i] = NewString(dbValueToString(val))
			}
			result = append(result, NewArray(rowArray))
		}

		if err := rows.Err(); err != nil {
			return newError("rows error: %v", err)
		}

		return NewArray(result)
	},
}

// BuiltinDbQueryMap executes a SQL query and returns results as a map grouped by a key column.
// All values are converted to strings.
// Usage: dbQueryMap(db, sql, keyColumn, args...)
var BuiltinDbQueryMap = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 3 {
			return newError("wrong number of arguments for dbQueryMap. got=%d, want>=3", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("first argument to 'dbQueryMap' must be DB, got %s", args[0].Type())
		}

		query, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbQueryMap' must be STRING, got %s", args[1].Type())
		}

		keyCol, ok := args[2].(*String)
		if !ok {
			return newError("third argument to 'dbQueryMap' must be STRING (key column), got %s", args[2].Type())
		}

		queryArgs := make([]interface{}, len(args)-3)
		for i, arg := range args[3:] {
			queryArgs[i] = objectToGoValue(arg)
		}

		rows, err := db.Value.Query(query.Value, queryArgs...)
		if err != nil {
			return newError("query failed: %v", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return newError("failed to get columns: %v", err)
		}

		// Find key column index
		keyColIdx := -1
		for i, col := range columns {
			if col == keyCol.Value {
				keyColIdx = i
				break
			}
		}
		if keyColIdx < 0 {
			return newError("key column '%s' not found in result", keyCol.Value)
		}

		result := make(map[HashKey]MapPair)

		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return newError("scan failed: %v", err)
			}

			// Build row map
			pairs := make(map[HashKey]MapPair)
			for i, col := range columns {
				key := NewString(col)
				pairs[key.HashKey()] = MapPair{
					Key:   key,
					Value: NewString(dbValueToString(values[i])),
				}
			}

			// Use key column value as map key
			keyValue := NewString(dbValueToString(values[keyColIdx]))
			result[keyValue.HashKey()] = MapPair{
				Key:   keyValue,
				Value: NewMap(pairs),
			}
		}

		if err := rows.Err(); err != nil {
			return newError("rows error: %v", err)
		}

		return NewMap(result)
	},
}

// BuiltinDbQueryMapArray executes a SQL query and returns results as a map of arrays grouped by a key column.
// All values are converted to strings.
// Usage: dbQueryMapArray(db, sql, keyColumn, args...)
var BuiltinDbQueryMapArray = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 3 {
			return newError("wrong number of arguments for dbQueryMapArray. got=%d, want>=3", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("first argument to 'dbQueryMapArray' must be DB, got %s", args[0].Type())
		}

		query, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbQueryMapArray' must be STRING, got %s", args[1].Type())
		}

		keyCol, ok := args[2].(*String)
		if !ok {
			return newError("third argument to 'dbQueryMapArray' must be STRING (key column), got %s", args[2].Type())
		}

		queryArgs := make([]interface{}, len(args)-3)
		for i, arg := range args[3:] {
			queryArgs[i] = objectToGoValue(arg)
		}

		rows, err := db.Value.Query(query.Value, queryArgs...)
		if err != nil {
			return newError("query failed: %v", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return newError("failed to get columns: %v", err)
		}

		// Find key column index
		keyColIdx := -1
		for i, col := range columns {
			if col == keyCol.Value {
				keyColIdx = i
				break
			}
		}
		if keyColIdx < 0 {
			return newError("key column '%s' not found in result", keyCol.Value)
		}

		// Use a map to collect arrays for each key
		tempResult := make(map[string][]Object)

		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return newError("scan failed: %v", err)
			}

			// Build row map
			pairs := make(map[HashKey]MapPair)
			for i, col := range columns {
				key := NewString(col)
				pairs[key.HashKey()] = MapPair{
					Key:   key,
					Value: NewString(dbValueToString(values[i])),
				}
			}

			// Get key value
			keyValue := dbValueToString(values[keyColIdx])
			tempResult[keyValue] = append(tempResult[keyValue], NewMap(pairs))
		}

		if err := rows.Err(); err != nil {
			return newError("rows error: %v", err)
		}

		// Convert to result map
		result := make(map[HashKey]MapPair)
		for keyStr, arr := range tempResult {
			keyObj := NewString(keyStr)
			result[keyObj.HashKey()] = MapPair{
				Key:   keyObj,
				Value: NewArray(arr),
			}
		}

		return NewMap(result)
	},
}

// BuiltinDbQueryCount executes a SQL query that returns a single integer value.
// Returns -1 on error.
var BuiltinDbQueryCount = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for dbQueryCount. got=%d, want>=2", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("first argument to 'dbQueryCount' must be DB, got %s", args[0].Type())
		}

		query, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbQueryCount' must be STRING, got %s", args[1].Type())
		}

		queryArgs := make([]interface{}, len(args)-2)
		for i, arg := range args[2:] {
			queryArgs[i] = objectToGoValue(arg)
		}

		rows, err := db.Value.Query(query.Value, queryArgs...)
		if err != nil {
			return NewInt(-1)
		}
		defer rows.Close()

		if !rows.Next() {
			return NewInt(-1)
		}

		var count int = -1
		if err := rows.Scan(&count); err != nil {
			return NewInt(-1)
		}

		return NewInt(int64(count))
	},
}

// BuiltinDbQueryFloat executes a SQL query that returns a single float value.
var BuiltinDbQueryFloat = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for dbQueryFloat. got=%d, want>=2", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("first argument to 'dbQueryFloat' must be DB, got %s", args[0].Type())
		}

		query, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbQueryFloat' must be STRING, got %s", args[1].Type())
		}

		queryArgs := make([]interface{}, len(args)-2)
		for i, arg := range args[2:] {
			queryArgs[i] = objectToGoValue(arg)
		}

		rows, err := db.Value.Query(query.Value, queryArgs...)
		if err != nil {
			return newError("query failed: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			return &Float{Value: 0}
		}

		var val float64
		if err := rows.Scan(&val); err != nil {
			return newError("scan failed: %v", err)
		}

		return &Float{Value: val}
	},
}

// BuiltinDbQueryString executes a SQL query that returns a single string value.
var BuiltinDbQueryString = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for dbQueryString. got=%d, want>=2", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("first argument to 'dbQueryString' must be DB, got %s", args[0].Type())
		}

		query, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbQueryString' must be STRING, got %s", args[1].Type())
		}

		queryArgs := make([]interface{}, len(args)-2)
		for i, arg := range args[2:] {
			queryArgs[i] = objectToGoValue(arg)
		}

		rows, err := db.Value.Query(query.Value, queryArgs...)
		if err != nil {
			return newError("query failed: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			return NewString("")
		}

		var val string
		if err := rows.Scan(&val); err != nil {
			return newError("scan failed: %v", err)
		}

		return NewString(val)
	},
}

// ============================================================
// Typed Query Functions (Preserve Native Data Types)
// ============================================================

// BuiltinDbQueryTyped executes a SQL query and returns results as an array of maps.
// Native data types are preserved (int, float, bool, string, time).
var BuiltinDbQueryTyped = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for dbQueryTyped. got=%d, want>=2", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("first argument to 'dbQueryTyped' must be DB, got %s", args[0].Type())
		}

		query, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbQueryTyped' must be STRING, got %s", args[1].Type())
		}

		queryArgs := make([]interface{}, len(args)-2)
		for i, arg := range args[2:] {
			queryArgs[i] = objectToGoValue(arg)
		}

		rows, err := db.Value.Query(query.Value, queryArgs...)
		if err != nil {
			return newError("query failed: %v", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return newError("failed to get columns: %v", err)
		}

		var result []Object

		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return newError("scan failed: %v", err)
			}

			pairs := make(map[HashKey]MapPair)
			for i, col := range columns {
				key := NewString(col)
				pairs[key.HashKey()] = MapPair{
					Key:   key,
					Value: dbValueToObjectPreserveType(values[i]),
				}
			}
			result = append(result, NewMap(pairs))
		}

		if err := rows.Err(); err != nil {
			return newError("rows error: %v", err)
		}

		return NewArray(result)
	},
}

// BuiltinDbQueryRowTyped executes a SQL query and returns a single row as a map.
// Native data types are preserved.
var BuiltinDbQueryRowTyped = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for dbQueryRowTyped. got=%d, want>=2", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("first argument to 'dbQueryRowTyped' must be DB, got %s", args[0].Type())
		}

		query, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbQueryRowTyped' must be STRING, got %s", args[1].Type())
		}

		queryArgs := make([]interface{}, len(args)-2)
		for i, arg := range args[2:] {
			queryArgs[i] = objectToGoValue(arg)
		}

		rows, err := db.Value.Query(query.Value, queryArgs...)
		if err != nil {
			return newError("query failed: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			return NULL
		}

		columns, err := rows.Columns()
		if err != nil {
			return newError("failed to get columns: %v", err)
		}

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return newError("scan failed: %v", err)
		}

		pairs := make(map[HashKey]MapPair)
		for i, col := range columns {
			key := NewString(col)
			pairs[key.HashKey()] = MapPair{
				Key:   key,
				Value: dbValueToObjectPreserveType(values[i]),
			}
		}

		return NewMap(pairs)
	},
}

// BuiltinDbQueryArrayTyped executes a SQL query and returns results as a 2D array.
// Native data types are preserved.
var BuiltinDbQueryArrayTyped = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for dbQueryArrayTyped. got=%d, want>=2", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("first argument to 'dbQueryArrayTyped' must be DB, got %s", args[0].Type())
		}

		query, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbQueryArrayTyped' must be STRING, got %s", args[1].Type())
		}

		queryArgs := make([]interface{}, len(args)-2)
		for i, arg := range args[2:] {
			queryArgs[i] = objectToGoValue(arg)
		}

		rows, err := db.Value.Query(query.Value, queryArgs...)
		if err != nil {
			return newError("query failed: %v", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return newError("failed to get columns: %v", err)
		}

		var result []Object

		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return newError("scan failed: %v", err)
			}

			rowArray := make([]Object, len(columns))
			for i, val := range values {
				rowArray[i] = dbValueToObjectPreserveType(val)
			}
			result = append(result, NewArray(rowArray))
		}

		if err := rows.Err(); err != nil {
			return newError("rows error: %v", err)
		}

		return NewArray(result)
	},
}

// BuiltinDbQueryValueTyped executes a SQL query that returns a single value.
// Native data types are preserved.
var BuiltinDbQueryValueTyped = &Builtin{
	Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for dbQueryValueTyped. got=%d, want>=2", len(args))
		}

		db, ok := args[0].(*DB)
		if !ok {
			return newError("first argument to 'dbQueryValueTyped' must be DB, got %s", args[0].Type())
		}

		query, ok := args[1].(*String)
		if !ok {
			return newError("second argument to 'dbQueryValueTyped' must be STRING, got %s", args[1].Type())
		}

		queryArgs := make([]interface{}, len(args)-2)
		for i, arg := range args[2:] {
			queryArgs[i] = objectToGoValue(arg)
		}

		rows, err := db.Value.Query(query.Value, queryArgs...)
		if err != nil {
			return newError("query failed: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			return NULL
		}

		var value interface{}
		if err := rows.Scan(&value); err != nil {
			return newError("scan failed: %v", err)
		}

		return dbValueToObjectPreserveType(value)
	},
}
