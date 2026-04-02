// pkg/objects/yaml_obj_test.go
package objects

import (
	"testing"
)

func TestParseYAML(t *testing.T) {
	tests := []string{
		"key: value",
		"list:\n  - item1\n  - item2",
		"nested:\n  inner: value",
	}

	for _, tc := range tests {
		result, err := ParseYAML(tc)
		if err != nil {
			t.Errorf("ParseYAML error for '%s': %v", tc, err)
			continue
		}
		if result == nil {
			t.Error("expected non-nil result")
		}
	}
}

func TestSerializeYAML(t *testing.T) {
	obj := NewMap(map[HashKey]MapPair{
		NewString("key").HashKey(): {Key: NewString("key"), Value: NewString("value")},
	})

	result := SerializeYAML(obj, 2)
	if result == "" {
		t.Error("expected non-empty YAML string")
	}
}

func TestYAMLRoundTrip(t *testing.T) {
	yaml := "name: test\nvalue: 123"
	obj, err := ParseYAML(yaml)
	if err != nil {
		t.Fatalf("ParseYAML error: %v", err)
	}

	serialized := SerializeYAML(obj, 2)
	if serialized == "" {
		t.Error("expected non-empty serialized YAML")
	}
}

func TestYAMLToJSON(t *testing.T) {
	yaml := "name: test\nvalue: 123"
	jsonStr, err := YAMLToJSON(yaml)
	if err != nil {
		t.Errorf("YAMLToJSON error: %v", err)
	}
	if jsonStr == "" {
		t.Error("expected non-empty JSON string")
	}
}

func TestJSONToYAML(t *testing.T) {
	jsonStr := `{"name": "test", "value": 123}`
	yamlStr, err := JSONToYAML(jsonStr, 2)
	if err != nil {
		t.Errorf("JSONToYAML error: %v", err)
	}
	if yamlStr == "" {
		t.Error("expected non-empty YAML string")
	}
}

func TestSplitYAMLDocuments(t *testing.T) {
	yaml := "---\nkey1: value1\n---\nkey2: value2"
	docs := SplitYAMLDocuments(yaml)
	if len(docs) != 2 {
		t.Errorf("expected 2 documents, got %d", len(docs))
	}
}

func TestJoinYAMLDocuments(t *testing.T) {
	docs := []Object{NewString("key1: value1"), NewString("key2: value2")}
	result := JoinYAMLDocuments(docs, 2)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestYAMLSet(t *testing.T) {
	elements := []Object{NewString("item1"), NewString("item2")}
	result := YAMLSet(elements)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestIsYAMLSet(t *testing.T) {
	elements := []Object{NewString("item1"), NewString("item2")}
	m := YAMLSet(elements)
	result := IsYAMLSet(m)
	if !result {
		t.Error("expected true for YAML set")
	}
}

func TestYAMLOMap(t *testing.T) {
	pairs := []MapPair{
		{Key: NewString("key1"), Value: NewString("value1")},
		{Key: NewString("key2"), Value: NewString("value2")},
	}
	result := YAMLOMap(pairs)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestParseYAMLPairs(t *testing.T) {
	yaml := "key1: value1\nkey2: value2"
	result, err := ParseYAMLPairs(yaml)
	if err != nil {
		t.Errorf("ParseYAMLPairs error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(result))
	}
}

func TestDetectYAMLValueType(t *testing.T) {
	result := DetectYAMLValueType("123")
	if result != "int" && result != "float" {
		t.Errorf("expected numeric type, got %s", result)
	}

	result = DetectYAMLValueType("true")
	if result != "bool" {
		t.Errorf("expected bool type, got %s", result)
	}

	result = DetectYAMLValueType("null")
	if result != "null" {
		t.Errorf("expected null type, got %s", result)
	}
}

func TestValidateYAML(t *testing.T) {
	valid := "name: test"
	err := ValidateYAML(valid)
	if err != nil {
		t.Errorf("ValidateYAML error: %v", err)
	}
}

func TestExtractYAMLAnchors(t *testing.T) {
	yaml := "defaults: &defaults\n  key: value\nanchor_test:\n  <<: *defaults"
	result := ExtractYAMLAnchors(yaml)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestNormalizeYAML(t *testing.T) {
	yaml := "name:  test\nvalue: 123"
	result, err := NormalizeYAML(yaml, 2)
	if err != nil {
		t.Errorf("NormalizeYAML error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestMergeYAMLMaps(t *testing.T) {
	m1 := NewMap(map[HashKey]MapPair{
		NewString("key1").HashKey(): {Key: NewString("key1"), Value: NewString("value1")},
	})
	m2 := NewMap(map[HashKey]MapPair{
		NewString("key2").HashKey(): {Key: NewString("key2"), Value: NewString("value2")},
	})
	result := MergeYAMLMaps(m1, m2)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestDeepMergeYAMLMaps(t *testing.T) {
	m1 := NewMap(map[HashKey]MapPair{
		NewString("key1").HashKey(): {Key: NewString("key1"), Value: NewString("value1")},
	})
	m2 := NewMap(map[HashKey]MapPair{
		NewString("key1").HashKey(): {Key: NewString("key1"), Value: NewString("value2")},
	})
	result := DeepMergeYAMLMaps(m1, m2)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestYAMLPathQuery(t *testing.T) {
	obj := NewMap(map[HashKey]MapPair{
		NewString("name").HashKey(): {Key: NewString("name"), Value: NewString("test")},
	})
	result := YAMLPathQuery(obj, ".name")
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestYAMLPathSet(t *testing.T) {
	obj := NewMap(map[HashKey]MapPair{
		NewString("name").HashKey(): {Key: NewString("name"), Value: NewString("test")},
	})
	result := YAMLPathSet(obj, ".name", NewString("updated"))
	if result == nil {
		t.Error("expected non-nil result")
	}
}
