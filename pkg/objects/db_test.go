// pkg/objects/db_test.go
// Tests for database (DB) object type
package objects

import (
	"database/sql"
	"strings"
	"sync"
	"testing"
)

func TestNewDB(t *testing.T) {
	// Create a mock DB (without actual connection)
	db := &DB{
		Value:          nil,
		DriverName:     "test-driver",
		DataSourceName: "test-dsn",
		mu:             sync.Mutex{},
		closed:         false,
	}

	if db.Type() != DBType {
		t.Errorf("expected type DB, got %s", db.Type())
	}

	if db.TypeTag() != TagDB {
		t.Errorf("expected TypeTag TagDB, got %d", db.TypeTag())
	}

	inspect := db.Inspect()
	if inspect == "" {
		t.Errorf("expected non-empty Inspect")
	}
	if !strings.Contains(inspect, "[db") {
		t.Errorf("expected Inspect to contain '[db', got %s", inspect)
	}

	// Test ToBool with nil Value
	if db.ToBool().Value {
		t.Errorf("expected ToBool() false for nil Value")
	}

	// Test with non-nil Value
	db.Value = &sql.DB{}
	if !db.ToBool().Value {
		t.Errorf("expected ToBool() true for connected DB")
	}

	key := db.HashKey()
	if key.Type != DBType {
		t.Errorf("expected HashKey.Type DB, got %s", key.Type)
	}
}

func TestDB_IsClosed(t *testing.T) {
	db := &DB{
		Value:  &sql.DB{},
		closed: false,
		mu:     sync.Mutex{},
	}

	if db.IsClosed() {
		t.Errorf("expected IsClosed() false")
	}

	db.closed = true
	if !db.IsClosed() {
		t.Errorf("expected IsClosed() true")
	}
}

func TestDB_Close(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *DB
		expectError  bool
		expectClosed bool
	}{
		{
			name: "nil value",
			setup: func() *DB {
				return &DB{Value: nil, closed: false}
			},
			expectError:  false,
			expectClosed: false, // Close shouldn't change closed flag if Value is nil
		},
		{
			name: "already closed",
			setup: func() *DB {
				return &DB{Value: &sql.DB{}, closed: true}
			},
			expectError:  false,
			expectClosed: true,
		},
		{
			name: "close with nil value",
			setup: func() *DB {
				return &DB{Value: nil, closed: false}
			},
			expectError:  false,
			expectClosed: false, // Close with nil Value does not change closed flag
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup()
			err := db.Close()

			if (err != nil) != tt.expectError {
				t.Errorf("Close() error = %v, wantError %v", err, tt.expectError)
			}

			if db.closed != tt.expectClosed {
				t.Errorf("closed = %v, want %v", db.closed, tt.expectClosed)
			}
		})
	}
}

func TestDB_GetMember(t *testing.T) {
	db := &DB{
		Value:          &sql.DB{},
		DriverName:     "mysql",
		DataSourceName: "user:pass@tcp(localhost:3306)/test",
		mu:             sync.Mutex{},
		closed:         false,
	}

	tests := []struct {
		name     string
		member   string
		wantType ObjectType
		wantObj  interface{}
	}{
		{
			name:     "driver",
			member:   "driver",
			wantType: StringType,
			wantObj:  "mysql",
		},
		{
			name:     "closed",
			member:   "closed",
			wantType: BoolType,
			wantObj:  false,
		},
		{
			name:     "stats",
			member:   "stats",
			wantType: MapType,
			wantObj:  nil, // Just check type
		},
		{
			name:     "nonexistent member",
			member:   "nonexistent",
			wantType: NullType,
			wantObj:  NULL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.GetMember(tt.member)

			if result.Type() != tt.wantType {
				t.Errorf("GetMember(%s) type = %s, want %s", tt.member, result.Type(), tt.wantType)
			}

			if tt.wantObj != nil {
				switch expected := tt.wantObj.(type) {
				case string:
					if str, ok := result.(*String); !ok || str.Value != expected {
						t.Errorf("GetMember(%s) = %v, want %s", tt.member, result, expected)
					}
				case bool:
					if b, ok := result.(*Bool); !ok || b.Value != expected {
						t.Errorf("GetMember(%s) = %v, want %t", tt.member, result, expected)
					}
				}
			}
		})
	}

	// Test closed connection
	db.closed = true
	result := db.GetMember("driver")
	if _, ok := result.(*Error); !ok {
		t.Errorf("expected error for closed connection, got %v", result)
	}

	db.Value = nil
	result = db.GetMember("driver")
	if _, ok := result.(*Error); !ok {
		t.Errorf("expected error for nil Value, got %v", result)
	}
}

func TestDB_getStats(t *testing.T) {
	db := &DB{
		Value: &sql.DB{},
	}

	stats := db.getStats()

	mapObj, ok := stats.(*Map)
	if !ok {
		t.Fatalf("expected Map, got %T", stats)
	}

	// Check for expected keys based on actual implementation
	expectedKeys := []string{
		"maxOpenConnections",
		"openConnections",
		"inUse",
		"idle",
		"waitCount",
		"waitDuration",
	}

	for _, key := range expectedKeys {
		k := NewString(key).HashKey()
		if _, exists := mapObj.Pairs[k]; !exists {
			t.Errorf("expected key '%s' in stats map", key)
		}
	}
}
