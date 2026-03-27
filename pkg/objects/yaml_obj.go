// pkg/objects/yaml_obj.go
// YAML object types and utility functions for Xxlang.
// This file implements comprehensive YAML 1.2 parsing and serialization without third-party libraries.
//
// Supported features:
// - Basic types: strings, integers, floats, booleans, null
// - Collections: mappings, sequences
// - Block styles: literal (|), folded (>), with chomping indicators (+,-) and indent indicators
// - Flow styles: inline mappings {}, inline sequences []
// - Anchors and aliases: &anchor, *alias
// - Merge keys: <<
// - Complex keys: ? explicit key syntax
// - Tags: !!str, !!int, !!float, !!bool, !!null, !!seq, !!map, !!timestamp, !!binary, !!set, !!omap
// - Multi-document: --- and ...
// - Comments: #
// - Special values: .inf, -.inf, .nan
// - Number formats: hex (0x), octal (0o), binary (0b), sexagesimal (60:00:00)
// - Precise error reporting with line and column numbers
package objects

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// YAMLError represents a YAML parsing error with location information
type YAMLError struct {
	Message  string
	Line     int
	Column   int
	Context  string
	Original error
}

// Error returns the error message with location
func (e *YAMLError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("YAML error at line %d, column %d: %s", e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("YAML error: %s", e.Message)
}

// Unwrap returns the underlying error
func (e *YAMLError) Unwrap() error {
	return e.Original
}

// newYAMLError creates a YAML error with location
func (p *yamlParser) newYAMLError(message string, args ...interface{}) *YAMLError {
	line := p.lineNum + 1
	col := 1
	if p.lineNum < len(p.lines) {
		col = p.getIndent(p.lines[p.lineNum]) + 1
	}
	return &YAMLError{
		Message: fmt.Sprintf(message, args...),
		Line:    line,
		Column:  col,
	}
}

// YAMLDocument represents a YAML document that can be loaded and queried
type YAMLDocument struct {
	Root   Object
	Source string
}

// Type returns the object type
func (y *YAMLDocument) Type() ObjectType { return YAMLDocumentType }

// TypeTag returns the type tag for fast type checking
func (y *YAMLDocument) TypeTag() TypeTag { return TagYAMLDocument }

// Inspect returns a string representation
func (y *YAMLDocument) Inspect() string {
	return fmt.Sprintf("YAMLDocument(%s)", y.Root.Inspect())
}

// ToBool returns true if the document is not nil
func (y *YAMLDocument) ToBool() *Bool { return TRUE }

// HashKey returns a hash key (always 0 for YAML documents)
func (y *YAMLDocument) HashKey() HashKey {
	return HashKey{Type: YAMLDocumentType, Value: 0}
}

// ParseYAML parses a YAML string and returns an Xxlang Object.
// This is the main entry point for YAML parsing.
func ParseYAML(s string) (Object, error) {
	parser := newYAMLParser(s)
	return parser.parse()
}

// ParseYAMLDocuments parses a YAML string with multiple documents.
func ParseYAMLDocuments(s string) ([]Object, error) {
	parser := newYAMLParser(s)
	return parser.parseDocuments()
}

// yamlParser is a stateful YAML parser supporting YAML 1.2 features
type yamlParser struct {
	lines       []string
	lineNum     int
	anchors     map[string]Object
	anchorLines map[string]int // Line numbers where anchors are defined
	tags        map[string]string
	docStart    int
	docEnd      int
	lastTag     string
	forwardRefs map[string][]int // Forward references: anchor -> line numbers where referenced
}

// newYAMLParser creates a new YAML parser
func newYAMLParser(s string) *yamlParser {
	return &yamlParser{
		lines:       strings.Split(s, "\n"),
		anchors:     make(map[string]Object),
		anchorLines: make(map[string]int),
		tags:        make(map[string]string),
		forwardRefs: make(map[string][]int),
	}
}

// parse starts parsing from the current position
func (p *yamlParser) parse() (Object, error) {
	// First pass: scan for all anchor definitions to support forward references
	p.scanAnchors()

	// Reset to beginning and skip empty lines
	p.lineNum = 0
	p.skipEmptyCommentsAndDirectives()

	if p.lineNum >= len(p.lines) {
		return NULL, nil
	}

	return p.parseValue(0)
}

// scanAnchors performs a first pass to find all anchor definitions
func (p *yamlParser) scanAnchors() {
	for i := 0; i < len(p.lines); i++ {
		line := p.lines[i]
		trimmed := strings.TrimLeft(line, " \t")

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Look for anchor definitions &name
		if idx := strings.Index(trimmed, "&"); idx != -1 {
			afterAmp := trimmed[idx+1:]
			// Extract anchor name
			endIdx := strings.IndexAny(afterAmp, " \t\n:,[]{}#")
			var anchorName string
			if endIdx == -1 {
				anchorName = afterAmp
			} else {
				anchorName = afterAmp[:endIdx]
			}
			if anchorName != "" {
				p.anchorLines[anchorName] = i
			}
		}
	}
}

// parseDocuments parses multiple YAML documents separated by ---
func (p *yamlParser) parseDocuments() ([]Object, error) {
	// First pass: scan for all anchor definitions to support forward references
	p.scanAnchors()
	p.lineNum = 0

	var docs []Object

	for p.lineNum < len(p.lines) {
		p.skipEmptyCommentsAndDirectives()
		if p.lineNum >= len(p.lines) {
			break
		}

		line := strings.TrimSpace(p.lines[p.lineNum])
		if line == "---" {
			p.lineNum++
			p.skipEmptyCommentsAndDirectives()
		}

		if p.lineNum >= len(p.lines) {
			break
		}

		line = strings.TrimSpace(p.lines[p.lineNum])
		if line == "..." {
			p.lineNum++
			continue
		}

		doc, err := p.parseValue(0)
		if err != nil {
			return nil, err
		}

		docs = append(docs, doc)

		for p.lineNum < len(p.lines) {
			line = strings.TrimSpace(p.lines[p.lineNum])
			if line == "---" || line == "..." {
				break
			}
			p.lineNum++
		}
	}

	if len(docs) == 0 {
		return []Object{NULL}, nil
	}

	return docs, nil
}

// parseValue parses a value at the given indentation level
func (p *yamlParser) parseValue(indent int) (Object, error) {
	if p.lineNum >= len(p.lines) {
		return NULL, nil
	}

	line := p.lines[p.lineNum]
	trimmed := strings.TrimLeft(line, " \t")

	if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "..." {
		return NULL, nil
	}

	// Handle explicit key (?)
	if strings.HasPrefix(trimmed, "?") && (len(trimmed) == 1 || trimmed[1] == ' ' || trimmed[1] == '\t' || trimmed[1] == '\n') {
		return p.parseExplicitKeyValue(indent)
	}

	// Check for tag at the start (e.g., !!set, !!omap, !local)
	if strings.HasPrefix(trimmed, "!!") || (strings.HasPrefix(trimmed, "!") && len(trimmed) > 1 && trimmed[1] != '!') {
		return p.parseTaggedValue(trimmed, indent)
	}

	// Check for anchor definition
	if strings.HasPrefix(trimmed, "&") {
		return p.parseAnchoredValue(indent)
	}

	// Check for alias reference
	if strings.HasPrefix(trimmed, "*") {
		return p.parseAlias(trimmed)
	}

	// Check for sequence item
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "-\t") || trimmed == "-" {
		return p.parseSequence(indent)
	}

	// Check for flow style sequence
	if strings.HasPrefix(trimmed, "[") {
		return p.parseFlowSequence(trimmed)
	}

	// Check for flow style mapping
	if strings.HasPrefix(trimmed, "{") {
		return p.parseFlowMapping(trimmed)
	}

	// Check for mapping (has colon)
	if colonIdx := p.findKeyColon(trimmed); colonIdx != -1 {
		return p.parseMapping(indent)
	}

	// Check for block scalar
	if strings.HasPrefix(trimmed, "|") || strings.HasPrefix(trimmed, ">") {
		return p.parseBlockScalar(trimmed, indent)
	}

	// Parse as scalar
	return p.parseScalar(trimmed)
}

// parseExplicitKeyValue parses an explicit key-value pair using ? and : syntax
func (p *yamlParser) parseExplicitKeyValue(indent int) (Object, error) {
	line := p.lines[p.lineNum]
	trimmed := strings.TrimLeft(line, " \t")

	// Skip the ? marker
	afterQuestion := strings.TrimSpace(trimmed[1:])

	p.lineNum++

	var key Object
	var err error

	if afterQuestion == "" {
		// Key is on next line(s)
		if p.lineNum < len(p.lines) {
			nextLine := p.lines[p.lineNum]
			nextIndent := p.getIndent(nextLine)
			if nextIndent > indent {
				key, err = p.parseValue(nextIndent)
				if err != nil {
					return nil, err
				}
			} else {
				key = NULL
			}
		} else {
			key = NULL
		}
	} else {
		// Key is inline after ?
		// Check if it starts a collection (sequence or mapping)
		if strings.HasPrefix(afterQuestion, "- ") || afterQuestion == "-" {
			// It's a sequence - need to parse as a sequence
			key, err = p.parseSequenceFromFirstItem(afterQuestion, indent+2)
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(afterQuestion, "{") {
			// Inline mapping
			key, err = p.parseFlowMapping(afterQuestion)
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(afterQuestion, "[") {
			// Inline sequence
			key, err = p.parseFlowSequence(afterQuestion)
			if err != nil {
				return nil, err
			}
		} else if colonIdx := p.findKeyColon(afterQuestion); colonIdx != -1 {
			// Inline mapping like "? key: value"
			key, err = p.parseInlineMapping(afterQuestion, indent+2)
			if err != nil {
				return nil, err
			}
		} else {
			// Simple scalar key
			key, err = p.parseScalar(afterQuestion)
			if err != nil {
				return nil, err
			}
		}
	}

	// Now look for : to get the value
	p.skipEmptyCommentsAndDirectives()
	if p.lineNum >= len(p.lines) {
		// No value, return a map with just the key mapping to null
		return p.wrapExplicitKey(key, NULL), nil
	}

	line = p.lines[p.lineNum]
	trimmed = strings.TrimLeft(line, " \t")

	if strings.HasPrefix(trimmed, ":") {
		// Value indicator found
		valueStr := strings.TrimSpace(trimmed[1:])
		p.lineNum++

		var value Object
		if valueStr == "" {
			// Value is on next line(s)
			if p.lineNum < len(p.lines) {
				nextLine := p.lines[p.lineNum]
				nextIndent := p.getIndent(nextLine)
				if nextIndent > indent {
					value, err = p.parseValue(nextIndent)
					if err != nil {
						return nil, err
					}
				} else {
					value = NULL
				}
			} else {
				value = NULL
			}
		} else {
			value, err = p.parseInlineValue(valueStr, indent)
			if err != nil {
				return nil, err
			}
		}

		return p.wrapExplicitKey(key, value), nil
	}

	// No value indicator, key maps to null
	return p.wrapExplicitKey(key, NULL), nil
}

// wrapExplicitKey wraps an explicit key into a map structure
func (p *yamlParser) wrapExplicitKey(key, value Object) Object {
	pairs := make(map[HashKey]MapPair)
	pairs[key.HashKey()] = MapPair{
		Key:   key,
		Value: value,
	}
	return NewMap(pairs)
}

// parseAnchoredValue parses a value with an anchor definition
func (p *yamlParser) parseAnchoredValue(indent int) (Object, error) {
	line := p.lines[p.lineNum]
	trimmed := strings.TrimLeft(line, " \t")

	anchorStart := 1
	anchorEnd := strings.IndexAny(trimmed[anchorStart:], " \t\n:")
	if anchorEnd == -1 {
		anchorEnd = len(trimmed)
	} else {
		anchorEnd += anchorStart
	}

	anchorName := trimmed[anchorStart:anchorEnd]
	remaining := strings.TrimSpace(trimmed[anchorEnd:])

	var value Object
	var err error

	if remaining == "" {
		p.lineNum++

		if p.lineNum < len(p.lines) {
			nextLine := p.lines[p.lineNum]
			nextIndent := p.getIndent(nextLine)
			if nextIndent > indent {
				value, err = p.parseValue(nextIndent)
				if err != nil {
					return nil, err
				}
			} else {
				value = NULL
			}
		} else {
			value = NULL
		}
	} else if strings.HasPrefix(remaining, ":") {
		p.lineNum++
		return p.parseMappingKeyWithAnchor(anchorName, remaining[1:], indent)
	} else {
		p.lineNum++
		value, err = p.parseScalar(remaining)
		if err != nil {
			return nil, err
		}
	}

	p.anchors[anchorName] = value

	return value, nil
}

// parseMappingKeyWithAnchor parses a mapping key that has an anchor
func (p *yamlParser) parseMappingKeyWithAnchor(anchorName, valueStr string, indent int) (Object, error) {
	p.lineNum--
	return p.parseMapping(indent)
}

// parseAlias parses an alias reference (*name)
// Supports forward references by parsing anchors on demand
func (p *yamlParser) parseAlias(trimmed string) (Object, error) {
	p.lineNum++

	aliasName := strings.TrimSpace(trimmed[1:])

	if spaceIdx := strings.IndexAny(aliasName, " \t,]}"); spaceIdx != -1 {
		aliasName = aliasName[:spaceIdx]
	}

	// Check if we already parsed this anchor
	if value, ok := p.anchors[aliasName]; ok {
		return yamlDeepCopyObject(value), nil
	}

	// Check if this anchor is defined elsewhere (forward reference)
	if anchorLine, exists := p.anchorLines[aliasName]; exists {
		// Save current position
		savedLineNum := p.lineNum

		// Jump to anchor definition and parse it
		p.lineNum = anchorLine
		if _, err := p.parseValue(0); err != nil {
			p.lineNum = savedLineNum
			return NULL, err
		}

		// Restore position
		p.lineNum = savedLineNum

		// Now the anchor should be parsed
		if value, ok := p.anchors[aliasName]; ok {
			return yamlDeepCopyObject(value), nil
		}
	}

	return NULL, p.newYAMLError("unknown alias: %s", aliasName)
}

// parseSequence parses a YAML sequence (list)
func (p *yamlParser) parseSequence(baseIndent int) (Object, error) {
	var elements []Object

	for p.lineNum < len(p.lines) {
		p.skipEmptyCommentsAndDirectives()
		if p.lineNum >= len(p.lines) {
			break
		}

		line := p.lines[p.lineNum]
		currentIndent := p.getIndent(line)

		if currentIndent < baseIndent {
			break
		}

		trimmed := strings.TrimLeft(line, " \t")

		if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "..." {
			break
		}

		if !strings.HasPrefix(trimmed, "-") {
			break
		}

		afterDash := trimmed[1:]

		if strings.HasPrefix(strings.TrimSpace(afterDash), "&") {
			p.lineNum++
			afterDashTrimmed := strings.TrimSpace(afterDash)
			endIdx := strings.IndexAny(afterDashTrimmed[1:], " \t\n")
			if endIdx == -1 {
				endIdx = len(afterDashTrimmed) - 1
			} else {
				endIdx++
			}
			anchorName := afterDashTrimmed[1:endIdx]
			remaining := strings.TrimSpace(afterDashTrimmed[endIdx+1:])

			var val Object
			var err error

			if remaining == "" {
				if p.lineNum < len(p.lines) {
					nextLine := p.lines[p.lineNum]
					nextIndent := p.getIndent(nextLine)
					if nextIndent > currentIndent {
						val, err = p.parseValue(nextIndent)
						if err != nil {
							return nil, err
						}
					} else {
						val = NULL
					}
				} else {
					val = NULL
				}
			} else {
				val, err = p.parseInlineValue(remaining, currentIndent)
				if err != nil {
					return nil, err
				}
			}

			p.anchors[anchorName] = val
			elements = append(elements, val)
			continue
		}

		if strings.HasPrefix(strings.TrimSpace(afterDash), "*") {
			p.lineNum++
			aliasName := strings.TrimSpace(strings.TrimSpace(afterDash)[1:])
			if spaceIdx := strings.IndexAny(aliasName, " \t"); spaceIdx != -1 {
				aliasName = aliasName[:spaceIdx]
			}
			if val, ok := p.anchors[aliasName]; ok {
				elements = append(elements, yamlDeepCopyObject(val))
			} else {
				elements = append(elements, NULL)
			}
			continue
		}

		p.lineNum++

		if strings.TrimSpace(afterDash) == "" {
			if p.lineNum < len(p.lines) {
				nextLine := p.lines[p.lineNum]
				nextIndent := p.getIndent(nextLine)
				trimmedNext := strings.TrimLeft(nextLine, " \t")

				if nextIndent > currentIndent && !strings.HasPrefix(trimmedNext, "-") {
					val, err := p.parseValue(nextIndent)
					if err != nil {
						return nil, err
					}
					elements = append(elements, val)
					continue
				}
			}
			elements = append(elements, NULL)
		} else {
			val, err := p.parseInlineValue(strings.TrimSpace(afterDash), currentIndent)
			if err != nil {
				return nil, err
			}
			elements = append(elements, val)
		}
	}

	return NewArray(elements), nil
}

// parseSequenceFromFirstItem parses a sequence where the first item is already extracted
// This is used for explicit key syntax like "? - item1\n  - item2"
func (p *yamlParser) parseSequenceFromFirstItem(firstItem string, baseIndent int) (Object, error) {
	var elements []Object

	// Parse the first item (already extracted, starts with "-")
	afterDash := strings.TrimSpace(firstItem[1:])

	var firstVal Object
	var err error

	if afterDash == "" {
		firstVal = NULL
	} else {
		firstVal, err = p.parseInlineValue(afterDash, baseIndent)
		if err != nil {
			return nil, err
		}
	}
	elements = append(elements, firstVal)

	// Continue parsing more sequence items from subsequent lines
	for p.lineNum < len(p.lines) {
		p.skipEmptyCommentsAndDirectives()
		if p.lineNum >= len(p.lines) {
			break
		}

		line := p.lines[p.lineNum]
		currentIndent := p.getIndent(line)

		if currentIndent < baseIndent {
			break
		}

		trimmed := strings.TrimLeft(line, " \t")

		if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "..." {
			break
		}

		// Check for mapping key - stop if we see one
		if p.isNewMappingKey(trimmed) {
			break
		}

		if !strings.HasPrefix(trimmed, "-") {
			break
		}

		afterDash := strings.TrimSpace(trimmed[1:])
		p.lineNum++

		var val Object
		if afterDash == "" {
			if p.lineNum < len(p.lines) {
				nextLine := p.lines[p.lineNum]
				nextIndent := p.getIndent(nextLine)
				trimmedNext := strings.TrimLeft(nextLine, " \t")

				if nextIndent > currentIndent && !strings.HasPrefix(trimmedNext, "-") {
					val, err = p.parseValue(nextIndent)
					if err != nil {
						return nil, err
					}
				} else {
					val = NULL
				}
			} else {
				val = NULL
			}
		} else {
			val, err = p.parseInlineValue(afterDash, currentIndent)
			if err != nil {
				return nil, err
			}
		}
		elements = append(elements, val)
	}

	return NewArray(elements), nil
}

// parseInlineValue parses a value that appears inline (after a key: or -)
func (p *yamlParser) parseInlineValue(s string, baseIndent int) (Object, error) {
	s = strings.TrimSpace(s)

	if s == "" {
		return NULL, nil
	}

	// Handle tag
	if strings.HasPrefix(s, "!!") || (strings.HasPrefix(s, "!") && !strings.HasPrefix(s, "!=")) {
		return p.parseTaggedValue(s, baseIndent)
	}

	// Handle anchor
	if strings.HasPrefix(s, "&") {
		endIdx := strings.IndexAny(s[1:], " \t")
		if endIdx == -1 {
			anchorName := s[1:]
			p.anchors[anchorName] = NULL
			return NULL, nil
		}
		endIdx++
		anchorName := s[1:endIdx]
		remaining := strings.TrimSpace(s[endIdx+1:])

		var val Object
		var err error

		if remaining == "" {
			val = NULL
		} else {
			val, err = p.parseInlineValue(remaining, baseIndent)
			if err != nil {
				return nil, err
			}
		}

		p.anchors[anchorName] = val
		return val, nil
	}

	// Handle alias
	if strings.HasPrefix(s, "*") {
		endIdx := strings.IndexAny(s[1:], " \t,]}")
		var aliasName string
		if endIdx == -1 {
			aliasName = s[1:]
		} else {
			aliasName = s[1 : endIdx+1]
		}
		if val, ok := p.anchors[aliasName]; ok {
			return yamlDeepCopyObject(val), nil
		}
		// Check for forward reference
		if anchorLine, exists := p.anchorLines[aliasName]; exists {
			savedLineNum := p.lineNum
			p.lineNum = anchorLine
			if _, err := p.parseValue(0); err != nil {
				p.lineNum = savedLineNum
				return NULL, err
			}
			p.lineNum = savedLineNum
			if val, ok := p.anchors[aliasName]; ok {
				return yamlDeepCopyObject(val), nil
			}
		}
		return NULL, p.newYAMLError("unknown alias: %s", aliasName)
	}

	// Check for flow sequence
	if strings.HasPrefix(s, "[") {
		return p.parseFlowSequence(s)
	}

	// Check for flow mapping
	if strings.HasPrefix(s, "{") {
		return p.parseFlowMapping(s)
	}

	// Check for block scalar
	if strings.HasPrefix(s, "|") || strings.HasPrefix(s, ">") {
		return p.parseBlockScalar(s, baseIndent)
	}

	// Check for inline mapping (contains unquoted colon)
	if colonIdx := p.findKeyColon(s); colonIdx != -1 {
		return p.parseInlineMapping(s, baseIndent+2)
	}

	// Parse as scalar
	return p.parseScalar(s)
}

// parseTaggedValue parses a value with a YAML tag
func (p *yamlParser) parseTaggedValue(s string, baseIndent int) (Object, error) {
	// Extract tag
	var tag string
	var value string

	if strings.HasPrefix(s, "!!") {
		endIdx := strings.IndexAny(s[2:], " \t\n")
		if endIdx == -1 {
			tag = s
			value = ""
		} else {
			tag = s[:endIdx+2]
			value = strings.TrimSpace(s[endIdx+2:])
		}
	} else {
		endIdx := strings.IndexAny(s[1:], " \t\n")
		if endIdx == -1 {
			tag = s
			value = ""
		} else {
			tag = s[:endIdx+1]
			value = strings.TrimSpace(s[endIdx+1:])
		}
	}

	// Handle different tags
	switch tag {
	case "!!str":
		// Force string type - always return as string
		if value == "" {
			return NewString(""), nil
		}
		// Remove quotes if present, otherwise return as-is
		if unquoted, err := p.unquoteYAMLStringChecked(value); err == nil && unquoted != value {
			return NewString(unquoted), nil
		}
		return NewString(value), nil

	case "!!int":
		if value == "" {
			return NewInt(0), nil
		}
		// Try to convert string to int
		if unquoted, err := p.unquoteYAMLStringChecked(value); err == nil && unquoted != value {
			value = unquoted
		}
		if i, ok := parseYAMLInt(value); ok {
			return NewInt(i), nil
		}
		// Try float to int conversion
		if f, ok := parseYAMLFloat(value); ok {
			return NewInt(int64(f)), nil
		}
		return NewInt(0), nil

	case "!!float":
		if value == "" {
			return NewFloat(0), nil
		}
		// Try to convert string to float
		if unquoted, err := p.unquoteYAMLStringChecked(value); err == nil && unquoted != value {
			value = unquoted
		}
		if f, ok := parseYAMLFloat(value); ok {
			return NewFloat(f), nil
		}
		if f, ok := parseYAMLSpecialFloat(value); ok {
			return NewFloat(f), nil
		}
		// Try int to float conversion
		if i, ok := parseYAMLInt(value); ok {
			return NewFloat(float64(i)), nil
		}
		return NewFloat(0), nil

	case "!!bool":
		if value == "" {
			return FALSE, nil
		}
		// Try to convert string to bool
		if unquoted, err := p.unquoteYAMLStringChecked(value); err == nil && unquoted != value {
			value = unquoted
		}
		if b, ok := parseYAMLBool(value); ok {
			return b, nil
		}
		// Any non-empty value that's not explicitly false-ish is true
		lower := strings.ToLower(strings.TrimSpace(value))
		if lower == "false" || lower == "no" || lower == "n" || lower == "0" || lower == "" {
			return FALSE, nil
		}
		return TRUE, nil

	case "!!null":
		return NULL, nil

	case "!!timestamp", "!!date":
		return p.parseTimestamp(value)

	case "!!binary":
		return p.parseBinary(value)

	case "!!seq":
		if value == "" {
			return NewArray([]Object{}), nil
		}
		if strings.HasPrefix(value, "[") {
			return p.parseFlowSequence(value)
		}
		return p.parseInlineValue(value, baseIndent)

	case "!!set":
		// !!set is a mapping where all values are null
		// The content follows on subsequent lines as explicit keys
		if value != "" {
			if strings.HasPrefix(value, "{") {
				// Flow style set
				return p.parseFlowSet(value)
			}
			// Inline value after tag
			return p.parseInlineValue(value, baseIndent)
		}
		// Move to next line and parse set entries
		p.lineNum++
		return p.parseSetContent(baseIndent)

	case "!!omap":
		// !!omap is an ordered map, represented as a sequence of single-key mappings
		if value != "" {
			if strings.HasPrefix(value, "[") {
				return p.parseFlowSequence(value)
			}
			return p.parseInlineValue(value, baseIndent)
		}
		// Move to next line and parse omap entries
		p.lineNum++
		return p.parseOMapContent(baseIndent)

	case "!!map":
		if value == "" {
			if strings.HasPrefix(value, "{") {
				return p.parseFlowMapping(value)
			}
			return NewMap(make(map[HashKey]MapPair)), nil
		}
		return p.parseInlineValue(value, baseIndent)

	default:
		// Local or unknown tag - parse value normally
		if value != "" {
			return p.parseInlineValue(value, baseIndent)
		}
		// Tag without value - check if there's content on next lines
		p.lineNum++
		if p.lineNum < len(p.lines) {
			nextLine := p.lines[p.lineNum]
			nextIndent := p.getIndent(nextLine)
			if nextIndent > baseIndent {
				return p.parseValue(nextIndent)
			}
		}
		p.lineNum--
		return NewString(s), nil
	}
}

// parseFlowSet parses a flow-style set like {a, b, c}
func (p *yamlParser) parseFlowSet(s string) (Object, error) {
	s = strings.TrimSpace(s)

	if !strings.HasPrefix(s, "{") {
		return nil, p.newYAMLError("invalid flow set: %s", s)
	}

	// Check if we need to collect more lines for multiline flow set
	if !strings.HasSuffix(s, "}") || !p.isFlowComplete(s, '{', '}') {
		collected := s

		for p.lineNum < len(p.lines) {
			p.lineNum++
			if p.lineNum >= len(p.lines) {
				break
			}
			nextLine := p.lines[p.lineNum]
			collected += " " + strings.TrimSpace(nextLine)

			if p.isFlowComplete(collected, '{', '}') {
				break
			}
		}

		if !p.isFlowComplete(collected, '{', '}') {
			return nil, p.newYAMLError("invalid flow set: %s", collected)
		}
		s = collected
	}

	content := strings.TrimSpace(s[1 : len(s)-1])
	if content == "" {
		return NewMap(make(map[HashKey]MapPair)), nil
	}

	// Parse as set - each element is a key with null value
	pairs := make(map[HashKey]MapPair)
	elements := p.parseFlowElements(content)

	for _, elem := range elements {
		// In a set, each element is just a key (no value, or value is implicitly null)
		key, err := p.parseInlineValue(elem, 0)
		if err != nil {
			return nil, err
		}
		pairs[key.HashKey()] = MapPair{
			Key:   key,
			Value: NULL,
		}
	}

	return NewMap(pairs), nil
}

// parseSetContent parses block-style set content (explicit keys with ?)
func (p *yamlParser) parseSetContent(baseIndent int) (Object, error) {
	pairs := make(map[HashKey]MapPair)

	for p.lineNum < len(p.lines) {
		p.skipEmptyCommentsAndDirectives()
		if p.lineNum >= len(p.lines) {
			break
		}

		line := p.lines[p.lineNum]
		currentIndent := p.getIndent(line)

		// Check if we're still at or below the set's indent level
		// When baseIndent is 0, we need to accept any non-negative indent
		if baseIndent > 0 && currentIndent < baseIndent {
			break
		}

		trimmed := strings.TrimLeft(line, " \t")

		if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "..." {
			break
		}

		// Set entries are explicit keys (?)
		// Check that ? is followed by space, tab, or end of value
		if strings.HasPrefix(trimmed, "?") && (len(trimmed) == 1 || trimmed[1] == ' ' || trimmed[1] == '\t' || trimmed[1] == '\n') {
			keyStr := strings.TrimSpace(trimmed[1:])
			p.lineNum++

			var key Object
			var err error

			if keyStr == "" {
				// Key on next line
				if p.lineNum < len(p.lines) {
					nextLine := p.lines[p.lineNum]
					nextIndent := p.getIndent(nextLine)
					if nextIndent > currentIndent {
						key, err = p.parseValue(nextIndent)
						if err != nil {
							return nil, err
						}
					} else {
						key = NULL
					}
				} else {
					key = NULL
				}
			} else {
				key, err = p.parseInlineValue(keyStr, currentIndent)
				if err != nil {
					return nil, err
				}
			}

			pairs[key.HashKey()] = MapPair{
				Key:   key,
				Value: NULL,
			}
		} else {
			// Not a set entry - check if we should break
			// If baseIndent is 0 and we see content at indent 0 that's not a set entry, break
			if currentIndent <= baseIndent {
				break
			}
		}
	}

	return NewMap(pairs), nil
}

// parseOMapContent parses block-style omap content
func (p *yamlParser) parseOMapContent(baseIndent int) (Object, error) {
	// OMap is represented as a sequence of single-key mappings
	// We return an OrderedMap to preserve order
	om := NewOrderedMap()

	for p.lineNum < len(p.lines) {
		p.skipEmptyCommentsAndDirectives()
		if p.lineNum >= len(p.lines) {
			break
		}

		line := p.lines[p.lineNum]
		currentIndent := p.getIndent(line)

		if currentIndent < baseIndent {
			break
		}

		trimmed := strings.TrimLeft(line, " \t")

		if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "..." {
			break
		}

		// OMap entries are sequence items with single-key mappings
		if strings.HasPrefix(trimmed, "- ") || trimmed == "-" {
			p.lineNum++
			afterDash := strings.TrimSpace(trimmed[1:])

			if afterDash == "" {
				// Key-value on next line
				if p.lineNum < len(p.lines) {
					nextLine := p.lines[p.lineNum]
					nextIndent := p.getIndent(nextLine)
					if nextIndent > currentIndent {
						entry, err := p.parseValue(nextIndent)
						if err != nil {
							return nil, err
						}
						if entryMap, ok := entry.(*Map); ok {
							for _, pair := range entryMap.Pairs {
								om.Set(pair.Key, pair.Value)
								break // Only first key-value pair per entry
							}
						}
					}
				}
			} else {
				// Parse the inline mapping
				entry, err := p.parseInlineValue(afterDash, currentIndent)
				if err != nil {
					return nil, err
				}
				if entryMap, ok := entry.(*Map); ok {
					for _, pair := range entryMap.Pairs {
						om.Set(pair.Key, pair.Value)
						break // Only first key-value pair per entry
					}
				}
			}
		} else {
			break
		}
	}

	return om, nil
}

// parseTimestamp parses a YAML timestamp
func (p *yamlParser) parseTimestamp(s string) (Object, error) {
	if s == "" {
		return NewString(""), nil
	}

	// Common timestamp formats
	formats := []string{
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return NewString(t.Format(time.RFC3339)), nil
		}
	}

	// Return as string if parsing fails
	return NewString(s), nil
}

// parseBinary parses a !!binary value (base64)
func (p *yamlParser) parseBinary(s string) (Object, error) {
	if s == "" {
		return NewArray([]Object{}), nil
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		// Try URL encoding
		decoded, err = base64.URLEncoding.DecodeString(s)
		if err != nil {
			return NewString(s), nil
		}
	}

	// Convert to array of integers
	elements := make([]Object, len(decoded))
	for i, b := range decoded {
		elements[i] = NewInt(int64(b))
	}

	return NewArray(elements), nil
}

// parseInlineMapping parses a mapping that starts inline
func (p *yamlParser) parseInlineMapping(firstLine string, baseIndent int) (Object, error) {
	pairs := make(map[HashKey]MapPair)

	collector := newYAMLKVCollector(firstLine)
	key, valueStr, err := collector.next()
	if err != nil {
		return nil, err
	}

	key = unquoteYAMLKey(key)

	var value Object
	if valueStr == "" {
		if p.lineNum < len(p.lines) {
			nextLine := p.lines[p.lineNum]
			nextIndent := p.getIndent(nextLine)
			trimmedNext := strings.TrimLeft(nextLine, " \t")

			if nextIndent >= baseIndent && !strings.HasPrefix(trimmedNext, "-") && !p.isNewMappingKey(trimmedNext) {
				var parseErr error
				value, parseErr = p.parseValue(nextIndent)
				if parseErr != nil {
					return nil, parseErr
				}
			} else {
				value = NULL
			}
		} else {
			value = NULL
		}
	} else {
		value, err = p.parseInlineValue(valueStr, baseIndent)
		if err != nil {
			return nil, err
		}
	}

	keyObj := NewString(key)
	pairs[keyObj.HashKey()] = MapPair{
		Key:   keyObj,
		Value: value,
	}

	for p.lineNum < len(p.lines) {
		p.skipEmptyCommentsAndDirectives()
		if p.lineNum >= len(p.lines) {
			break
		}

		line := p.lines[p.lineNum]
		currentIndent := p.getIndent(line)

		if currentIndent < baseIndent {
			break
		}

		trimmed := strings.TrimLeft(line, " \t")

		if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "..." {
			break
		}

		if strings.HasPrefix(trimmed, "-") && (len(trimmed) == 1 || trimmed[1] == ' ' || trimmed[1] == '\t') {
			break
		}

		if strings.HasPrefix(trimmed, "<<:") || strings.HasPrefix(trimmed, "<< :") {
			p.lineNum++
			mergeValue := strings.TrimPrefix(trimmed, "<<")
			mergeValue = strings.TrimPrefix(strings.TrimSpace(mergeValue), ":")
			mergeValue = strings.TrimSpace(mergeValue)

			var mergeObj Object
			var mergeErr error
			if mergeValue == "" {
				if p.lineNum < len(p.lines) {
					nextLine := p.lines[p.lineNum]
					nextIndent := p.getIndent(nextLine)
					if nextIndent > currentIndent {
						mergeObj, mergeErr = p.parseValue(nextIndent)
						if mergeErr != nil {
							return nil, mergeErr
						}
					}
				}
			} else {
				mergeObj, mergeErr = p.parseInlineValue(mergeValue, currentIndent)
				if mergeErr != nil {
					return nil, mergeErr
				}
			}

			if mergeMap, ok := mergeObj.(*Map); ok {
				for _, pair := range mergeMap.Pairs {
					if _, exists := pairs[pair.Key.HashKey()]; !exists {
						pairs[pair.Key.HashKey()] = pair
					}
				}
			} else if mergeArr, ok := mergeObj.(*Array); ok {
				for _, elem := range mergeArr.Elements {
					if mergeMap, ok := elem.(*Map); ok {
						for _, pair := range mergeMap.Pairs {
							if _, exists := pairs[pair.Key.HashKey()]; !exists {
								pairs[pair.Key.HashKey()] = pair
							}
						}
					}
				}
			}
			continue
		}

		cIdx := p.findKeyColon(trimmed)
		if cIdx == -1 {
			break
		}

		p.lineNum++

		k := strings.TrimSpace(trimmed[:cIdx])
		k = unquoteYAMLKey(k)
		vStr := strings.TrimSpace(trimmed[cIdx+1:])

		var v Object
		if vStr == "" {
			if p.lineNum < len(p.lines) {
				nextLine := p.lines[p.lineNum]
				nextIndent := p.getIndent(nextLine)
				if nextIndent > currentIndent {
					v, err = p.parseValue(nextIndent)
					if err != nil {
						return nil, err
					}
				} else {
					v = NULL
				}
			} else {
				v = NULL
			}
		} else {
			v, err = p.parseInlineValue(vStr, currentIndent)
			if err != nil {
				return nil, err
			}
		}

		kObj := NewString(k)
		pairs[kObj.HashKey()] = MapPair{
			Key:   kObj,
			Value: v,
		}
	}

	return NewMap(pairs), nil
}

// parseMapping parses a YAML mapping (dictionary)
func (p *yamlParser) parseMapping(baseIndent int) (Object, error) {
	pairs := make(map[HashKey]MapPair)

	for p.lineNum < len(p.lines) {
		p.skipEmptyCommentsAndDirectives()
		if p.lineNum >= len(p.lines) {
			break
		}

		line := p.lines[p.lineNum]
		currentIndent := p.getIndent(line)

		if currentIndent < baseIndent {
			break
		}

		trimmed := strings.TrimLeft(line, " \t")

		if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "..." {
			break
		}

		if strings.HasPrefix(trimmed, "-") && (len(trimmed) == 1 || trimmed[1] == ' ' || trimmed[1] == '\t') {
			break
		}

		if strings.HasPrefix(trimmed, "<<:") || strings.HasPrefix(trimmed, "<< :") {
			p.lineNum++
			mergeValue := strings.TrimPrefix(trimmed, "<<")
			mergeValue = strings.TrimPrefix(strings.TrimSpace(mergeValue), ":")
			mergeValue = strings.TrimSpace(mergeValue)

			var mergeObj Object
			var mergeErr error
			if mergeValue == "" {
				if p.lineNum < len(p.lines) {
					nextLine := p.lines[p.lineNum]
					nextIndent := p.getIndent(nextLine)
					if nextIndent > currentIndent {
						mergeObj, mergeErr = p.parseValue(nextIndent)
						if mergeErr != nil {
							return nil, mergeErr
						}
					}
				}
			} else {
				mergeObj, mergeErr = p.parseInlineValue(mergeValue, currentIndent)
				if mergeErr != nil {
					return nil, mergeErr
				}
			}

			if mergeMap, ok := mergeObj.(*Map); ok {
				for _, pair := range mergeMap.Pairs {
					if _, exists := pairs[pair.Key.HashKey()]; !exists {
						pairs[pair.Key.HashKey()] = pair
					}
				}
			} else if mergeArr, ok := mergeObj.(*Array); ok {
				for _, elem := range mergeArr.Elements {
					if mergeMap, ok := elem.(*Map); ok {
						for _, pair := range mergeMap.Pairs {
							if _, exists := pairs[pair.Key.HashKey()]; !exists {
								pairs[pair.Key.HashKey()] = pair
							}
						}
					}
				}
			}
			continue
		}

		colonIdx := p.findKeyColon(trimmed)
		if colonIdx == -1 {
			break
		}

		p.lineNum++

		key := strings.TrimSpace(trimmed[:colonIdx])
		key = unquoteYAMLKey(key)

		valueStr := strings.TrimSpace(trimmed[colonIdx+1:])

		anchorName := ""
		if strings.HasPrefix(key, "&") {
			endIdx := strings.IndexAny(key[1:], " \t")
			if endIdx != -1 {
				endIdx++
				anchorName = key[1:endIdx]
				key = strings.TrimSpace(key[endIdx+1:])
			}
		}

		valueAnchorName := ""
		if strings.HasPrefix(valueStr, "&") {
			endIdx := strings.IndexAny(valueStr[1:], " \t")
			if endIdx == -1 {
				valueAnchorName = valueStr[1:]
				valueStr = ""
			}
		}

		var value Object
		var err error

		if valueStr == "" {
			if p.lineNum < len(p.lines) {
				nextLine := p.lines[p.lineNum]
				nextIndent := p.getIndent(nextLine)

				if nextIndent > currentIndent {
					value, err = p.parseValue(nextIndent)
					if err != nil {
						return nil, err
					}
				} else {
					value = NULL
				}
			} else {
				value = NULL
			}

			if valueAnchorName != "" {
				p.anchors[valueAnchorName] = value
			}
		} else {
			value, err = p.parseInlineValue(valueStr, currentIndent)
			if err != nil {
				return nil, err
			}
		}

		if anchorName != "" {
			p.anchors[anchorName] = value
		}

		keyObj := NewString(key)
		pairs[keyObj.HashKey()] = MapPair{
			Key:   keyObj,
			Value: value,
		}
	}

	return NewMap(pairs), nil
}

// parseFlowSequence parses an inline sequence like [a, b, c]
// Supports multiline flow sequences
func (p *yamlParser) parseFlowSequence(s string) (Object, error) {
	s = strings.TrimSpace(s)

	if !strings.HasPrefix(s, "[") {
		return nil, p.newYAMLError("invalid flow sequence: %s", s)
	}

	// Check if we need to collect more lines for multiline flow sequence
	if !strings.HasSuffix(s, "]") || !p.isFlowComplete(s, '[', ']') {
		collected := s

		// Collect remaining lines until we find the closing bracket
		for p.lineNum < len(p.lines) {
			nextLine := p.lines[p.lineNum]
			collected += " " + strings.TrimSpace(nextLine)
			p.lineNum++

			if p.isFlowComplete(collected, '[', ']') {
				break
			}
		}

		if !p.isFlowComplete(collected, '[', ']') {
			return nil, p.newYAMLError("invalid flow sequence: %s", collected)
		}
		s = collected
	}

	content := strings.TrimSpace(s[1 : len(s)-1])
	if content == "" {
		return NewArray([]Object{}), nil
	}

	elements := p.parseFlowElements(content)

	result := make([]Object, len(elements))
	for i, elem := range elements {
		var err error
		result[i], err = p.parseInlineValue(elem, 0)
		if err != nil {
			return nil, err
		}
	}

	return NewArray(result), nil
}

// parseFlowMapping parses an inline mapping like {key: value, key2: value2}
// Supports multiline flow mappings
func (p *yamlParser) parseFlowMapping(s string) (Object, error) {
	s = strings.TrimSpace(s)

	if !strings.HasPrefix(s, "{") {
		return nil, p.newYAMLError("invalid flow mapping: %s", s)
	}

	// Check if we need to collect more lines for multiline flow mapping
	if !strings.HasSuffix(s, "}") || !p.isFlowComplete(s, '{', '}') {
		collected := s

		// Collect remaining lines until we find the closing brace
		for p.lineNum < len(p.lines) {
			nextLine := p.lines[p.lineNum]
			collected += " " + strings.TrimSpace(nextLine)
			p.lineNum++

			if p.isFlowComplete(collected, '{', '}') {
				break
			}
		}

		if !p.isFlowComplete(collected, '{', '}') {
			return nil, p.newYAMLError("invalid flow mapping: %s", collected)
		}
		s = collected
	}

	content := strings.TrimSpace(s[1 : len(s)-1])
	if content == "" {
		return NewMap(make(map[HashKey]MapPair)), nil
	}

	pairs := make(map[HashKey]MapPair)
	elements := p.parseFlowElements(content)

	for _, elem := range elements {
		colonIdx := p.findKeyColon(elem)
		if colonIdx == -1 {
			continue
		}

		key := strings.TrimSpace(elem[:colonIdx])
		key = unquoteYAMLKey(key)
		valueStr := strings.TrimSpace(elem[colonIdx+1:])

		value, err := p.parseInlineValue(valueStr, 0)
		if err != nil {
			return nil, err
		}

		keyObj := NewString(key)
		pairs[keyObj.HashKey()] = MapPair{
			Key:   keyObj,
			Value: value,
		}
	}

	return NewMap(pairs), nil
}

// isFlowComplete checks if a flow structure is complete (balanced brackets)
func (p *yamlParser) isFlowComplete(s string, openChar, closeChar rune) bool {
	depth := 0
	inString := false
	stringChar := rune(0)

	for i, ch := range s {
		switch {
		case inString:
			if ch == stringChar && (i == 0 || s[i-1] != '\\') {
				inString = false
			}
		case ch == '"' || ch == '\'':
			inString = true
			stringChar = ch
		case ch == '[' || ch == '{':
			depth++
		case ch == ']' || ch == '}':
			depth--
		}
	}

	return depth == 0
}

// parseFlowElements splits flow content by commas, respecting nested structures
func (p *yamlParser) parseFlowElements(content string) []string {
	var elements []string
	var current strings.Builder
	depth := 0
	inString := false
	stringChar := rune(0)

	for i, ch := range content {
		switch {
		case inString:
			current.WriteRune(ch)
			if ch == stringChar && (i == 0 || content[i-1] != '\\') {
				inString = false
			}
		case ch == '"' || ch == '\'':
			inString = true
			stringChar = ch
			current.WriteRune(ch)
		case ch == '[' || ch == '{':
			depth++
			current.WriteRune(ch)
		case ch == ']' || ch == '}':
			depth--
			current.WriteRune(ch)
		case ch == ',' && depth == 0:
			elements = append(elements, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		elements = append(elements, strings.TrimSpace(current.String()))
	}

	return elements
}

// parseBlockScalar parses a block scalar (literal | or folded >)
func (p *yamlParser) parseBlockScalar(indicator string, baseIndent int) (Object, error) {
	indicator = strings.TrimSpace(indicator)

	chomping := "clip"
	if strings.Contains(indicator, "+") {
		chomping = "keep"
	} else if strings.Contains(indicator, "-") {
		chomping = "strip"
	}

	isFolded := strings.HasPrefix(indicator, ">")

	// Check for explicit indent indicator (e.g., |2, >4)
	explicitIndent := 0
	indicator = strings.TrimPrefix(indicator, "|")
	indicator = strings.TrimPrefix(indicator, ">")
	indicator = strings.TrimSuffix(indicator, "+")
	indicator = strings.TrimSuffix(indicator, "-")
	indicator = strings.TrimSpace(indicator)

	if len(indicator) > 0 {
		if indent, err := strconv.Atoi(indicator); err == nil {
			explicitIndent = indent
		}
	}

	// Determine the indentation for the block content
	blockIndent := explicitIndent
	if blockIndent == 0 {
		blockIndent = -1
	}
	var lines []string

	for p.lineNum < len(p.lines) {
		line := p.lines[p.lineNum]

		if strings.TrimSpace(line) == "" {
			lines = append(lines, "")
			p.lineNum++
			continue
		}

		currentIndent := p.getIndent(line)

		if blockIndent == -1 {
			if currentIndent <= baseIndent {
				break
			}
			blockIndent = currentIndent
		}

		if currentIndent < blockIndent {
			break
		}

		if currentIndent >= blockIndent {
			lines = append(lines, line[blockIndent:])
		} else {
			lines = append(lines, strings.TrimLeft(line, " \t"))
		}
		p.lineNum++
	}

	var result string
	if isFolded {
		result = p.processFoldedLines(lines, chomping)
	} else {
		result = p.processLiteralLines(lines, chomping)
	}

	return NewString(result), nil
}

// processLiteralLines processes lines for literal block style (|)
func (p *yamlParser) processLiteralLines(lines []string, chomping string) string {
	if len(lines) == 0 {
		return ""
	}

	result := strings.Join(lines, "\n")

	switch chomping {
	case "keep":
		result += "\n"
	case "strip":
		result = strings.TrimRight(result, "\n")
	default:
		result = strings.TrimRight(result, "\n") + "\n"
	}

	return result
}

// processFoldedLines processes lines for folded block style (>)
func (p *yamlParser) processFoldedLines(lines []string, chomping string) string {
	if len(lines) == 0 {
		return ""
	}

	var result strings.Builder
	var paragraph strings.Builder

	for i, line := range lines {
		if line == "" {
			if paragraph.Len() > 0 {
				result.WriteString(paragraph.String())
				result.WriteString("\n")
				paragraph.Reset()
			}
			result.WriteString("\n")
		} else if i > 0 && lines[i-1] != "" {
			paragraph.WriteString(" ")
			paragraph.WriteString(line)
		} else {
			paragraph.WriteString(line)
		}
	}

	if paragraph.Len() > 0 {
		result.WriteString(paragraph.String())
	}

	resultStr := result.String()

	switch chomping {
	case "keep":
		resultStr = strings.TrimRight(resultStr, "\n") + "\n"
	case "strip":
		resultStr = strings.TrimRight(resultStr, "\n")
	default:
		resultStr = strings.TrimRight(resultStr, "\n") + "\n"
	}

	return resultStr
}

// parseScalar parses a scalar value with type detection
func (p *yamlParser) parseScalar(s string) (Object, error) {
	s = strings.TrimSpace(s)

	if s == "" {
		return NULL, nil
	}

	// Strip comments from scalar values (but not from quoted strings)
	s = p.stripComment(s)

	if unquoted, err := p.unquoteYAMLStringChecked(s); err != nil {
		return nil, err
	} else if unquoted != s {
		// If it was a quoted string, always return as string (don't parse further)
		return NewString(unquoted), nil
	}

	if isYAMLNull(s) {
		return NULL, nil
	}

	if b, ok := parseYAMLBool(s); ok {
		return b, nil
	}

	if f, ok := parseYAMLSpecialFloat(s); ok {
		return NewFloat(f), nil
	}

	// Check for timestamp before sexagesimal (timestamps contain colons too)
	if ts, ok := parseYAMLTimestamp(s); ok {
		return ts, nil
	}

	if i, ok := parseYAMLInt(s); ok {
		return NewInt(i), nil
	}

	if f, ok := parseYAMLFloat(s); ok {
		return NewFloat(f), nil
	}

	// Try sexagesimal number (e.g., 60:00:00)
	if f, ok := parseYAMLSexagesimal(s); ok {
		return NewFloat(f), nil
	}

	return NewString(s), nil
}

// stripComment removes inline comments from a scalar value
// Comments are not stripped from quoted strings
func (p *yamlParser) stripComment(s string) string {
	// If the string is quoted, don't strip comments
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s
		}
	}

	// Find comment position, respecting strings
	inString := false
	stringChar := byte(0)
	depth := 0

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case inString:
			if ch == stringChar && i > 0 && s[i-1] != '\\' {
				inString = false
			}
		case ch == '"' || ch == '\'':
			inString = true
			stringChar = ch
		case ch == '[' || ch == '{':
			depth++
		case ch == ']' || ch == '}':
			depth--
		case ch == '#' && depth == 0 && !inString:
			// Check if there's whitespace before the #
			if i == 0 || s[i-1] == ' ' || s[i-1] == '\t' {
				return strings.TrimSpace(s[:i])
			}
		}
	}

	return s
}

// Helper functions

// findKeyColon finds the colon that separates key from value
func (p *yamlParser) findKeyColon(s string) int {
	inString := false
	stringChar := rune(0)
	depth := 0

	for i, ch := range s {
		switch {
		case inString:
			if ch == stringChar && (i == 0 || s[i-1] != '\\') {
				inString = false
			}
		case ch == '"' || ch == '\'':
			inString = true
			stringChar = ch
		case ch == '[' || ch == '{':
			depth++
		case ch == ']' || ch == '}':
			depth--
		case ch == ':' && depth == 0:
			if i+1 >= len(s) || s[i+1] == ' ' || s[i+1] == '\t' || s[i+1] == '\n' {
				return i
			}
		}
	}

	return -1
}

// isNewMappingKey checks if a line starts a new mapping key
func (p *yamlParser) isNewMappingKey(s string) bool {
	colonIdx := p.findKeyColon(s)
	if colonIdx == -1 {
		return false
	}

	depth := 0
	for i, ch := range s {
		if ch == '[' || ch == '{' {
			depth++
		} else if ch == ']' || ch == '}' {
			depth--
		}
		if i == colonIdx && depth == 0 {
			return true
		}
	}

	return false
}

// skipEmptyCommentsAndDirectives skips empty lines, comments, and YAML directives
func (p *yamlParser) skipEmptyCommentsAndDirectives() {
	for p.lineNum < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.lineNum])
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "%YAML") || strings.HasPrefix(line, "%TAG") {
			p.lineNum++
			continue
		}
		break
	}
}

// getIndent returns the indentation level of a line
func (p *yamlParser) getIndent(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 2
		} else {
			break
		}
	}
	return count
}

// yamlKVCollector helps collect key-value pairs from a line
type yamlKVCollector struct {
	line string
	pos  int
}

func newYAMLKVCollector(line string) *yamlKVCollector {
	return &yamlKVCollector{line: line}
}

func (c *yamlKVCollector) next() (key, value string, err error) {
	for c.pos < len(c.line) && (c.line[c.pos] == ' ' || c.line[c.pos] == '\t') {
		c.pos++
	}

	if c.pos >= len(c.line) {
		return "", "", nil
	}

	inString := false
	stringChar := byte(0)
	keyStart := c.pos

	for c.pos < len(c.line) {
		ch := c.line[c.pos]

		if inString {
			if ch == stringChar && c.pos > 0 && c.line[c.pos-1] != '\\' {
				inString = false
			}
		} else if ch == '"' || ch == '\'' {
			inString = true
			stringChar = ch
		} else if ch == ':' {
			key = strings.TrimSpace(c.line[keyStart:c.pos])
			c.pos++

			for c.pos < len(c.line) && (c.line[c.pos] == ' ' || c.line[c.pos] == '\t') {
				c.pos++
			}

			if c.pos < len(c.line) {
				value = c.line[c.pos:]
				c.pos = len(c.line)
			}

			return key, value, nil
		}

		c.pos++
	}

	return "", "", fmt.Errorf("no key-value pair found")
}

// unquoteYAMLKey removes quotes from a YAML key
func unquoteYAMLKey(s string) string {
	return unquoteYAMLString(s)
}

// unquoteYAMLString removes surrounding quotes from a YAML string
func unquoteYAMLString(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			quote := s[0]
			result := s[1 : len(s)-1]
			if quote == '"' {
				result = unescapeYAMLString(result)
			}
			return result
		}
	}
	return s
}

// unquoteYAMLStringChecked removes surrounding quotes from a YAML string with validation
// Returns an error if the string has unclosed quotes
// Supports multiline strings with backslash continuation
func (p *yamlParser) unquoteYAMLStringChecked(s string) (string, error) {
	if len(s) >= 2 {
		if s[0] == '"' {
			// Handle multiline double-quoted strings with backslash continuation
			if s[len(s)-1] != '"' {
				// Check if this is a multiline string with continuation (ends with \)
				// or just an unclosed quote
				collected := s
				// Note: p.lineNum was already incremented by the caller, so it points to the next line
				// We need to read from the current lineNum position
				for p.lineNum < len(p.lines) {
					// Check if current collected string ends with backslash (continuation)
					if strings.HasSuffix(collected, "\\") {
						// Read the next line (lineNum already points to it)
						nextLine := p.lines[p.lineNum]
						trimmedNext := strings.TrimLeft(nextLine, " \t")
						// Remove the backslash and join with next line
						collected = collected[:len(collected)-1] + trimmedNext
						p.lineNum++ // Move to next line for potential further continuation
						// Check if we now have a closing quote
						if strings.HasSuffix(collected, "\"") {
							// Make sure it's not an escaped quote
							if len(collected) >= 2 && collected[len(collected)-2] != '\\' {
								break
							}
						}
					} else {
						// Not a continuation, just read next line
						nextLine := p.lines[p.lineNum]
						trimmedNext := strings.TrimLeft(nextLine, " \t")
						collected += "\n" + trimmedNext
						p.lineNum++
						// Check if we now have a closing quote
						if strings.HasSuffix(trimmedNext, "\"") {
							break
						}
					}
				}
				s = collected
			}

			if !strings.HasSuffix(s, "\"") || len(s) < 2 {
				return "", p.newYAMLError("unclosed double-quoted string")
			}
			result := s[1 : len(s)-1]
			return unescapeYAMLString(result), nil
		}
		if s[0] == '\'' {
			// Handle multiline single-quoted strings
			if s[len(s)-1] != '\'' {
				collected := s
				// Note: p.lineNum was already incremented by the caller
				for p.lineNum < len(p.lines) {
					nextLine := p.lines[p.lineNum]
					trimmedNext := strings.TrimLeft(nextLine, " \t")
					collected += "\n" + trimmedNext
					p.lineNum++

					// Check for closing quote (must not be '' which is escaped quote)
					// A closing quote is ' that is not followed by another '
					if strings.HasSuffix(trimmedNext, "'") {
						// Count trailing quotes
						count := 0
						for i := len(trimmedNext) - 1; i >= 0 && trimmedNext[i] == '\''; i-- {
							count++
						}
						// Odd number of quotes means it's closed
						if count%2 == 1 {
							break
						}
					}
				}
				s = collected
			}

			if !strings.HasSuffix(s, "'") || len(s) < 2 {
				return "", p.newYAMLError("unclosed single-quoted string")
			}
			result := s[1 : len(s)-1]
			// Single-quoted strings only escape '' to '
			return strings.ReplaceAll(result, "''", "'"), nil
		}
	}
	return s, nil
}

// unescapeYAMLString handles escape sequences in double-quoted YAML strings
func unescapeYAMLString(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				result.WriteByte('\n')
				i += 2
			case 'r':
				result.WriteByte('\r')
				i += 2
			case 't':
				result.WriteByte('\t')
				i += 2
			case '\\':
				result.WriteByte('\\')
				i += 2
			case '"':
				result.WriteByte('"')
				i += 2
			case '/':
				result.WriteByte('/')
				i += 2
			case 'b':
				result.WriteByte('\b')
				i += 2
			case 'f':
				result.WriteByte('\f')
				i += 2
			case '0':
				if i+2 < len(s) && s[i+2] >= '0' && s[i+2] <= '7' {
					val := int(s[i+2] - '0')
					if i+3 < len(s) && s[i+3] >= '0' && s[i+3] <= '7' {
						val = val*8 + int(s[i+3]-'0')
						result.WriteByte(byte(val))
						i += 4
						continue
					}
					result.WriteByte(byte(val))
					i += 3
					continue
				}
				result.WriteByte(0)
				i += 2
			case 'x':
				if i+3 < len(s) {
					hex := s[i+2 : i+4]
					if val, err := strconv.ParseInt(hex, 16, 32); err == nil {
						result.WriteByte(byte(val))
						i += 4
						continue
					}
				}
				result.WriteByte(s[i])
				i++
			case 'u':
				if i+5 < len(s) {
					hex := s[i+2 : i+6]
					if val, err := strconv.ParseInt(hex, 16, 32); err == nil {
						result.WriteRune(rune(val))
						i += 6
						continue
					}
				}
				result.WriteByte(s[i])
				i++
			case 'U':
				if i+9 < len(s) {
					hex := s[i+2 : i+10]
					if val, err := strconv.ParseInt(hex, 16, 32); err == nil {
						result.WriteRune(rune(val))
						i += 10
						continue
					}
				}
				result.WriteByte(s[i])
				i++
			case ' ', '\t':
				// Escaped space/tab in YAML (line continuation in double-quoted strings)
				i += 2
			case '\n':
				// Escaped newline - continue on next line
				i += 2
				for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
					i++
				}
			// YAML 1.2 additional escape sequences
			case 'L', 'l':
				// Line separator (U+2028)
				result.WriteRune('\u2028')
				i += 2
			case 'N':
				// Next line character (U+0085)
				result.WriteRune('\u0085')
				i += 2
			case 'P', 'p':
				// Paragraph separator (U+2029)
				result.WriteRune('\u2029')
				i += 2
			case '_':
				// Non-breaking space (U+00A0)
				result.WriteRune('\u00A0')
				i += 2
			case 'Z', 'z':
				// No-break space - same as _ (U+00A0)
				result.WriteRune('\u00A0')
				i += 2
			default:
				result.WriteByte(s[i+1])
				i += 2
			}
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}

// isYAMLNull checks if a string represents a null value
func isYAMLNull(s string) bool {
	s = strings.ToLower(s)
	return s == "null" || s == "~" || s == ""
}

// parseYAMLBool parses a YAML boolean
func parseYAMLBool(s string) (*Bool, bool) {
	lower := strings.ToLower(s)
	switch lower {
	case "true", "yes", "on":
		return TRUE, true
	case "false", "no", "off":
		return FALSE, true
	default:
		return FALSE, false
	}
}

// parseYAMLSpecialFloat parses special float values
func parseYAMLSpecialFloat(s string) (float64, bool) {
	lower := strings.ToLower(s)
	switch lower {
	case ".inf", "+.inf":
		return math.Inf(1), true
	case "-.inf":
		return math.Inf(-1), true
	case ".nan":
		return math.NaN(), true
	default:
		return 0, false
	}
}

// parseYAMLInt parses a YAML integer
func parseYAMLInt(s string) (int64, bool) {
	// Remove underscores for YAML 1.2 numeric format
	s = strings.ReplaceAll(s, "_", "")

	// Handle sign
	sign := int64(1)
	if strings.HasPrefix(s, "+") {
		s = s[1:]
	} else if strings.HasPrefix(s, "-") {
		sign = -1
		s = s[1:]
	}

	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return sign * i, true
	}

	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		if i, err := strconv.ParseInt(s[2:], 16, 64); err == nil {
			return sign * i, true
		}
	}

	if strings.HasPrefix(s, "0o") || strings.HasPrefix(s, "0O") {
		if i, err := strconv.ParseInt(s[2:], 8, 64); err == nil {
			return sign * i, true
		}
	}

	if strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B") {
		if i, err := strconv.ParseInt(s[2:], 2, 64); err == nil {
			return sign * i, true
		}
	}

	if strings.HasPrefix(s, "0") && len(s) > 1 {
		if i, err := strconv.ParseInt(s, 8, 64); err == nil {
			return sign * i, true
		}
	}

	return 0, false
}

// parseYAMLTimestamp parses YAML timestamp values
// Supports formats: YYYY-MM-DD, YYYY-MM-DD HH:MM:SS, YYYY-MM-DDTHH:MM:SS, RFC3339
func parseYAMLTimestamp(s string) (Object, bool) {
	// Common timestamp formats
	formats := []string{
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05.999999999Z",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	// Quick check: must start with a digit and look like a date
	if len(s) < 10 {
		return nil, false
	}
	if s[0] < '0' || s[0] > '9' {
		return nil, false
	}
	// Must have a dash in position 4 (YYYY-MM)
	if len(s) < 5 || s[4] != '-' {
		return nil, false
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			// Return as string representation for now
			// Could return a Time object if Xxlang has one
			return NewString(t.Format("2006-01-02T15:04:05Z07:00")), true
		}
	}

	return nil, false
}

// parseYAMLFloat parses a YAML float
func parseYAMLFloat(s string) (float64, bool) {
	// Remove underscores for YAML 1.2 numeric format
	s = strings.ReplaceAll(s, "_", "")
	f, err := strconv.ParseFloat(s, 64)
	return f, err == nil
}

// parseYAMLSexagesimal parses sexagesimal (base 60) numbers like 60:00:00
func parseYAMLSexagesimal(s string) (float64, bool) {
	if !strings.Contains(s, ":") {
		return 0, false
	}

	parts := strings.Split(s, ":")
	if len(parts) < 2 {
		return 0, false
	}

	var result float64
	for _, part := range parts {
		var val float64
		if _, err := fmt.Sscanf(part, "%f", &val); err != nil {
			return 0, false
		}
		result = result*60 + val
	}

	return result, true
}

// yamlDeepCopyObject creates a deep copy of an object for YAML alias resolution
func yamlDeepCopyObject(obj Object) Object {
	if obj == nil {
		return NULL
	}

	switch o := obj.(type) {
	case *Null:
		return NULL
	case *Bool:
		if o.Value {
			return TRUE
		}
		return FALSE
	case *Int:
		return NewInt(o.Value)
	case *Float:
		return NewFloat(o.Value)
	case *String:
		return NewString(o.Value)
	case *Array:
		elements := make([]Object, len(o.Elements))
		for i, elem := range o.Elements {
			elements[i] = yamlDeepCopyObject(elem)
		}
		return NewArray(elements)
	case *Map:
		pairs := make(map[HashKey]MapPair)
		for k, v := range o.Pairs {
			pairs[k] = MapPair{
				Key:   yamlDeepCopyObject(v.Key),
				Value: yamlDeepCopyObject(v.Value),
			}
		}
		return NewMap(pairs)
	default:
		return obj
	}
}

// ============================================================
// Serialization functions
// ============================================================

// SerializeYAML converts an Xxlang Object to a YAML string.
func SerializeYAML(obj Object, indent int) string {
	return serializeYAMLValue(obj, 0, indent)
}

// serializeYAMLValue recursively serializes a value to YAML
func serializeYAMLValue(obj Object, level, indent int) string {
	if obj == nil || obj == NULL {
		return "null"
	}

	switch o := obj.(type) {
	case *Null:
		return "null"
	case *Bool:
		if o.Value {
			return "true"
		}
		return "false"
	case *Int:
		return strconv.FormatInt(o.Value, 10)
	case *Float:
		return serializeYAMLFloat(o.Value)
	case *String:
		return serializeYAMLString(o.Value)
	case *Array:
		return serializeYAMLArray(o.Elements, level, indent)
	case *Map:
		return serializeYAMLMap(o.Pairs, level, indent)
	case *OrderedMap:
		return serializeYAMLOrderedMap(o, level, indent)
	default:
		return serializeYAMLString(o.Inspect())
	}
}

// serializeYAMLFloat formats a float for YAML
func serializeYAMLFloat(f float64) string {
	if math.IsNaN(f) {
		return ".nan"
	}
	if math.IsInf(f, 1) {
		return ".inf"
	}
	if math.IsInf(f, -1) {
		return "-.inf"
	}

	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.Contains(s, ".") && !strings.Contains(s, "e") && !strings.Contains(s, "E") {
		s += ".0"
	}
	return s
}

// serializeYAMLString formats a string for YAML
func serializeYAMLString(s string) string {
	if s == "" {
		return `""`
	}

	lower := strings.ToLower(s)
	needsQuotes := lower == "null" || lower == "true" || lower == "false" || lower == "yes" || lower == "no" || lower == "~"

	if _, err := strconv.ParseFloat(s, 64); err == nil {
		needsQuotes = true
	}

	for _, ch := range s {
		if ch == ':' || ch == '{' || ch == '}' || ch == '[' || ch == ']' || ch == ',' || ch == '&' || ch == '*' || ch == '?' || ch == '|' || ch == '-' || ch == '<' || ch == '>' || ch == '=' || ch == '!' || ch == '%' || ch == '@' || ch == '`' {
			needsQuotes = true
			break
		}
	}

	if len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		needsQuotes = true
	}

	if strings.Contains(s, "\n") {
		return "|\n" + serializeYAMLBlock(s, 0, 2)
	}

	if needsQuotes {
		return quoteYAMLString(s)
	}

	return s
}

// quoteYAMLString quotes a string for YAML
func quoteYAMLString(s string) string {
	var result strings.Builder
	result.WriteByte('"')
	for _, ch := range s {
		switch ch {
		case '"':
			result.WriteString(`\"`)
		case '\\':
			result.WriteString(`\\`)
		case '\n':
			result.WriteString(`\n`)
		case '\r':
			result.WriteString(`\r`)
		case '\t':
			result.WriteString(`\t`)
		case '\b':
			result.WriteString(`\b`)
		case '\f':
			result.WriteString(`\f`)
		default:
			if ch < 32 || ch > 126 {
				result.WriteString(fmt.Sprintf("\\u%04x", ch))
			} else {
				result.WriteRune(ch)
			}
		}
	}
	result.WriteByte('"')
	return result.String()
}

// serializeYAMLBlock formats a multiline string in literal block style
func serializeYAMLBlock(s string, level, indent int) string {
	indentStr := strings.Repeat(" ", (level+1)*indent)
	lines := strings.Split(s, "\n")
	var result strings.Builder
	for i, line := range lines {
		if i > 0 {
			result.WriteByte('\n')
		}
		result.WriteString(indentStr)
		result.WriteString(line)
	}
	return result.String()
}

// serializeYAMLArray serializes an array to YAML
func serializeYAMLArray(elements []Object, level, indent int) string {
	if len(elements) == 0 {
		return "[]"
	}

	indentStr := strings.Repeat(" ", level*indent)
	var result strings.Builder

	for i, elem := range elements {
		if i > 0 {
			result.WriteByte('\n')
		}
		result.WriteString(indentStr)
		result.WriteString("- ")

		elemYAML := serializeYAMLValue(elem, level+1, indent)

		if isArrayOrMap(elem) {
			lines := strings.Split(elemYAML, "\n")
			if len(lines) > 0 {
				result.WriteString(lines[0])
				for j := 1; j < len(lines); j++ {
					result.WriteByte('\n')
					result.WriteString(strings.Repeat(" ", (level+1)*indent))
					result.WriteString(lines[j])
				}
			}
		} else {
			result.WriteString(elemYAML)
		}
	}

	return result.String()
}

// serializeYAMLMap serializes a map to YAML
func serializeYAMLMap(pairs map[HashKey]MapPair, level, indent int) string {
	if len(pairs) == 0 {
		return "{}"
	}

	indentStr := strings.Repeat(" ", level*indent)
	var result strings.Builder

	first := true
	for _, pair := range pairs {
		if !first {
			result.WriteByte('\n')
		}
		first = false

		result.WriteString(indentStr)

		if keyStr, ok := pair.Key.(*String); ok {
			result.WriteString(serializeYAMLString(keyStr.Value))
		} else {
			result.WriteString(serializeYAMLString(pair.Key.Inspect()))
		}
		result.WriteString(": ")

		elemYAML := serializeYAMLValue(pair.Value, level+1, indent)

		if isArrayOrMap(pair.Value) {
			lines := strings.Split(elemYAML, "\n")
			if len(lines) > 0 {
				if len(lines) == 1 && (lines[0] == "[]" || lines[0] == "{}") {
					result.WriteString(lines[0])
				} else {
					result.WriteByte('\n')
					result.WriteString(strings.Repeat(" ", (level+1)*indent))
					result.WriteString(lines[0])
					for j := 1; j < len(lines); j++ {
						result.WriteByte('\n')
						result.WriteString(strings.Repeat(" ", (level+1)*indent))
						result.WriteString(lines[j])
					}
				}
			}
		} else {
			result.WriteString(elemYAML)
		}
	}

	return result.String()
}

// serializeYAMLOrderedMap serializes an ordered map to YAML
func serializeYAMLOrderedMap(om *OrderedMap, level, indent int) string {
	if len(om.orderSlice) == 0 {
		return "{}"
	}

	indentStr := strings.Repeat(" ", level*indent)
	var result strings.Builder

	for i, pair := range om.orderSlice {
		if i > 0 {
			result.WriteByte('\n')
		}

		result.WriteString(indentStr)

		if keyStr, ok := pair.Key.(*String); ok {
			result.WriteString(serializeYAMLString(keyStr.Value))
		} else {
			result.WriteString(serializeYAMLString(pair.Key.Inspect()))
		}
		result.WriteString(": ")

		elemYAML := serializeYAMLValue(pair.Value, level+1, indent)

		if isArrayOrMap(pair.Value) {
			lines := strings.Split(elemYAML, "\n")
			if len(lines) > 0 {
				if len(lines) == 1 && (lines[0] == "[]" || lines[0] == "{}") {
					result.WriteString(lines[0])
				} else {
					result.WriteByte('\n')
					result.WriteString(strings.Repeat(" ", (level+1)*indent))
					result.WriteString(lines[0])
					for j := 1; j < len(lines); j++ {
						result.WriteByte('\n')
						result.WriteString(strings.Repeat(" ", (level+1)*indent))
						result.WriteString(lines[j])
					}
				}
			}
		} else {
			result.WriteString(elemYAML)
		}
	}

	return result.String()
}

// isArrayOrMap checks if an object is an array or map
func isArrayOrMap(obj Object) bool {
	if obj == nil {
		return false
	}
	switch obj.(type) {
	case *Array, *Map, *OrderedMap:
		return true
	default:
		return false
	}
}

// ============================================================
// Additional utility functions for YAML module
// ============================================================

// YAMLPathQuery queries a YAML object using a simple path syntax
func YAMLPathQuery(obj Object, path string) Object {
	if path == "" {
		return obj
	}

	parts := strings.Split(path, ".")
	current := obj

	for _, part := range parts {
		if current == nil || current == NULL {
			return NULL
		}

		// Handle array index
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			arr, ok := current.(*Array)
			if !ok {
				return NULL
			}
			idxStr := part[1 : len(part)-1]
			idx, err := strconv.Atoi(idxStr)
			if err != nil || idx < 0 || idx >= len(arr.Elements) {
				return NULL
			}
			current = arr.Elements[idx]
			continue
		}

		// Handle map key
		m, ok := current.(*Map)
		if !ok {
			return NULL
		}

		key := NewString(part)
		pair, exists := m.Pairs[key.HashKey()]
		if !exists {
			return NULL
		}
		current = pair.Value
	}

	return current
}

// YAMLPathSet sets a value in a YAML object using a path
func YAMLPathSet(obj Object, path string, value Object) Object {
	if path == "" {
		return value
	}

	parts := strings.Split(path, ".")
	return yamlPathSetRecursive(obj, parts, 0, value)
}

func yamlPathSetRecursive(obj Object, parts []string, index int, value Object) Object {
	if index >= len(parts) {
		return value
	}

	part := parts[index]

	if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
		arr, ok := obj.(*Array)
		if !ok {
			return obj
		}

		idxStr := part[1 : len(part)-1]
		idx, err := strconv.Atoi(idxStr)
		if err != nil || idx < 0 || idx >= len(arr.Elements) {
			return obj
		}

		newElements := make([]Object, len(arr.Elements))
		copy(newElements, arr.Elements)
		newElements[idx] = yamlPathSetRecursive(newElements[idx], parts, index+1, value)

		return NewArray(newElements)
	}

	m, ok := obj.(*Map)
	if !ok {
		m = NewMap(make(map[HashKey]MapPair))
	}

	newPairs := make(map[HashKey]MapPair)
	for k, v := range m.Pairs {
		newPairs[k] = v
	}

	key := NewString(part)
	if index == len(parts)-1 {
		newPairs[key.HashKey()] = MapPair{
			Key:   key,
			Value: value,
		}
	} else {
		var existingValue Object = NULL
		if pair, exists := newPairs[key.HashKey()]; exists {
			existingValue = pair.Value
		}
		newPairs[key.HashKey()] = MapPair{
			Key:   key,
			Value: yamlPathSetRecursive(existingValue, parts, index+1, value),
		}
	}

	return NewMap(newPairs)
}

// ValidateYAML validates a YAML string and returns detailed error information
func ValidateYAML(s string) error {
	parser := newYAMLParser(s)
	_, err := parser.parse()
	return err
}

// ExtractYAMLAnchors extracts all anchor definitions from a YAML document
func ExtractYAMLAnchors(s string) map[string]Object {
	parser := newYAMLParser(s)
	parser.parse()
	return parser.anchors
}

// NormalizeYAML normalizes YAML content by parsing and re-serializing
func NormalizeYAML(s string, indent int) (string, error) {
	obj, err := ParseYAML(s)
	if err != nil {
		return "", err
	}
	return SerializeYAML(obj, indent), nil
}

// MergeYAMLMaps merges multiple YAML maps
func MergeYAMLMaps(maps ...*Map) *Map {
	result := make(map[HashKey]MapPair)

	for _, m := range maps {
		if m == nil {
			continue
		}
		for k, v := range m.Pairs {
			result[k] = v
		}
	}

	return NewMap(result)
}

// DeepMergeYAMLMaps performs a deep merge of YAML maps
func DeepMergeYAMLMaps(maps ...*Map) *Map {
	if len(maps) == 0 {
		return NewMap(make(map[HashKey]MapPair))
	}

	result := make(map[HashKey]MapPair)

	for k, v := range maps[0].Pairs {
		result[k] = v
	}

	for i := 1; i < len(maps); i++ {
		if maps[i] == nil {
			continue
		}

		for _, pair := range maps[i].Pairs {
			existing, exists := result[pair.Key.HashKey()]
			if exists {
				existingMap, existingIsMap := existing.Value.(*Map)
				newMap, newIsMap := pair.Value.(*Map)

				if existingIsMap && newIsMap {
					merged := DeepMergeYAMLMaps(existingMap, newMap)
					result[pair.Key.HashKey()] = MapPair{
						Key:   pair.Key,
						Value: merged,
					}
					continue
				}
			}
			result[pair.Key.HashKey()] = pair
		}
	}

	return NewMap(result)
}

// YAMLToJSON converts a YAML string to JSON
func YAMLToJSON(yamlStr string) (string, error) {
	obj, err := ParseYAML(yamlStr)
	if err != nil {
		return "", err
	}

	jsonBytes, err := ObjectToJSON(obj, ObjectToJSONOptions{Indent: true, IndentStr: "  "})
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// JSONToYAML converts a JSON string to YAML
func JSONToYAML(jsonStr string, indent int) (string, error) {
	obj, err := JSONToObject(jsonStr)
	if err != nil {
		return "", err
	}

	return SerializeYAML(obj, indent), nil
}

// SplitYAMLDocuments splits a YAML string into individual documents
func SplitYAMLDocuments(s string) []string {
	var documents []string
	var current bytes.Buffer
	inDocument := false

	lines := strings.Split(s, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			if inDocument && current.Len() > 0 {
				documents = append(documents, current.String())
				current.Reset()
			}
			inDocument = true
			continue
		}

		if trimmed == "..." {
			if current.Len() > 0 {
				documents = append(documents, current.String())
				current.Reset()
			}
			inDocument = false
			continue
		}

		if current.Len() > 0 {
			current.WriteByte('\n')
		}
		current.WriteString(line)
		inDocument = true
	}

	if current.Len() > 0 {
		documents = append(documents, current.String())
	}

	return documents
}

// JoinYAMLDocuments joins multiple YAML documents into a multi-document YAML string
func JoinYAMLDocuments(docs []Object, indent int) string {
	var result bytes.Buffer

	for i, doc := range docs {
		if i > 0 {
			result.WriteString("\n---\n")
		}
		result.WriteString(SerializeYAML(doc, indent))
	}

	return result.String()
}

// YAMLSet wraps an array to represent a YAML set (unique values)
func YAMLSet(elements []Object) *Map {
	pairs := make(map[HashKey]MapPair)
	for _, elem := range elements {
		pairs[elem.HashKey()] = MapPair{
			Key:   elem,
			Value: NULL,
		}
	}
	return NewMap(pairs)
}

// IsYAMLSet checks if a map represents a YAML set (all values are null)
func IsYAMLSet(m *Map) bool {
	for _, pair := range m.Pairs {
		if pair.Value != NULL {
			return false
		}
	}
	return len(m.Pairs) > 0
}

// YAMLOMap creates an ordered map from key-value pairs
func YAMLOMap(pairs []MapPair) *OrderedMap {
	om := NewOrderedMap()
	for _, pair := range pairs {
		om.Set(pair.Key, pair.Value)
	}
	return om
}

// ParseYAMLPairs parses key-value pairs from a YAML mapping
func ParseYAMLPairs(yamlStr string) ([]MapPair, error) {
	obj, err := ParseYAML(yamlStr)
	if err != nil {
		return nil, err
	}

	m, ok := obj.(*Map)
	if !ok {
		return nil, fmt.Errorf("YAML must be a mapping")
	}

	pairs := make([]MapPair, 0, len(m.Pairs))
	for _, pair := range m.Pairs {
		pairs = append(pairs, pair)
	}

	return pairs, nil
}

// DetectYAMLType detects the type of a YAML value without full parsing
var yamlTypeRegex = regexp.MustCompile(`^\s*`)

// DetectYAMLValueType detects the type of a YAML value string
func DetectYAMLValueType(s string) string {
	s = strings.TrimSpace(s)

	if s == "" || strings.ToLower(s) == "null" || s == "~" {
		return "null"
	}

	if strings.ToLower(s) == "true" || strings.ToLower(s) == "false" ||
		strings.ToLower(s) == "yes" || strings.ToLower(s) == "no" ||
		strings.ToLower(s) == "on" || strings.ToLower(s) == "off" {
		return "bool"
	}

	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		return "array"
	}

	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		return "map"
	}

	if strings.HasPrefix(s, "|") || strings.HasPrefix(s, ">") {
		return "block"
	}

	if _, err := strconv.ParseInt(s, 0, 64); err == nil {
		return "int"
	}

	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return "float"
	}

	if strings.ToLower(s) == ".inf" || strings.ToLower(s) == "-.inf" || strings.ToLower(s) == ".nan" {
		return "float"
	}

	if strings.HasPrefix(s, "- ") || s == "-" {
		return "array"
	}

	if strings.Contains(s, ":") {
		return "map"
	}

	return "string"
}
