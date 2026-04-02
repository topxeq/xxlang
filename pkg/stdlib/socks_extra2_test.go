// pkg/stdlib/socks_extra2_test.go
// Additional tests for socks module to increase coverage of basic methods.
package stdlib

import (
	"testing"
)

// TestSocksServer_BasicMethods covers Type, TypeTag, Inspect, ToBool, HashKey.
func TestSocksServer_BasicMethods(t *testing.T) {
	s := NewSocksServer()
	// Type
	typ := s.Type()
	if typ == "" {
		t.Error("Type() returned empty")
	}
	// TypeTag
	_ = s.TypeTag()
	// Inspect
	insp := s.Inspect()
	if insp == "" {
		t.Error("Inspect() returned empty")
	}
	// ToBool: initially not running
	b := s.ToBool()
	if b == nil {
		t.Error("ToBool() returned nil")
	} else if b.Value != false {
		t.Errorf("ToBool() = %v, want false", b.Value)
	}
	// HashKey
	hk := s.HashKey()
	if hk.Type == "" {
		t.Error("HashKey() returned empty Type")
	}
}

// TestSocksClient_BasicMethods covers Type, TypeTag, Inspect, ToBool, HashKey.
func TestSocksClient_BasicMethods(t *testing.T) {
	c := NewSocksClient()
	// Type
	typ := c.Type()
	if typ == "" {
		t.Error("Type() returned empty")
	}
	// TypeTag
	_ = c.TypeTag()
	// Inspect
	insp := c.Inspect()
	if insp == "" {
		t.Error("Inspect() returned empty")
	}
	// ToBool: initially not connected
	b := c.ToBool()
	if b == nil {
		t.Error("ToBool() returned nil")
	} else if b.Value != false {
		t.Errorf("ToBool() = %v, want false", b.Value)
	}
	// HashKey
	hk := c.HashKey()
	if hk.Type == "" {
		t.Error("HashKey() returned empty Type")
	}
}
