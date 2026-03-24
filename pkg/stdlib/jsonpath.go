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
	hasStart, hasEnd  bool

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

	return nil, s, nil
}

// parseBracketSegment parses a bracket notation segment [...]
func parseBracketSegment(s string) (*pathSegment, string, error) {
	if len(s) == 0 || s[0] != '[' {
		return nil, "", fmt.Errorf("expected '['")
	}

	// Find matching ]
	end := strings.Index(s, "]")
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
// Supports simple expressions like: @.price < 10, @.name == "test"
func (jp *JSONPath) matchesFilter(obj objects.Object, expr string) bool {
	expr = strings.TrimSpace(expr)

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

	return false
}

// evalFilterExpr evaluates a filter expression like @.field
func (jp *JSONPath) evalFilterExpr(obj objects.Object, expr string) objects.Object {
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
	collectPaths(obj, "$", &paths)
	return paths
}

// collectPaths recursively collects all paths in an object
func collectPaths(obj objects.Object, prefix string, paths *[]string) {
	switch v := obj.(type) {
	case *objects.Map:
		for _, pair := range v.Pairs {
			if keyStr, ok := pair.Key.(*objects.String); ok {
				path := prefix + "." + escapeFieldName(keyStr.Value)
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
