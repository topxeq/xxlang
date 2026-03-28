// pkg/objects/toml_encoder.go
// TOML encoder - pure Go implementation
package objects

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// EncodeToml encodes a TomlValue to a TOML string
func EncodeToml(root *TomlValue) string {
	if root == nil {
		return ""
	}

	var sb strings.Builder
	encodeTomlValue(&sb, root, "", 0)
	return sb.String()
}

// encodeTomlValue encodes a TomlValue to the string builder
func encodeTomlValue(sb *strings.Builder, v *TomlValue, key string, depth int) {
	if v == nil {
		return
	}

	switch v.Type {
	case TomlTable:
		encodeTomlTable(sb, v, key, depth)
	case TomlArray:
		encodeTomlArray(sb, v, key, depth)
	case TomlString:
		encodeTomlString(sb, v.Value.(string), key)
	case TomlInteger:
		sb.WriteString(key)
		sb.WriteString(" = ")
		sb.WriteString(fmt.Sprintf("%d", v.Value.(int64)))
		sb.WriteString("\n")
	case TomlFloat:
		encodeTomlFloat(sb, v.Value.(float64), key)
	case TomlBoolean:
		sb.WriteString(key)
		sb.WriteString(" = ")
		if v.Value.(bool) {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
		sb.WriteString("\n")
	case TomlDatetime:
		encodeTomlDatetime(sb, v.Value.(time.Time), key)
	case TomlDate:
		encodeTomlDate(sb, v.Value.(time.Time), key)
	case TomlTime:
		encodeTomlTime(sb, v.Value.(time.Time), key)
	}
}

// encodeTomlTable encodes a table
func encodeTomlTable(sb *strings.Builder, v *TomlValue, key string, depth int) {
	table := v.Value.(map[string]*TomlValue)

	if depth == 0 {
		// Root level: output key-value pairs first, then tables
		// First, output all non-table values
		for _, k := range sortedKeys(table) {
			val := table[k]
			if val.Type != TomlTable && val.Type != TomlArray {
				encodeTomlValue(sb, val, k, depth)
			}
		}

		// Then, output tables and arrays
		for _, k := range sortedKeys(table) {
			val := table[k]
			if val.Type == TomlTable {
				// Check if it's an array of tables (contains only tables with numeric keys)
				subTable := val.Value.(map[string]*TomlValue)
				if isArrayTable(subTable) {
					encodeArrayTables(sb, subTable, k, depth)
				} else {
					sb.WriteString("\n[")
					sb.WriteString(escapeKey(k))
					sb.WriteString("]\n")
					encodeTomlTable(sb, val, "", depth+1)
				}
			} else if val.Type == TomlArray {
				arr := val.Value.([]*TomlValue)
				if len(arr) > 0 && arr[0].Type == TomlTable {
					// Array of tables
					encodeTomlArrayOfTables(sb, arr, k, depth)
				} else {
					// Regular array
					encodeTomlValue(sb, val, k, depth)
				}
			}
		}
	} else {
		// Nested level: output all key-value pairs
		for _, k := range sortedKeys(table) {
			val := table[k]
			if val.Type == TomlTable {
				subTable := val.Value.(map[string]*TomlValue)
				if isArrayTable(subTable) {
					encodeArrayTables(sb, subTable, k, depth)
				} else {
					sb.WriteString("\n[")
					if key != "" {
						sb.WriteString(escapeKey(key))
						sb.WriteString(".")
					}
					sb.WriteString(escapeKey(k))
					sb.WriteString("]\n")
					encodeTomlTable(sb, val, "", depth+1)
				}
			} else if val.Type == TomlArray {
				arr := val.Value.([]*TomlValue)
				if len(arr) > 0 && arr[0].Type == TomlTable {
					encodeTomlArrayOfTables(sb, arr, k, depth)
				} else {
					encodeTomlValue(sb, val, k, depth)
				}
			} else {
				encodeTomlValue(sb, val, k, depth)
			}
		}
	}
}

// encodeTomlArray encodes an array
func encodeTomlArray(sb *strings.Builder, v *TomlValue, key string, depth int) {
	arr := v.Value.([]*TomlValue)

	sb.WriteString(escapeKey(key))
	sb.WriteString(" = [\n")

	for _, elem := range arr {
		sb.WriteString("  ")
		encodeTomlInlineValue(sb, elem)
		sb.WriteString(",\n")
	}

	sb.WriteString("]\n")
}

// encodeTomlArrayOfTableses encodes an array of tables
func encodeTomlArrayOfTables(sb *strings.Builder, arr []*TomlValue, key string, depth int) {
	for _, table := range arr {
		sb.WriteString("\n[[")
		sb.WriteString(escapeKey(key))
		sb.WriteString("]]\n")

		t := table.Value.(map[string]*TomlValue)
		for _, k := range sortedKeys(t) {
			val := t[k]
			if val.Type == TomlTable {
				sb.WriteString("\n[")
				sb.WriteString(escapeKey(key))
				sb.WriteString(".")
				sb.WriteString(escapeKey(k))
				sb.WriteString("]\n")
				encodeTomlTable(sb, val, "", depth+1)
			} else {
				encodeTomlValue(sb, val, k, depth)
			}
		}
	}
}

// encodeArrayTables handles special array-table detection
func encodeArrayTables(sb *strings.Builder, table map[string]*TomlValue, prefix string, depth int) {
	for _, k := range sortedKeys(table) {
		val := table[k]
		sb.WriteString("\n[[")
		sb.WriteString(escapeKey(prefix))
		sb.WriteString("]]\n")
		encodeTomlTable(sb, val, "", depth+1)
	}
}

// encodeTomlInlineValue encodes a value inline (for arrays/inline tables)
func encodeTomlInlineValue(sb *strings.Builder, v *TomlValue) {
	switch v.Type {
	case TomlString:
		sb.WriteString(encodeQuotedString(v.Value.(string)))
	case TomlInteger:
		sb.WriteString(fmt.Sprintf("%d", v.Value.(int64)))
	case TomlFloat:
		sb.WriteString(fmt.Sprintf("%v", v.Value.(float64)))
	case TomlBoolean:
		if v.Value.(bool) {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case TomlDatetime:
		sb.WriteString(v.Value.(time.Time).Format(time.RFC3339))
	case TomlDate:
		sb.WriteString(v.Value.(time.Time).Format("2006-01-02"))
	case TomlTime:
		sb.WriteString(v.Value.(time.Time).Format("15:04:05"))
	case TomlArray:
		arr := v.Value.([]*TomlValue)
		sb.WriteString("[")
		for i, elem := range arr {
			if i > 0 {
				sb.WriteString(", ")
			}
			encodeTomlInlineValue(sb, elem)
		}
		sb.WriteString("]")
	case TomlTable:
		table := v.Value.(map[string]*TomlValue)
		sb.WriteString("{")
		first := true
		for _, k := range sortedKeys(table) {
			if !first {
				sb.WriteString(", ")
			}
			first = false
			sb.WriteString(escapeKey(k))
			sb.WriteString(" = ")
			encodeTomlInlineValue(sb, table[k])
		}
		sb.WriteString("}")
	}
}

// encodeTomlString encodes a string value
func encodeTomlString(sb *strings.Builder, s string, key string) {
	sb.WriteString(escapeKey(key))
	sb.WriteString(" = ")
	sb.WriteString(encodeQuotedString(s))
	sb.WriteString("\n")
}

// encodeTomlFloat encodes a float value
func encodeTomlFloat(sb *strings.Builder, f float64, key string) {
	sb.WriteString(escapeKey(key))
	sb.WriteString(" = ")

	if math.IsInf(f, 1) {
		sb.WriteString("inf")
	} else if math.IsInf(f, -1) {
		sb.WriteString("-inf")
	} else if math.IsNaN(f) {
		sb.WriteString("nan")
	} else {
		// Use enough precision
		if f == float64(int64(f)) {
			sb.WriteString(fmt.Sprintf("%.1f", f))
		} else {
			sb.WriteString(fmt.Sprintf("%v", f))
		}
	}
	sb.WriteString("\n")
}

// encodeTomlDatetime encodes a datetime value
func encodeTomlDatetime(sb *strings.Builder, t time.Time, key string) {
	sb.WriteString(escapeKey(key))
	sb.WriteString(" = ")
	sb.WriteString(t.Format(time.RFC3339))
	sb.WriteString("\n")
}

// encodeTomlDate encodes a date value
func encodeTomlDate(sb *strings.Builder, t time.Time, key string) {
	sb.WriteString(escapeKey(key))
	sb.WriteString(" = ")
	sb.WriteString(t.Format("2006-01-02"))
	sb.WriteString("\n")
}

// encodeTomlTime encodes a time value
func encodeTomlTime(sb *strings.Builder, t time.Time, key string) {
	sb.WriteString(escapeKey(key))
	sb.WriteString(" = ")
	sb.WriteString(t.Format("15:04:05"))
	sb.WriteString("\n")
}

// ============================================================
// Helper functions
// ============================================================

// escapeKey escapes a key if necessary
func escapeKey(key string) string {
	// Check if it's a valid bare key
	valid := true
	for _, ch := range key {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-') {
			valid = false
			break
		}
	}

	if valid && len(key) > 0 {
		return key
	}

	// Use quoted key
	return encodeQuotedString(key)
}

// encodeQuotedString encodes a string with quotes and escaping
func encodeQuotedString(s string) string {
	var sb strings.Builder
	sb.WriteString("\"")

	for _, ch := range s {
		switch ch {
		case '"':
			sb.WriteString("\\\"")
		case '\\':
			sb.WriteString("\\\\")
		case '\b':
			sb.WriteString("\\b")
		case '\t':
			sb.WriteString("\\t")
		case '\n':
			sb.WriteString("\\n")
		case '\f':
			sb.WriteString("\\f")
		case '\r':
			sb.WriteString("\\r")
		default:
			if ch < 32 || ch > 126 {
				sb.WriteString(fmt.Sprintf("\\u%04X", ch))
			} else {
				sb.WriteRune(ch)
			}
		}
	}

	sb.WriteString("\"")
	return sb.String()
}

// sortedKeys returns sorted keys of a map
func sortedKeys(m map[string]*TomlValue) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// isArrayTable checks if a table represents an array of tables
func isArrayTable(table map[string]*TomlValue) bool {
	// Simple heuristic: if all keys are numeric strings and all values are tables
	if len(table) == 0 {
		return false
	}
	for k, v := range table {
		// Check if key is numeric
		for _, ch := range k {
			if ch < '0' || ch > '9' {
				return false
			}
		}
		// Check if value is a table
		if v.Type != TomlTable {
			return false
		}
	}
	return true
}

// EncodeToToml encodes any value to TOML
func EncodeToToml(v interface{}) string {
	tv := CreateTomlValueFromInterface(v)
	return EncodeToml(tv)
}
