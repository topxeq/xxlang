// pkg/server/xhp_test.go
package server

import (
	"testing"
)

func TestProcessXHP_SimpleExpression(t *testing.T) {
	content := `<p>Result: <?xhp return "1" + "2" ?></p>`
	expected := `<p>Result: 12</p>`

	result, err := ProcessXHP(content, "test.xhp", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("ProcessXHP failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessXHP_MultipleBlocks(t *testing.T) {
	content := `<p><?xhp return "a" ?><?xhp return "b" ?></p>`
	expected := `<p>ab</p>`

	result, err := ProcessXHP(content, "test.xhp", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("ProcessXHP failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessXHP_NoBlocks(t *testing.T) {
	content := `<p>Plain HTML</p>`
	expected := `<p>Plain HTML</p>`

	result, err := ProcessXHP(content, "test.xhp", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("ProcessXHP failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessXHP_EmptyBlock(t *testing.T) {
	content := `<p><?xhp ?></p>`
	expected := `<p></p>`

	result, err := ProcessXHP(content, "test.xhp", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("ProcessXHP failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessXHP_UnclosedBlock(t *testing.T) {
	content := `<p><?xhp return "test"</p>`

	_, err := ProcessXHP(content, "test.xhp", nil, nil, nil, nil)
	if err == nil {
		t.Error("Expected error for unclosed block")
	}
}

func TestProcessXHP_ParameterAccess(t *testing.T) {
	content := `<p>Hello, <?xhp return paraMapG["name"] ?>!</p>`
	expected := `<p>Hello, World!</p>`

	paraMap := map[string]string{"name": "World"}
	result, err := ProcessXHP(content, "test.xhp", nil, nil, paraMap, nil)
	if err != nil {
		t.Fatalf("ProcessXHP failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessXHP_ComplexCode(t *testing.T) {
	content := `<?xhp
		var a = 10
		var b = 20
		return toStr(a + b)
	?>`
	expected := `30`

	result, err := ProcessXHP(content, "test.xhp", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("ProcessXHP failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessXHP_SharedContext(t *testing.T) {
	// Test that variables defined in one block are available in another
	// Variables are shared because all code blocks run in the same scope
	content := `<?xhp var greeting = "Hello" ?><?xhp echo(greeting) ?>, World!`
	expected := `Hello, World!`

	result, err := ProcessXHP(content, "test.xhp", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("ProcessXHP failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessXHP_SharedContextCounter(t *testing.T) {
	// Test that variables can be modified across blocks
	content := `<?xhp var counter = 0 ?><?xhp counter = counter + 1 ?><?xhp echo(counter) ?>`
	expected := `1`

	result, err := ProcessXHP(content, "test.xhp", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("ProcessXHP failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessXHP_ComplexSharedContext(t *testing.T) {
	// Test complex data sharing between blocks
	content := `Start:<?xhp var data = {"name": "Alice", "age": 30} ?>Name: <?xhp echo(data.name) ?>, Age: <?xhp echo(data.age) ?>:End`
	expected := `Start:Name: Alice, Age: 30:End`

	result, err := ProcessXHP(content, "test.xhp", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("ProcessXHP failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessXHP_EchoFunction(t *testing.T) {
	// Test echo function for output
	content := `<?xhp echo("Hello") ?> <?xhp echo("World") ?>`
	expected := `Hello World`

	result, err := ProcessXHP(content, "test.xhp", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("ProcessXHP failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessXHP_ReturnAndEcho(t *testing.T) {
	// Test mixing return and echo
	content := `<?xhp echo("A") ?><?xhp return "B" ?><?xhp echo("C") ?>`
	expected := `ABC`

	result, err := ProcessXHP(content, "test.xhp", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("ProcessXHP failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
