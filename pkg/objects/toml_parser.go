// pkg/objects/toml_parser.go
// TOML parser - pure Go implementation
package objects

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Parser holds the parser state
type Parser struct {
	lexer  *Lexer
	tokens []Token
	pos    int
	errors []string
}

// ParseToml parses a TOML string and returns a document
func ParseToml(input string) (*TomlDocument, error) {
	lexer := NewLexer(input)
	tokens := lexer.AllTokens()

	// Check for lexer errors
	for _, token := range tokens {
		if token.Type == TokenError {
			return nil, fmt.Errorf("lexer error at line %d, col %d: %s", token.Line, token.Col, token.Value)
		}
	}

	p := &Parser{
		lexer:  lexer,
		tokens: tokens,
		pos:    0,
		errors: []string{},
	}

	doc := p.parse()

	if len(p.errors) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(p.errors, "; "))
	}

	return doc, nil
}

// parse starts parsing the document
func (p *Parser) parse() *TomlDocument {
	doc := NewTomlDocument()
	currentTable := doc.root.Value.(map[string]*TomlValue)
	var currentTablePath []string

	for p.pos < len(p.tokens) {
		token := p.current()

		switch token.Type {
		case TokenEOF:
			return doc
		case TokenNewline, TokenComment:
			p.advance()
			continue
		case TokenLBracket:
			// Check if this is [[ for array of tables
			p.advance()
			if p.current().Type == TokenLBracket {
				// Array of tables: [[section]]
				p.advance()
				tablePath := p.parseTablePath()
				// Expect two closing brackets
				if p.current().Type != TokenRBracket {
					p.addError("expected ']' after table array path")
					return doc
				}
				p.advance()
				if p.current().Type != TokenRBracket {
					p.addError("expected ']]' after table array path")
					return doc
				}
				p.advance()
				currentTablePath = tablePath
				currentTable = p.getOrCreateArrayTable(doc, tablePath)
			} else {
				// Regular table: [section]
				tablePath := p.parseTablePath()
				if p.current().Type != TokenRBracket {
					p.addError("expected ']' after table path")
					return doc
				}
				p.advance()
				currentTablePath = tablePath
				currentTable = p.getOrCreateTable(doc, tablePath)
			}
		case TokenKey:
			// Key-value pair
			keyPath, value := p.parseKeyValue()
			if len(keyPath) > 0 && value != nil {
				// Handle dotted keys by creating nested tables
				p.setDottedKey(currentTable, keyPath, value)
			}
		default:
			p.addError(fmt.Sprintf("unexpected token: %s", token.Type))
			p.advance()
		}
	}

	_ = currentTablePath // track current table path
	return doc
}

// current returns the current token
func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

// advance moves to the next token
func (p *Parser) advance() {
	p.pos++
}

// addError adds an error to the parser
func (p *Parser) addError(msg string) {
	token := p.current()
	p.errors = append(p.errors, fmt.Sprintf("line %d, col %d: %s", token.Line, token.Col, msg))
}

// skipWhitespaceAndComments skips whitespace and comments
func (p *Parser) skipWhitespaceAndComments() {
	for p.current().Type == TokenNewline || p.current().Type == TokenComment {
		p.advance()
	}
}

// parseTablePath parses a table path like [section.subsection]
func (p *Parser) parseTablePath() []string {
	var path []string

	for {
		token := p.current()
		if token.Type == TokenKey {
			path = append(path, token.Value)
			p.advance()
		} else if token.Type == TokenString {
			path = append(path, token.Value)
			p.advance()
		} else if token.Type == TokenLiteralString {
			path = append(path, token.Value)
			p.advance()
		} else {
			break
		}

		// Check for dot separator
		if p.current().Type == TokenDot {
			p.advance()
		} else {
			break
		}
	}

	return path
}

// parseKeyValue parses a key = value pair
func (p *Parser) parseKeyValue() ([]string, *TomlValue) {
	// Parse key
	keyToken := p.current()
	if keyToken.Type != TokenKey && keyToken.Type != TokenString && keyToken.Type != TokenLiteralString {
		p.addError(fmt.Sprintf("expected key, got %s", keyToken.Type))
		return nil, nil
	}

	// Build key path as array of strings
	// This handles dotted keys correctly, including quoted keys with dots
	var keyPath []string
	keyPath = append(keyPath, keyToken.Value)
	p.advance()

	// Parse dot-separated keys
	for p.current().Type == TokenDot {
		p.advance()
		nextKey := p.current()
		if nextKey.Type != TokenKey && nextKey.Type != TokenString && nextKey.Type != TokenLiteralString {
			p.addError("expected key after dot")
			return nil, nil
		}
		keyPath = append(keyPath, nextKey.Value)
		p.advance()
	}

	// Parse equals
	if p.current().Type != TokenEquals {
		p.addError(fmt.Sprintf("expected '=', got %s", p.current().Type))
		return nil, nil
	}
	p.advance()

	// Parse value
	value := p.parseValue()

	return keyPath, value
}

// parseValue parses a TOML value
func (p *Parser) parseValue() *TomlValue {
	token := p.current()

	switch token.Type {
	case TokenString:
		p.advance()
		return &TomlValue{Type: TomlString, Value: token.Value}
	case TokenMultiLineString:
		p.advance()
		return &TomlValue{Type: TomlString, Value: token.Value}
	case TokenLiteralString:
		p.advance()
		return &TomlValue{Type: TomlString, Value: token.Value}
	case TokenMultiLineLiteralString:
		p.advance()
		return &TomlValue{Type: TomlString, Value: token.Value}
	case TokenInteger:
		p.advance()
		return parseInteger(token.Value)
	case TokenFloat:
		p.advance()
		return parseFloat(token.Value)
	case TokenBoolean:
		p.advance()
		return &TomlValue{Type: TomlBoolean, Value: token.Value == "true"}
	case TokenDatetime:
		p.advance()
		return parseDatetime(token.Value)
	case TokenDate:
		p.advance()
		return parseDate(token.Value)
	case TokenTime:
		p.advance()
		return parseTime(token.Value)
	case TokenLBracket:
		return p.parseArray()
	case TokenLBrace:
		return p.parseInlineTable()
	default:
		p.addError(fmt.Sprintf("unexpected value type: %s", token.Type))
		p.advance()
		return &TomlValue{Type: TomlNull}
	}
}

// parseArray parses an array [...]
func (p *Parser) parseArray() *TomlValue {
	p.advance() // skip [

	var elements []*TomlValue

	for {
		p.skipWhitespaceAndComments()

		if p.current().Type == TokenRBracket {
			p.advance()
			break
		}

		if p.current().Type == TokenEOF {
			p.addError("unexpected EOF in array")
			break
		}

		value := p.parseValue()
		if value != nil {
			elements = append(elements, value)
		}

		p.skipWhitespaceAndComments()

		// Check for comma or end
		if p.current().Type == TokenComma {
			p.advance()
			continue
		} else if p.current().Type == TokenRBracket {
			p.advance()
			break
		} else {
			p.addError(fmt.Sprintf("expected ',' or ']' in array, got %s", p.current().Type))
			break
		}
	}

	return &TomlValue{Type: TomlArray, Value: elements}
}

// parseInlineTable parses an inline table {...}
func (p *Parser) parseInlineTable() *TomlValue {
	p.advance() // skip {

	table := make(map[string]*TomlValue)

	for {
		p.skipWhitespaceAndComments()

		if p.current().Type == TokenRBrace {
			p.advance()
			break
		}

		if p.current().Type == TokenEOF {
			p.addError("unexpected EOF in inline table")
			break
		}

		keyPath, value := p.parseKeyValue()
		if len(keyPath) > 0 && value != nil {
			// Inline tables can also have dotted keys
			p.setDottedKey(table, keyPath, value)
		}

		p.skipWhitespaceAndComments()

		// Check for comma or end
		if p.current().Type == TokenComma {
			p.advance()
			continue
		} else if p.current().Type != TokenRBrace {
			p.addError(fmt.Sprintf("expected ',' or '}' in inline table, got %s", p.current().Type))
			break
		}
	}

	return &TomlValue{Type: TomlTable, Value: table}
}

// getOrCreateTable gets or creates a table at the specified path
func (p *Parser) getOrCreateTable(doc *TomlDocument, path []string) map[string]*TomlValue {
	current := doc.root.Value.(map[string]*TomlValue)

	for i, key := range path {
		if val, exists := current[key]; exists {
			if val.Type == TomlTable {
				current = val.Value.(map[string]*TomlValue)
			} else if val.Type == TomlArray {
				// If it's an array, get the last table in the array
				// This allows defining subtables under array tables
				arr := val.Value.([]*TomlValue)
				if len(arr) > 0 && arr[len(arr)-1].Type == TomlTable {
					current = arr[len(arr)-1].Value.(map[string]*TomlValue)
				} else {
					p.addError(fmt.Sprintf("'%s' is an array but doesn't contain tables", key))
					return current
				}
			} else {
				// Not a table or array, create error
				p.addError(fmt.Sprintf("'%s' is not a table", key))
				return current
			}
		} else {
			// Create new table
			newTable := &TomlValue{
				Type:  TomlTable,
				Value: make(map[string]*TomlValue),
			}
			current[key] = newTable
			current = newTable.Value.(map[string]*TomlValue)
		}

		_ = i // avoid unused variable warning
	}

	return current
}

// getOrCreateArrayTable gets or creates a table in an array of tables
func (p *Parser) getOrCreateArrayTable(doc *TomlDocument, path []string) map[string]*TomlValue {
	current := doc.root.Value.(map[string]*TomlValue)

	for i, key := range path {
		isLast := i == len(path)-1

		if val, exists := current[key]; exists {
			if isLast {
				// Last key should be an array
				if val.Type != TomlArray {
					p.addError(fmt.Sprintf("'%s' is not an array of tables", key))
					return current
				}
				// Add new table to the array
				arr := val.Value.([]*TomlValue)
				newTable := &TomlValue{
					Type:  TomlTable,
					Value: make(map[string]*TomlValue),
				}
				arr = append(arr, newTable)
				val.Value = arr
				return newTable.Value.(map[string]*TomlValue)
			} else {
				if val.Type == TomlArray {
					// Get the last table in the array
					arr := val.Value.([]*TomlValue)
					if len(arr) > 0 {
						current = arr[len(arr)-1].Value.(map[string]*TomlValue)
					} else {
						p.addError(fmt.Sprintf("array '%s' is empty", key))
						return current
					}
				} else if val.Type == TomlTable {
					current = val.Value.(map[string]*TomlValue)
				} else {
					p.addError(fmt.Sprintf("'%s' is not a table or array", key))
					return current
				}
			}
		} else {
			if isLast {
				// Create new array with one table
				newTable := &TomlValue{
					Type:  TomlTable,
					Value: make(map[string]*TomlValue),
				}
				newArray := &TomlValue{
					Type:  TomlArray,
					Value: []*TomlValue{newTable},
				}
				current[key] = newArray
				return newTable.Value.(map[string]*TomlValue)
			} else {
				// Create new table
				newTable := &TomlValue{
					Type:  TomlTable,
					Value: make(map[string]*TomlValue),
				}
				current[key] = newTable
				current = newTable.Value.(map[string]*TomlValue)
			}
		}
	}

	return current
}

// setDottedKey sets a value at a key path, creating nested tables as needed
func (p *Parser) setDottedKey(table map[string]*TomlValue, keyPath []string, value *TomlValue) {
	if len(keyPath) == 0 {
		p.addError("empty key path")
		return
	}

	if len(keyPath) == 1 {
		// Simple key, set directly
		table[keyPath[0]] = value
		return
	}

	// Navigate/create nested tables for all but the last part
	current := table
	for i := 0; i < len(keyPath)-1; i++ {
		part := keyPath[i]
		if val, exists := current[part]; exists {
			if val.Type != TomlTable {
				// Not a table, create error
				p.addError(fmt.Sprintf("'%s' is not a table", part))
				return
			}
			current = val.Value.(map[string]*TomlValue)
		} else {
			// Create new table
			newTable := &TomlValue{
				Type:  TomlTable,
				Value: make(map[string]*TomlValue),
			}
			current[part] = newTable
			current = newTable.Value.(map[string]*TomlValue)
		}
	}

	// Set the value at the last part
	current[keyPath[len(keyPath)-1]] = value
}

// ============================================================
// Value parsing helpers
// ============================================================

func parseInteger(s string) *TomlValue {
	var val int64
	var err error
	var base int = 10

	// Handle different bases
	if len(s) > 2 {
		switch s[:2] {
		case "0x", "0X":
			base = 16
			s = s[2:]
		case "0o", "0O":
			base = 8
			s = s[2:]
		case "0b", "0B":
			base = 2
			s = s[2:]
		}
	}

	// Remove underscores
	s = strings.ReplaceAll(s, "_", "")

	// Handle sign
	sign := int64(1)
	if len(s) > 0 && s[0] == '-' {
		sign = -1
		s = s[1:]
	} else if len(s) > 0 && s[0] == '+' {
		s = s[1:]
	}

	val, err = strconv.ParseInt(s, base, 64)
	if err != nil {
		return &TomlValue{Type: TomlInteger, Value: int64(0)}
	}

	return &TomlValue{Type: TomlInteger, Value: sign * val}
}

func parseFloat(s string) *TomlValue {
	// Remove underscores
	s = strings.ReplaceAll(s, "_", "")

	// Handle special values
	lower := strings.ToLower(s)
	if lower == "inf" || lower == "+inf" || lower == "infinity" || lower == "+infinity" {
		return &TomlValue{Type: TomlFloat, Value: math.Inf(1)}
	}
	if lower == "-inf" || lower == "-infinity" {
		return &TomlValue{Type: TomlFloat, Value: math.Inf(-1)}
	}
	if lower == "nan" || lower == "+nan" || lower == "-nan" {
		return &TomlValue{Type: TomlFloat, Value: math.NaN()}
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return &TomlValue{Type: TomlFloat, Value: float64(0)}
	}

	return &TomlValue{Type: TomlFloat, Value: val}
}

func parseDatetime(s string) *TomlValue {
	// Try various datetime formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05Z",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return &TomlValue{Type: TomlDatetime, Value: t}
		}
	}

	// Return zero time if parsing fails
	return &TomlValue{Type: TomlDatetime, Value: time.Time{}}
}

func parseDate(s string) *TomlValue {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return &TomlValue{Type: TomlDate, Value: time.Time{}}
	}
	return &TomlValue{Type: TomlDate, Value: t}
}

func parseTime(s string) *TomlValue {
	// Parse time (return as time.Time with zero date)
	t, err := time.Parse("15:04:05", s)
	if err != nil {
		// Try with nanoseconds
		t, err = time.Parse("15:04:05.999999999", s)
		if err != nil {
			return &TomlValue{Type: TomlTime, Value: time.Time{}}
		}
	}
	return &TomlValue{Type: TomlTime, Value: t}
}

// ============================================================
// TOML validation
// ============================================================

// ValidateToml validates a TOML string
func ValidateToml(input string) (bool, string) {
	doc, err := ParseToml(input)
	if err != nil {
		return false, err.Error()
	}
	_ = doc
	return true, ""
}
