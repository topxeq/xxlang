// pkg/stdlib/jsonpath.go
// JSONPath implementation for querying and manipulating JSON objects.
// JSONPath syntax reference: https://goessner.net/articles/JsonPath/
package stdlib

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

// JSONPath represents a parsed JSONPath expression
type JSONPath struct {
	segments []pathSegment
}

// pathSegment represents a single segment in a JSONPath
type pathSegment struct {
	// Type of segment: "root", "field", "index", "wildcard", "recursive", "filter", "slice"
	typ string

	// For field segments
	fieldName string

	// For index segments
	index int

	// For slice segments
	start, end, step int
	hasStart, hasEnd bool

	// For filter segments
	filterExpr string

	// For multiple indices
	indices []int
}

// ParseJSONPath parses a JSONPath expression string into a JSONPath struct
func ParseJSONPath(path string) (*JSONPath, error) {
	if path == "" || path == "$" {
		return &JSONPath{segments: []pathSegment{{typ: "root"}}}, nil
	}

	if !strings.HasPrefix(path, "$") {
		// Treat as relative path starting with field name
		path = "$." + path
	}

	jp := &JSONPath{
		segments: []pathSegment{{typ: "root"}},
	}

	// Remove leading $
	rest := path[1:]

	for len(rest) > 0 {
		seg, remaining, err := parseNextSegment(rest)
		if err != nil {
			return nil, err
		}
		if seg == nil {
			break
		}
		jp.segments = append(jp.segments, *seg)
		rest = remaining
	}

	return jp, nil
}

// parseNextSegment parses the next segment from the path string
func parseNextSegment(s string) (*pathSegment, string, error) {
	if len(s) == 0 {
		return nil, "", nil
	}

	// Check for recursive descent
	if strings.HasPrefix(s, "..") {
		remaining := s[2:]
		// Check if followed by field name or wildcard
		if len(remaining) > 0 && remaining[0] == '[' {
			// Recursive with bracket notation
			seg, rem, err := parseBracketSegment(remaining)
			if err != nil {
				return nil, "", err
			}
			seg.typ = "recursive_" + seg.typ
			return seg, rem, nil
		} else if len(remaining) > 0 && remaining[0] == '*' {
			return &pathSegment{typ: "recursive_wildcard"}, remaining[1:], nil
		} else {
			// Recursive with field name
			field, rem := parseFieldName(remaining)
			if field == "" {
				return nil, "", fmt.Errorf("expected field name after ..")
			}
			return &pathSegment{typ: "recursive_field", fieldName: field}, rem, nil
		}
	}

	// Check for bracket notation
	if s[0] == '[' {
		return parseBracketSegment(s)
	}

	// Check for dot notation
	if s[0] == '.' {
		remaining := s[1:]
		if len(remaining) == 0 {
			return nil, "", fmt.Errorf("unexpected end after '.'")
		}
		if remaining[0] == '*' {
			return &pathSegment{typ: "wildcard"}, remaining[1:], nil
		}
		field, rem := parseFieldName(remaining)
		if field == "" {
			return nil, "", fmt.Errorf("expected field name after '.'")
		}
		return &pathSegment{typ: "field", fieldName: field}, rem, nil
	}

	// Invalid start character - unexpected token
	return nil, "", fmt.Errorf("unexpected token '%c' at start of segment", s[0])
}

// parseBracketSegment parses a bracket notation segment [...]
func parseBracketSegment(s string) (*pathSegment, string, error) {
	if len(s) == 0 || s[0] != '[' {
		return nil, "", fmt.Errorf("expected '['")
	}

	// Find matching ] by tracking bracket depth
	depth := 0
	end := -1
	for i, c := range s {
		switch c {
		case '[', '(':
			depth++
		case ']', ')':
			depth--
			if depth == 0 && c == ']' {
				end = i
			}
		}
		if end >= 0 {
			break
		}
	}
	if end == -1 {
		return nil, "", fmt.Errorf("unmatched '['")
	}

	content := strings.TrimSpace(s[1:end])
	remaining := s[end+1:]

	if content == "*" {
		return &pathSegment{typ: "wildcard"}, remaining, nil
	}

	// Check for filter expression
	if strings.HasPrefix(content, "?(") {
		return &pathSegment{typ: "filter", filterExpr: content[2 : len(content)-1]}, remaining, nil
	}

	// Check for slice
	if strings.Contains(content, ":") {
		return parseSliceSegment(content, remaining)
	}

	// Check for multiple indices
	if strings.Contains(content, ",") {
		return parseMultiIndexSegment(content, remaining)
	}

	// Check for quoted string (field name)
	if (content[0] == '\'' && content[len(content)-1] == '\'') ||
		(content[0] == '"' && content[len(content)-1] == '"') {
		fieldName := content[1 : len(content)-1]
		return &pathSegment{typ: "field", fieldName: fieldName}, remaining, nil
	}

	// Try to parse as index
	idx, err := strconv.Atoi(content)
	if err != nil {
		return nil, "", fmt.Errorf("invalid bracket content: %s", content)
	}

	return &pathSegment{typ: "index", index: idx}, remaining, nil
}

// parseSliceSegment parses a slice segment like [start:end:step]
func parseSliceSegment(content, remaining string) (*pathSegment, string, error) {
	parts := strings.Split(content, ":")

	// Validate: slice can have at most 3 parts (start:end:step)
	if len(parts) > 3 {
		return nil, "", fmt.Errorf("invalid slice: too many parts (max 3, got %d)", len(parts))
	}

	seg := &pathSegment{typ: "slice", step: 1}

	if parts[0] != "" {
		idx, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, "", fmt.Errorf("invalid slice start: %s", parts[0])
		}
		seg.start = idx
		seg.hasStart = true
	}

	if len(parts) > 1 && parts[1] != "" {
		idx, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, "", fmt.Errorf("invalid slice end: %s", parts[1])
		}
		seg.end = idx
		seg.hasEnd = true
	}

	if len(parts) > 2 && parts[2] != "" {
		idx, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, "", fmt.Errorf("invalid slice step: %s", parts[2])
		}
		seg.step = idx
	}

	return seg, remaining, nil
}

// parseMultiIndexSegment parses multiple indices like [0,1,2]
func parseMultiIndexSegment(content, remaining string) (*pathSegment, string, error) {
	parts := strings.Split(content, ",")
	indices := make([]int, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		idx, err := strconv.Atoi(p)
		if err != nil {
			return nil, "", fmt.Errorf("invalid index: %s", p)
		}
		indices = append(indices, idx)
	}

	return &pathSegment{typ: "multi_index", indices: indices}, remaining, nil
}

// parseFieldName parses a field name from the start of the string
func parseFieldName(s string) (string, string) {
	if len(s) == 0 {
		return "", s
	}

	// Check for quoted field name
	if s[0] == '\'' || s[0] == '"' {
		quote := s[0]
		for i := 1; i < len(s); i++ {
			if s[i] == quote {
				return s[1:i], s[i+1:]
			}
		}
		return s[1:], ""
	}

	// Unquoted field name
	var i int
	for i = 0; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c > 127) {
			break
		}
	}

	return s[:i], s[i:]
}

// Get retrieves values from an object using the JSONPath
func (jp *JSONPath) Get(obj objects.Object) []objects.Object {
	if len(jp.segments) == 0 {
		return []objects.Object{obj}
	}

	// Start with root segment
	results := []objects.Object{obj}

	for _, seg := range jp.segments {
		if seg.typ == "root" {
			continue
		}

		var newResults []objects.Object
		for _, current := range results {
			newResults = append(newResults, jp.applySegment(current, seg)...)
		}
		results = newResults

		if len(results) == 0 {
			break
		}
	}

	return results
}

// applySegment applies a single path segment to an object
func (jp *JSONPath) applySegment(obj objects.Object, seg pathSegment) []objects.Object {
	switch seg.typ {
	case "field":
		return jp.getField(obj, seg.fieldName)
	case "index":
		return jp.getIndex(obj, seg.index)
	case "wildcard":
		return jp.getWildcard(obj)
	case "slice":
		return jp.getSlice(obj, seg)
	case "multi_index":
		return jp.getMultiIndex(obj, seg.indices)
	case "recursive_field":
		return jp.getRecursiveField(obj, seg.fieldName)
	case "recursive_wildcard":
		return jp.getRecursiveWildcard(obj)
	case "filter":
		return jp.getFilter(obj, seg.filterExpr)
	default:
		return nil
	}
}

// getField gets a field from an object
func (jp *JSONPath) getField(obj objects.Object, fieldName string) []objects.Object {
	m, ok := obj.(*objects.Map)
	if !ok {
		return nil
	}

	key := objects.NewString(fieldName)
	hashKey := key.HashKey()
	pair, exists := m.Pairs[hashKey]
	if !exists {
		return nil
	}

	return []objects.Object{pair.Value}
}

// getIndex gets an element by index from an array
func (jp *JSONPath) getIndex(obj objects.Object, index int) []objects.Object {
	arr, ok := obj.(*objects.Array)
	if !ok {
		return nil
	}

	if index < 0 {
		index = len(arr.Elements) + index
	}

	if index < 0 || index >= len(arr.Elements) {
		return nil
	}

	return []objects.Object{arr.Elements[index]}
}

// getWildcard gets all values from an object or array
func (jp *JSONPath) getWildcard(obj objects.Object) []objects.Object {
	switch v := obj.(type) {
	case *objects.Map:
		var results []objects.Object
		for _, pair := range v.Pairs {
			results = append(results, pair.Value)
		}
		return results
	case *objects.Array:
		return v.Elements
	default:
		return nil
	}
}

// getSlice gets a slice of an array
func (jp *JSONPath) getSlice(obj objects.Object, seg pathSegment) []objects.Object {
	arr, ok := obj.(*objects.Array)
	if !ok {
		return nil
	}

	length := len(arr.Elements)
	start := 0
	end := length

	if seg.hasStart {
		start = seg.start
		if start < 0 {
			start = length + start
		}
	}

	if seg.hasEnd {
		end = seg.end
		if end < 0 {
			end = length + end
		}
	}

	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start >= end {
		return nil
	}

	var results []objects.Object
	step := seg.step
	if step <= 0 {
		step = 1
	}

	for i := start; i < end; i += step {
		results = append(results, arr.Elements[i])
	}

	return results
}

// getMultiIndex gets multiple elements by index
func (jp *JSONPath) getMultiIndex(obj objects.Object, indices []int) []objects.Object {
	arr, ok := obj.(*objects.Array)
	if !ok {
		return nil
	}

	var results []objects.Object
	for _, idx := range indices {
		if idx < 0 {
			idx = len(arr.Elements) + idx
		}
		if idx >= 0 && idx < len(arr.Elements) {
			results = append(results, arr.Elements[idx])
		}
	}

	return results
}

// getRecursiveField recursively searches for a field
func (jp *JSONPath) getRecursiveField(obj objects.Object, fieldName string) []objects.Object {
	var results []objects.Object
	jp.collectRecursiveField(obj, fieldName, &results)
	return results
}

// collectRecursiveField recursively collects values for a field name
func (jp *JSONPath) collectRecursiveField(obj objects.Object, fieldName string, results *[]objects.Object) {
	// First, try to get the field directly
	if m, ok := obj.(*objects.Map); ok {
		key := objects.NewString(fieldName)
		hashKey := key.HashKey()
		if pair, exists := m.Pairs[hashKey]; exists {
			*results = append(*results, pair.Value)
		}

		// Recurse into all values
		for _, pair := range m.Pairs {
			jp.collectRecursiveField(pair.Value, fieldName, results)
		}
	}

	// Recurse into arrays
	if arr, ok := obj.(*objects.Array); ok {
		for _, elem := range arr.Elements {
			jp.collectRecursiveField(elem, fieldName, results)
		}
	}
}

// getRecursiveWildcard recursively gets all values
func (jp *JSONPath) getRecursiveWildcard(obj objects.Object) []objects.Object {
	var results []objects.Object
	jp.collectRecursiveWildcard(obj, &results)
	return results
}

// collectRecursiveWildcard recursively collects all values
func (jp *JSONPath) collectRecursiveWildcard(obj objects.Object, results *[]objects.Object) {
	switch v := obj.(type) {
	case *objects.Map:
		for _, pair := range v.Pairs {
			*results = append(*results, pair.Value)
			jp.collectRecursiveWildcard(pair.Value, results)
		}
	case *objects.Array:
		for _, elem := range v.Elements {
			*results = append(*results, elem)
			jp.collectRecursiveWildcard(elem, results)
		}
	}
}

// getFilter filters array elements based on an expression
func (jp *JSONPath) getFilter(obj objects.Object, expr string) []objects.Object {
	arr, ok := obj.(*objects.Array)
	if !ok {
		return nil
	}

	var results []objects.Object
	for _, elem := range arr.Elements {
		if jp.matchesFilter(elem, expr) {
			results = append(results, elem)
		}
	}

	return results
}

// matchesFilter checks if an element matches a filter expression
// Supports:
// - Comparison: @.price < 10, @.name == "test"
// - Logical: @.price > 5 && @.price < 20, @.active || @.pending
// - Regex: @.name =~ "pattern"
// - Contains: @.name contains "test", @.tags contains "tag1"
// - In: @.category in ["fiction", "drama"]
// - Not in: @.category nin ["fiction", "drama"]
// - Starts with: @.name startsWith "The"
// - Ends with: @.name endsWith "ing"
// - Exists: @.field (checks if field exists and is not null)
// - Empty: empty(@.field)
// - Between: @.price between [10, 100]
// - Is null: @.field isNull
// - Is not null: @.field isNotNull
// - Is type: @.value isType "number"
// - Absent: @.optional absent
func (jp *JSONPath) matchesFilter(obj objects.Object, expr string) bool {
	expr = strings.TrimSpace(expr)

	// Handle logical OR (lower precedence than AND)
	if orIdx := findLogicalOp(expr, "||"); orIdx >= 0 {
		left := strings.TrimSpace(expr[:orIdx])
		right := strings.TrimSpace(expr[orIdx+2:])
		return jp.matchesFilter(obj, left) || jp.matchesFilter(obj, right)
	}

	// Handle logical AND
	if andIdx := findLogicalOp(expr, "&&"); andIdx >= 0 {
		left := strings.TrimSpace(expr[:andIdx])
		right := strings.TrimSpace(expr[andIdx+2:])
		return jp.matchesFilter(obj, left) && jp.matchesFilter(obj, right)
	}

	// Handle NOT
	if strings.HasPrefix(expr, "!") {
		inner := strings.TrimSpace(expr[1:])
		// Special handling for !@.field - check if field is falsy
		if strings.HasPrefix(inner, "@") && !strings.ContainsAny(inner, " \t") {
			// This is a simple field reference - check if it's falsy
			val := jp.evalFilterExpr(obj, inner)
			return isFalsy(val)
		}
		return !jp.matchesFilter(obj, inner)
	}

	// Handle parentheses
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		return jp.matchesFilter(obj, expr[1:len(expr)-1])
	}

	// Handle "in" operator: @.category in ["a", "b"]
	if idx := findKeywordOpCase(expr, " in "); idx >= 0 {
		left := strings.TrimSpace(expr[:idx])
		right := strings.TrimSpace(expr[idx+4:])
		return jp.evalInOperator(obj, left, right, false)
	}

	// Handle "nin" (not in) operator: @.category nin ["a", "b"]
	if idx := findKeywordOpCase(expr, " nin "); idx >= 0 {
		left := strings.TrimSpace(expr[:idx])
		right := strings.TrimSpace(expr[idx+5:])
		return jp.evalInOperator(obj, left, right, true)
	}

	// Handle "contains" operator: @.name contains "test"
	if idx := findKeywordOpCase(expr, " contains "); idx >= 0 {
		left := strings.TrimSpace(expr[:idx])
		right := strings.TrimSpace(expr[idx+10:])
		return jp.evalContainsOperator(obj, left, right)
	}

	// Handle "startsWith" operator: @.name startsWith "The"
	if idx := findKeywordOpCase(expr, " startsWith "); idx >= 0 {
		left := strings.TrimSpace(expr[:idx])
		right := strings.TrimSpace(expr[idx+12:])
		return jp.evalStartsWithOperator(obj, left, right)
	}

	// Handle "endsWith" operator: @.name endsWith "ing"
	if idx := findKeywordOpCase(expr, " endsWith "); idx >= 0 {
		left := strings.TrimSpace(expr[:idx])
		right := strings.TrimSpace(expr[idx+10:])
		return jp.evalEndsWithOperator(obj, left, right)
	}

	// Handle "matches" or "=~" operator (regex): @.name =~ "pattern"
	if strings.Contains(expr, "=~") {
		parts := strings.SplitN(expr, "=~", 2)
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			return jp.evalRegexOperator(obj, left, right)
		}
	}

	// Handle "empty" function: empty(@.field)
	if strings.HasPrefix(expr, "empty(") && strings.HasSuffix(expr, ")") {
		inner := expr[6 : len(expr)-1]
		return jp.evalEmptyFunction(obj, inner)
	}

	// Handle "length" or "size" function: length(@.field) > 0
	if strings.HasPrefix(expr, "length(") || strings.HasPrefix(expr, "size(") {
		return jp.evalLengthFunction(obj, expr)
	}

	// Handle "between" operator: @.price between [10, 100]
	if idx := findKeywordOpCase(expr, " between "); idx >= 0 {
		left := strings.TrimSpace(expr[:idx])
		right := strings.TrimSpace(expr[idx+9:])
		return jp.evalBetweenOperator(obj, left, right)
	}

	// Handle "isNotNull" operator: @.field isNotNull
	if idx := findKeywordOpCase(expr, " isNotNull"); idx >= 0 {
		left := strings.TrimSpace(expr[:idx])
		return jp.evalIsNotNullOperator(obj, left)
	}

	// Handle "isNull" operator: @.field isNull
	if idx := findKeywordOpCase(expr, " isNull"); idx >= 0 {
		left := strings.TrimSpace(expr[:idx])
		return jp.evalIsNullOperator(obj, left)
	}

	// Handle "isType" operator: @.value isType "number"
	if idx := findKeywordOpCase(expr, " isType "); idx >= 0 {
		left := strings.TrimSpace(expr[:idx])
		right := strings.TrimSpace(expr[idx+8:])
		return jp.evalIsTypeOperator(obj, left, right)
	}

	// Handle "absent" operator: @.optional absent
	if idx := findKeywordOpCase(expr, " absent"); idx >= 0 {
		left := strings.TrimSpace(expr[:idx])
		return jp.evalAbsentOperator(obj, left)
	}

	// Handle comparison expressions
	comparisonOps := []string{"==", "!=", "<=", ">=", "<", ">"}
	for _, op := range comparisonOps {
		if strings.Contains(expr, op) {
			parts := strings.SplitN(expr, op, 2)
			if len(parts) != 2 {
				continue
			}

			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			leftVal := jp.evalFilterExpr(obj, left)
			rightVal := jp.evalFilterValue(right)

			return compareObjectsForFilter(leftVal, rightVal, op)
		}
	}

	// Handle existence check: @.field (returns true if field exists and is not null)
	if strings.HasPrefix(expr, "@") {
		val := jp.evalFilterExpr(obj, expr)
		_, isNull := val.(*objects.Null)
		return val != nil && !isNull
	}

	return false
}

// isFalsy checks if a value is falsy (null, false, 0, empty string, empty array, empty map)
func isFalsy(obj objects.Object) bool {
	if obj == nil {
		return true
	}
	switch v := obj.(type) {
	case *objects.Null:
		return true
	case *objects.Bool:
		return !v.Value
	case *objects.Int:
		return v.Value == 0
	case *objects.Float:
		return v.Value == 0.0
	case *objects.String:
		return v.Value == ""
	case *objects.Array:
		return len(v.Elements) == 0
	case *objects.Map:
		return len(v.Pairs) == 0
	default:
		return false
	}
}

// findKeywordOpCase finds a keyword operator (case-insensitive after @)
func findKeywordOpCase(expr string, op string) int {
	lowerExpr := strings.ToLower(expr)
	lowerOp := strings.ToLower(op)

	depth := 0
	inString := false
	stringChar := byte(0)

	for i := 0; i <= len(lowerExpr)-len(lowerOp); i++ {
		c := lowerExpr[i]

		if inString {
			if c == stringChar && (i == 0 || expr[i-1] != '\\') {
				inString = false
			}
			continue
		}

		switch c {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case '"', '\'':
			inString = true
			stringChar = c
		}

		if depth == 0 && lowerExpr[i:i+len(lowerOp)] == lowerOp {
			return i
		}
	}

	return -1
}

// findLogicalOp finds a logical operator (&& or ||) at the top level (not inside parentheses or strings)
func findLogicalOp(expr string, op string) int {
	depth := 0
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(expr)-1; i++ {
		c := expr[i]

		if inString {
			if c == stringChar && (i == 0 || expr[i-1] != '\\') {
				inString = false
			}
			continue
		}

		switch c {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case '"', '\'':
			inString = true
			stringChar = c
		}

		if depth == 0 && expr[i:i+len(op)] == op {
			return i
		}
	}

	return -1
}

// findKeywordOp finds a keyword operator at the top level
func findKeywordOp(expr string, op string) int {
	depth := 0
	inString := false
	stringChar := byte(0)

	for i := 0; i <= len(expr)-len(op); i++ {
		c := expr[i]

		if inString {
			if c == stringChar && (i == 0 || expr[i-1] != '\\') {
				inString = false
			}
			continue
		}

		switch c {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case '"', '\'':
			inString = true
			stringChar = c
		}

		if depth == 0 && expr[i:i+len(op)] == op {
			return i
		}
	}

	return -1
}

// evalInOperator evaluates the "in" and "nin" operators
func (jp *JSONPath) evalInOperator(obj objects.Object, left string, right string, negate bool) bool {
	leftVal := jp.evalFilterExpr(obj, left)
	rightVal := jp.evalFilterValue(right)

	arr, ok := rightVal.(*objects.Array)
	if !ok {
		return false
	}

	found := false
	for _, elem := range arr.Elements {
		if objectsEqual(leftVal, elem) {
			found = true
			break
		}
	}

	if negate {
		return !found
	}
	return found
}

// evalContainsOperator evaluates the "contains" operator
func (jp *JSONPath) evalContainsOperator(obj objects.Object, left string, right string) bool {
	leftVal := jp.evalFilterExpr(obj, left)
	rightVal := jp.evalFilterValue(right)

	// For strings: check substring
	if leftStr, ok := leftVal.(*objects.String); ok {
		if rightStr, ok := rightVal.(*objects.String); ok {
			return strings.Contains(leftStr.Value, rightStr.Value)
		}
	}

	// For arrays: check if element exists
	if leftArr, ok := leftVal.(*objects.Array); ok {
		for _, elem := range leftArr.Elements {
			if objectsEqual(elem, rightVal) {
				return true
			}
		}
	}

	return false
}

// evalStartsWithOperator evaluates the "startsWith" operator
func (jp *JSONPath) evalStartsWithOperator(obj objects.Object, left string, right string) bool {
	leftVal := jp.evalFilterExpr(obj, left)
	rightVal := jp.evalFilterValue(right)

	leftStr, ok1 := leftVal.(*objects.String)
	rightStr, ok2 := rightVal.(*objects.String)
	if !ok1 || !ok2 {
		return false
	}

	return strings.HasPrefix(leftStr.Value, rightStr.Value)
}

// evalEndsWithOperator evaluates the "endsWith" operator
func (jp *JSONPath) evalEndsWithOperator(obj objects.Object, left string, right string) bool {
	leftVal := jp.evalFilterExpr(obj, left)
	rightVal := jp.evalFilterValue(right)

	leftStr, ok1 := leftVal.(*objects.String)
	rightStr, ok2 := rightVal.(*objects.String)
	if !ok1 || !ok2 {
		return false
	}

	return strings.HasSuffix(leftStr.Value, rightStr.Value)
}

// evalRegexOperator evaluates the "=~" regex operator
func (jp *JSONPath) evalRegexOperator(obj objects.Object, left string, right string) bool {
	leftVal := jp.evalFilterExpr(obj, left)
	rightVal := jp.evalFilterValue(right)

	leftStr, ok1 := leftVal.(*objects.String)
	rightStr, ok2 := rightVal.(*objects.String)
	if !ok1 || !ok2 {
		return false
	}

	matched, err := regexp.MatchString(rightStr.Value, leftStr.Value)
	if err != nil {
		return false
	}
	return matched
}

// evalBetweenOperator evaluates the "between" operator
// Syntax: @.price between [10, 100] - checks if value is >= 10 and <= 100
func (jp *JSONPath) evalBetweenOperator(obj objects.Object, left string, right string) bool {
	leftVal := jp.evalFilterExpr(obj, left)
	rightVal := jp.evalFilterValue(right)

	// right should be an array with exactly 2 elements [min, max]
	arr, ok := rightVal.(*objects.Array)
	if !ok || len(arr.Elements) != 2 {
		return false
	}

	minVal := arr.Elements[0]
	maxVal := arr.Elements[1]

	// Check if leftVal >= minVal && leftVal <= maxVal
	cmpMin := compareNumbersForFilter(leftVal, minVal)
	cmpMax := compareNumbersForFilter(leftVal, maxVal)

	return cmpMin >= 0 && cmpMax <= 0
}

// evalIsNullOperator evaluates the "isNull" operator
// Syntax: @.field isNull - checks if field exists and its value is null
func (jp *JSONPath) evalIsNullOperator(obj objects.Object, left string) bool {
	leftVal := jp.evalFilterExpr(obj, left)

	if leftVal == nil {
		return true
	}

	_, isNull := leftVal.(*objects.Null)
	return isNull
}

// evalIsNotNullOperator evaluates the "isNotNull" operator
// Syntax: @.field isNotNull - checks if field exists and its value is not null
func (jp *JSONPath) evalIsNotNullOperator(obj objects.Object, left string) bool {
	leftVal := jp.evalFilterExpr(obj, left)

	if leftVal == nil {
		return false
	}

	_, isNull := leftVal.(*objects.Null)
	return !isNull
}

// evalIsTypeOperator evaluates the "isType" operator
// Syntax: @.value isType "number" - checks if value is of the specified type
// Supported types: "number", "string", "boolean", "array", "object", "null", "int", "float"
func (jp *JSONPath) evalIsTypeOperator(obj objects.Object, left string, right string) bool {
	leftVal := jp.evalFilterExpr(obj, left)
	rightVal := jp.evalFilterValue(right)

	typeName, ok := rightVal.(*objects.String)
	if !ok {
		return false
	}

	if leftVal == nil {
		return typeName.Value == "null"
	}

	switch strings.ToLower(typeName.Value) {
	case "number":
		_, isInt := leftVal.(*objects.Int)
		_, isFloat := leftVal.(*objects.Float)
		return isInt || isFloat
	case "int", "integer":
		_, isInt := leftVal.(*objects.Int)
		return isInt
	case "float":
		_, isFloat := leftVal.(*objects.Float)
		return isFloat
	case "string":
		_, isString := leftVal.(*objects.String)
		return isString
	case "boolean", "bool":
		_, isBool := leftVal.(*objects.Bool)
		return isBool
	case "array":
		_, isArray := leftVal.(*objects.Array)
		return isArray
	case "object", "map":
		_, isMap := leftVal.(*objects.Map)
		return isMap
	case "null":
		_, isNull := leftVal.(*objects.Null)
		return isNull
	default:
		return false
	}
}

// evalAbsentOperator evaluates the "absent" operator
// Syntax: @.optional absent - checks if field does not exist (different from null)
func (jp *JSONPath) evalAbsentOperator(obj objects.Object, left string) bool {
	// Check if the field path is valid and starts with @
	if !strings.HasPrefix(left, "@") {
		return false
	}

	// Parse the path after @
	pathStr := "$" + left[1:]
	path, err := ParseJSONPath(pathStr)
	if err != nil {
		return true // Invalid path is considered absent
	}

	// Check if the path exists in the object
	// We need to check if the field actually exists, not just its value
	segments := path.segments
	if len(segments) < 2 {
		return false
	}

	// Navigate through all but the last segment
	current := obj
	for i := 1; i < len(segments)-1; i++ {
		seg := segments[i]
		next := jp.navigateSegment(current, seg)
		if next == nil {
			return true // Parent path doesn't exist, so field is absent
		}
		current = next
	}

	// Check if the final field exists
	lastSeg := segments[len(segments)-1]
	return !jp.fieldExists(current, lastSeg)
}

// navigateSegment navigates to a single segment, returning nil if not found
func (jp *JSONPath) navigateSegment(obj objects.Object, seg pathSegment) objects.Object {
	switch seg.typ {
	case "field":
		m, ok := obj.(*objects.Map)
		if !ok {
			return nil
		}
		key := objects.NewString(seg.fieldName)
		hashKey := key.HashKey()
		if pair, exists := m.Pairs[hashKey]; exists {
			return pair.Value
		}
		return nil
	case "index":
		arr, ok := obj.(*objects.Array)
		if !ok {
			return nil
		}
		idx := seg.index
		if idx < 0 {
			idx = len(arr.Elements) + idx
		}
		if idx < 0 || idx >= len(arr.Elements) {
			return nil
		}
		return arr.Elements[idx]
	default:
		return nil
	}
}

// fieldExists checks if a field exists in an object
func (jp *JSONPath) fieldExists(obj objects.Object, seg pathSegment) bool {
	switch seg.typ {
	case "field":
		m, ok := obj.(*objects.Map)
		if !ok {
			return false
		}
		key := objects.NewString(seg.fieldName)
		hashKey := key.HashKey()
		_, exists := m.Pairs[hashKey]
		return exists
	case "index":
		arr, ok := obj.(*objects.Array)
		if !ok {
			return false
		}
		idx := seg.index
		if idx < 0 {
			idx = len(arr.Elements) + idx
		}
		return idx >= 0 && idx < len(arr.Elements)
	default:
		return false
	}
}

// evalEmptyFunction evaluates the "empty()" function
func (jp *JSONPath) evalEmptyFunction(obj objects.Object, inner string) bool {
	val := jp.evalFilterExpr(obj, inner)

	if val == nil {
		return true
	}

	switch v := val.(type) {
	case *objects.Null:
		return true
	case *objects.String:
		return v.Value == ""
	case *objects.Array:
		return len(v.Elements) == 0
	case *objects.Map:
		return len(v.Pairs) == 0
	default:
		return false
	}
}

// evalLengthFunction evaluates length() or size() functions
func (jp *JSONPath) evalLengthFunction(obj objects.Object, expr string) bool {
	// Parse the function call
	var inner string
	var rest string

	if strings.HasPrefix(expr, "length(") {
		// Find the matching closing paren
		depth := 1
		start := 7
		for i := start; i < len(expr); i++ {
			if expr[i] == '(' {
				depth++
			} else if expr[i] == ')' {
				depth--
				if depth == 0 {
					inner = expr[start:i]
					rest = strings.TrimSpace(expr[i+1:])
					break
				}
			}
		}
	} else if strings.HasPrefix(expr, "size(") {
		depth := 1
		start := 5
		for i := start; i < len(expr); i++ {
			if expr[i] == '(' {
				depth++
			} else if expr[i] == ')' {
				depth--
				if depth == 0 {
					inner = expr[start:i]
					rest = strings.TrimSpace(expr[i+1:])
					break
				}
			}
		}
	}

	if inner == "" || rest == "" {
		return false
	}

	val := jp.evalFilterExpr(obj, inner)

	var length int64
	switch v := val.(type) {
	case *objects.String:
		length = int64(len(v.Value))
	case *objects.Array:
		length = int64(len(v.Elements))
	case *objects.Map:
		length = int64(len(v.Pairs))
	default:
		return false
	}

	// Evaluate the comparison
	rest = strings.TrimSpace(rest)
	comparisonOps := []string{"==", "!=", "<=", ">=", "<", ">"}
	for _, op := range comparisonOps {
		if strings.HasPrefix(rest, op) {
			right := strings.TrimSpace(rest[len(op):])
			rightVal := jp.evalFilterValue(right)
			return compareObjectsForFilter(objects.NewInt(length), rightVal, op)
		}
	}

	return false
}

// evalFilterExpr evaluates a filter expression like @.field
func (jp *JSONPath) evalFilterExpr(obj objects.Object, expr string) objects.Object {
	expr = strings.TrimSpace(expr)

	// Handle nested function calls on fields: @.field.length()
	if idx := strings.Index(expr, ".length()"); idx > 0 && strings.HasSuffix(expr, ".length()") {
		fieldPath := expr[:idx]
		fieldVal := jp.evalFilterExpr(obj, fieldPath)
		switch v := fieldVal.(type) {
		case *objects.String:
			return objects.NewInt(int64(len(v.Value)))
		case *objects.Array:
			return objects.NewInt(int64(len(v.Elements)))
		case *objects.Map:
			return objects.NewInt(int64(len(v.Pairs)))
		default:
			return objects.NULL
		}
	}

	if idx := strings.Index(expr, ".size()"); idx > 0 && strings.HasSuffix(expr, ".size()") {
		fieldPath := expr[:idx]
		fieldVal := jp.evalFilterExpr(obj, fieldPath)
		switch v := fieldVal.(type) {
		case *objects.String:
			return objects.NewInt(int64(len(v.Value)))
		case *objects.Array:
			return objects.NewInt(int64(len(v.Elements)))
		case *objects.Map:
			return objects.NewInt(int64(len(v.Pairs)))
		default:
			return objects.NULL
		}
	}

	if !strings.HasPrefix(expr, "@") {
		return objects.NULL
	}

	// Parse the path after @
	path, err := ParseJSONPath("$" + expr[1:])
	if err != nil {
		return objects.NULL
	}

	results := path.Get(obj)
	if len(results) == 0 {
		return objects.NULL
	}

	return results[0]
}

// evalFilterValue evaluates a literal value in a filter expression
func (jp *JSONPath) evalFilterValue(s string) objects.Object {
	s = strings.TrimSpace(s)

	// Check for array literal: [1, 2, 3] or ["a", "b"]
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		return jp.parseArrayLiteral(s)
	}

	// Check for string literal
	if (len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"') ||
		(len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'') {
		return objects.NewString(s[1 : len(s)-1])
	}

	// Check for number
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		if f == float64(int64(f)) {
			return objects.NewInt(int64(f))
		}
		return objects.NewFloat(f)
	}

	// Check for boolean
	if s == "true" {
		return objects.TRUE
	}
	if s == "false" {
		return objects.FALSE
	}
	if s == "null" {
		return objects.NULL
	}

	return objects.NULL
}

// parseArrayLiteral parses an array literal like [1, 2, 3] or ["a", "b"]
func (jp *JSONPath) parseArrayLiteral(s string) objects.Object {
	s = strings.TrimSpace(s[1 : len(s)-1]) // Remove brackets
	if s == "" {
		return objects.NewArray([]objects.Object{})
	}

	var elements []objects.Object
	depth := 0
	inString := false
	stringChar := byte(0)
	start := 0

	for i := 0; i <= len(s); i++ {
		var c byte
		if i < len(s) {
			c = s[i]
		}

		if inString {
			if c == stringChar && (i == 0 || s[i-1] != '\\') {
				inString = false
			}
			continue
		}

		switch c {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case '"', '\'':
			inString = true
			stringChar = c
		}

		if (c == ',' && depth == 0) || i == len(s) {
			part := strings.TrimSpace(s[start:i])
			if part != "" {
				elements = append(elements, jp.evalFilterValue(part))
			}
			start = i + 1
		}
	}

	return objects.NewArray(elements)
}

// compareObjectsForFilter compares two objects with an operator
func compareObjectsForFilter(left, right objects.Object, op string) bool {
	switch op {
	case "==":
		return objectsEqual(left, right)
	case "!=":
		return !objectsEqual(left, right)
	case "<":
		return compareNumbersForFilter(left, right) < 0
	case ">":
		return compareNumbersForFilter(left, right) > 0
	case "<=":
		return compareNumbersForFilter(left, right) <= 0
	case ">=":
		return compareNumbersForFilter(left, right) >= 0
	}
	return false
}

// objectsEqual checks if two objects are equal
func objectsEqual(a, b objects.Object) bool {
	if a == nil || b == nil {
		return a == b
	}

	switch va := a.(type) {
	case *objects.Int:
		if vb, ok := b.(*objects.Int); ok {
			return va.Value == vb.Value
		}
		if vb, ok := b.(*objects.Float); ok {
			return float64(va.Value) == vb.Value
		}
	case *objects.Float:
		if vb, ok := b.(*objects.Float); ok {
			return va.Value == vb.Value
		}
		if vb, ok := b.(*objects.Int); ok {
			return va.Value == float64(vb.Value)
		}
	case *objects.String:
		if vb, ok := b.(*objects.String); ok {
			return va.Value == vb.Value
		}
	case *objects.Bool:
		if vb, ok := b.(*objects.Bool); ok {
			return va.Value == vb.Value
		}
	case *objects.Null:
		_, ok := b.(*objects.Null)
		return ok
	}

	return false
}

// compareNumbersForFilter compares two numeric objects
func compareNumbersForFilter(a, b objects.Object) int {
	var aVal, bVal float64

	switch va := a.(type) {
	case *objects.Int:
		aVal = float64(va.Value)
	case *objects.Float:
		aVal = va.Value
	default:
		return 0
	}

	switch vb := b.(type) {
	case *objects.Int:
		bVal = float64(vb.Value)
	case *objects.Float:
		bVal = vb.Value
	default:
		return 0
	}

	if aVal < bVal {
		return -1
	}
	if aVal > bVal {
		return 1
	}
	return 0
}

// Set sets a value at the specified path
func (jp *JSONPath) Set(obj objects.Object, value objects.Object) (objects.Object, error) {
	if len(jp.segments) <= 1 {
		return value, nil
	}

	// Clone the object to avoid mutation
	result := cloneObject(obj)

	// Navigate to the parent of the target
	current := result
	for i := 1; i < len(jp.segments)-1; i++ {
		seg := jp.segments[i]
		next := jp.navigateOrCreate(current, seg)
		if next == nil {
			return nil, fmt.Errorf("cannot navigate to path segment %d", i)
		}
		current = next
	}

	// Set the value at the final segment
	lastSeg := jp.segments[len(jp.segments)-1]
	if err := jp.setValue(current, lastSeg, value); err != nil {
		return nil, err
	}

	return result, nil
}

// navigateOrCreate navigates through an object, creating intermediate objects if needed
func (jp *JSONPath) navigateOrCreate(obj objects.Object, seg pathSegment) objects.Object {
	switch seg.typ {
	case "field":
		m, ok := obj.(*objects.Map)
		if !ok {
			return nil
		}
		key := objects.NewString(seg.fieldName)
		hashKey := key.HashKey()
		if pair, exists := m.Pairs[hashKey]; exists {
			return pair.Value
		}
		// Create new map entry
		newObj := objects.NewMap(make(map[objects.HashKey]objects.MapPair))
		m.Pairs[hashKey] = objects.MapPair{Key: key, Value: newObj}
		return newObj
	case "index":
		arr, ok := obj.(*objects.Array)
		if !ok {
			return nil
		}
		if seg.index < 0 || seg.index >= len(arr.Elements) {
			return nil
		}
		return arr.Elements[seg.index]
	}
	return nil
}

// setValue sets a value at a path segment
func (jp *JSONPath) setValue(obj objects.Object, seg pathSegment, value objects.Object) error {
	switch seg.typ {
	case "field":
		m, ok := obj.(*objects.Map)
		if !ok {
			return fmt.Errorf("cannot set field on non-map object")
		}
		key := objects.NewString(seg.fieldName)
		hashKey := key.HashKey()
		m.Pairs[hashKey] = objects.MapPair{Key: key, Value: value}
		return nil
	case "index":
		arr, ok := obj.(*objects.Array)
		if !ok {
			return fmt.Errorf("cannot set index on non-array object")
		}
		idx := seg.index
		if idx < 0 {
			idx = len(arr.Elements) + idx
		}
		if idx < 0 || idx >= len(arr.Elements) {
			return fmt.Errorf("index out of bounds: %d", idx)
		}
		arr.Elements[idx] = value
		return nil
	}
	return fmt.Errorf("unsupported segment type for setting: %s", seg.typ)
}

// Delete deletes values at the specified path
func (jp *JSONPath) Delete(obj objects.Object) (objects.Object, error) {
	if len(jp.segments) <= 1 {
		return nil, fmt.Errorf("cannot delete root")
	}

	// Clone the object
	result := cloneObject(obj)

	// Navigate to the parent
	current := result
	for i := 1; i < len(jp.segments)-1; i++ {
		seg := jp.segments[i]
		next := jp.navigate(current, seg)
		if next == nil {
			return result, nil // Path doesn't exist, nothing to delete
		}
		current = next
	}

	// Delete at the final segment
	lastSeg := jp.segments[len(jp.segments)-1]
	jp.deleteValue(current, lastSeg)

	return result, nil
}

// navigate navigates through an object without creating intermediate objects
func (jp *JSONPath) navigate(obj objects.Object, seg pathSegment) objects.Object {
	switch seg.typ {
	case "field":
		m, ok := obj.(*objects.Map)
		if !ok {
			return nil
		}
		key := objects.NewString(seg.fieldName)
		hashKey := key.HashKey()
		if pair, exists := m.Pairs[hashKey]; exists {
			return pair.Value
		}
	case "index":
		arr, ok := obj.(*objects.Array)
		if !ok {
			return nil
		}
		idx := seg.index
		if idx < 0 {
			idx = len(arr.Elements) + idx
		}
		if idx >= 0 && idx < len(arr.Elements) {
			return arr.Elements[idx]
		}
	}
	return nil
}

// deleteValue deletes a value at a path segment
func (jp *JSONPath) deleteValue(obj objects.Object, seg pathSegment) {
	switch seg.typ {
	case "field":
		if m, ok := obj.(*objects.Map); ok {
			key := objects.NewString(seg.fieldName)
			hashKey := key.HashKey()
			delete(m.Pairs, hashKey)
		}
	case "index":
		if arr, ok := obj.(*objects.Array); ok {
			idx := seg.index
			if idx < 0 {
				idx = len(arr.Elements) + idx
			}
			if idx >= 0 && idx < len(arr.Elements) {
				arr.Elements = append(arr.Elements[:idx], arr.Elements[idx+1:]...)
			}
		}
	}
}

// Paths returns all JSONPath strings that lead to values in the object
func Paths(obj objects.Object) []string {
	var paths []string
	collectPaths(obj, "", &paths)
	return paths
}

// collectPaths recursively collects all paths in an object
func collectPaths(obj objects.Object, prefix string, paths *[]string) {
	switch v := obj.(type) {
	case *objects.Map:
		for _, pair := range v.Pairs {
			if keyStr, ok := pair.Key.(*objects.String); ok {
				var path string
				if prefix == "" {
					path = escapeFieldName(keyStr.Value)
				} else {
					path = prefix + "." + escapeFieldName(keyStr.Value)
				}
				*paths = append(*paths, path)
				collectPaths(pair.Value, path, paths)
			}
		}
	case *objects.Array:
		for i, elem := range v.Elements {
			path := fmt.Sprintf("%s[%d]", prefix, i)
			*paths = append(*paths, path)
			collectPaths(elem, path, paths)
		}
	}
}

// escapeFieldName escapes a field name for use in a path
func escapeFieldName(name string) string {
	// Check if name needs escaping
	needsEscape := false
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_') {
			needsEscape = true
			break
		}
	}

	if needsEscape {
		return fmt.Sprintf("['%s']", strings.ReplaceAll(name, "'", "\\'"))
	}
	return name
}

// cloneObject creates a deep clone of an object
func cloneObject(obj objects.Object) objects.Object {
	if obj == nil {
		return nil
	}

	switch v := obj.(type) {
	case *objects.Map:
		pairs := make(map[objects.HashKey]objects.MapPair)
		for key, pair := range v.Pairs {
			pairs[key] = objects.MapPair{
				Key:   pair.Key,
				Value: cloneObject(pair.Value),
			}
		}
		return objects.NewMap(pairs)
	case *objects.Array:
		elements := make([]objects.Object, len(v.Elements))
		for i, elem := range v.Elements {
			elements[i] = cloneObject(elem)
		}
		return objects.NewArray(elements)
	default:
		// Immutable objects don't need cloning
		return obj
	}
}

// JSONPathMatch represents a match result with its path
type JSONPathMatch struct {
	Path  string
	Value objects.Object
}

// GetWithPath retrieves values along with their paths
func (jp *JSONPath) GetWithPath(obj objects.Object) []JSONPathMatch {
	if len(jp.segments) == 0 {
		return []JSONPathMatch{{Path: "$", Value: obj}}
	}

	results := []JSONPathMatch{{Path: "$", Value: obj}}

	for _, seg := range jp.segments {
		if seg.typ == "root" {
			continue
		}

		var newResults []JSONPathMatch
		for _, current := range results {
			newResults = append(newResults, jp.applySegmentWithPath(current, seg)...)
		}
		results = newResults

		if len(results) == 0 {
			break
		}
	}

	return results
}

// applySegmentWithPath applies a segment and tracks the path
func (jp *JSONPath) applySegmentWithPath(match JSONPathMatch, seg pathSegment) []JSONPathMatch {
	var results []JSONPathMatch

	switch seg.typ {
	case "field":
		m, ok := match.Value.(*objects.Map)
		if !ok {
			return nil
		}
		key := objects.NewString(seg.fieldName)
		hashKey := key.HashKey()
		if pair, exists := m.Pairs[hashKey]; exists {
			path := match.Path + "." + escapeFieldName(seg.fieldName)
			results = append(results, JSONPathMatch{Path: path, Value: pair.Value})
		}

	case "index":
		arr, ok := match.Value.(*objects.Array)
		if !ok {
			return nil
		}
		idx := seg.index
		if idx < 0 {
			idx = len(arr.Elements) + idx
		}
		if idx >= 0 && idx < len(arr.Elements) {
			path := fmt.Sprintf("%s[%d]", match.Path, idx)
			results = append(results, JSONPathMatch{Path: path, Value: arr.Elements[idx]})
		}

	case "wildcard":
		switch v := match.Value.(type) {
		case *objects.Map:
			for _, pair := range v.Pairs {
				if keyStr, ok := pair.Key.(*objects.String); ok {
					path := match.Path + "." + escapeFieldName(keyStr.Value)
					results = append(results, JSONPathMatch{Path: path, Value: pair.Value})
				}
			}
		case *objects.Array:
			for i, elem := range v.Elements {
				path := fmt.Sprintf("%s[%d]", match.Path, i)
				results = append(results, JSONPathMatch{Path: path, Value: elem})
			}
		}

	case "recursive_field":
		jp.collectRecursiveFieldWithPath(match.Value, seg.fieldName, match.Path, &results)

	case "recursive_wildcard":
		jp.collectRecursiveWildcardWithPath(match.Value, match.Path, &results)
	}

	return results
}

// collectRecursiveFieldWithPath recursively collects field matches with paths
func (jp *JSONPath) collectRecursiveFieldWithPath(obj objects.Object, fieldName, currentPath string, results *[]JSONPathMatch) {
	if m, ok := obj.(*objects.Map); ok {
		// Check for the field
		key := objects.NewString(fieldName)
		hashKey := key.HashKey()
		if pair, exists := m.Pairs[hashKey]; exists {
			path := currentPath + "." + escapeFieldName(fieldName)
			*results = append(*results, JSONPathMatch{Path: path, Value: pair.Value})
		}

		// Recurse
		for _, pair := range m.Pairs {
			if keyStr, ok := pair.Key.(*objects.String); ok {
				newPath := currentPath + "." + escapeFieldName(keyStr.Value)
				jp.collectRecursiveFieldWithPath(pair.Value, fieldName, newPath, results)
			}
		}
	}

	if arr, ok := obj.(*objects.Array); ok {
		for i, elem := range arr.Elements {
			newPath := fmt.Sprintf("%s[%d]", currentPath, i)
			jp.collectRecursiveFieldWithPath(elem, fieldName, newPath, results)
		}
	}
}

// collectRecursiveWildcardWithPath recursively collects all values with paths
func (jp *JSONPath) collectRecursiveWildcardWithPath(obj objects.Object, currentPath string, results *[]JSONPathMatch) {
	switch v := obj.(type) {
	case *objects.Map:
		for _, pair := range v.Pairs {
			if keyStr, ok := pair.Key.(*objects.String); ok {
				path := currentPath + "." + escapeFieldName(keyStr.Value)
				*results = append(*results, JSONPathMatch{Path: path, Value: pair.Value})
				jp.collectRecursiveWildcardWithPath(pair.Value, path, results)
			}
		}
	case *objects.Array:
		for i, elem := range v.Elements {
			path := fmt.Sprintf("%s[%d]", currentPath, i)
			*results = append(*results, JSONPathMatch{Path: path, Value: elem})
			jp.collectRecursiveWildcardWithPath(elem, path, results)
		}
	}
}

// filterRegex is used for simple filter expression parsing
var filterRegex = regexp.MustCompile(`@\.(\w+)\s*(==|!=|<|>|<=|>=)\s*(.+)`)
