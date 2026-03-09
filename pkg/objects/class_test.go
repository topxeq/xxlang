// pkg/objects/class_test.go
package objects

import "testing"

func TestClassObject(t *testing.T) {
	class := &Class{
		Name:       "Person",
		SuperClass: nil,
		Methods:    make(map[string]Object),
		Fields:     make(map[string]Object),
	}

	if class.Type() != ClassType {
		t.Errorf("expected ClassType, got %s", class.Type())
	}
	if class.Inspect() != "class Person" {
		t.Errorf("expected 'class Person', got %s", class.Inspect())
	}
}

func TestInstanceObject(t *testing.T) {
	class := &Class{Name: "Person"}
	instance := &Instance{
		Class:  class,
		Fields: map[string]Object{"name": &String{Value: "Alice"}},
	}

	if instance.Type() != InstanceType {
		t.Errorf("expected InstanceType, got %s", instance.Type())
	}
	if instance.Inspect() != "Person instance" {
		t.Errorf("expected 'Person instance', got %s", instance.Inspect())
	}
}
