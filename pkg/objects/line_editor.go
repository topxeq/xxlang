// pkg/objects/line_editor.go
// LineEditor object for Xxlang - line-based text editing functionality.
package objects

import (
	"bufio"
	"errors"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

// errNoFilePath is returned when trying to save without a file path
var errNoFilePath = errors.New("no file path associated with LineEditor")

// LineEditor represents a line-based text editor object.
// It stores lines in memory and provides various editing operations.
type LineEditor struct {
	lines    []string // Internal storage of lines (without newline characters)
	modified bool     // Whether content has been modified
	filePath string   // Associated file path (may be empty)
}

// Type returns the object type.
func (le *LineEditor) Type() ObjectType { return LineEditorType }

// TypeTag returns the fast type tag.
func (le *LineEditor) TypeTag() TypeTag { return TagLineEditor }

// Inspect returns a string representation of the LineEditor.
func (le *LineEditor) Inspect() string {
	return "LineEditor(lines=" + strconv.Itoa(len(le.lines)) + ", modified=" + strconv.FormatBool(le.modified) + ")"
}

// ToBool returns true (LineEditor is always truthy).
func (le *LineEditor) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the LineEditor.
func (le *LineEditor) HashKey() HashKey {
	return HashKey{
		Type:  LineEditorType,
		Value: uint64(uintptr(unsafe.Pointer(le))),
	}
}

// NewLineEditor creates a new empty LineEditor.
func NewLineEditor() *LineEditor {
	return &LineEditor{
		lines:    make([]string, 0),
		modified: false,
		filePath: "",
	}
}

// NewLineEditorWithCapacity creates a new LineEditor with initial capacity.
func NewLineEditorWithCapacity(capacity int) *LineEditor {
	return &LineEditor{
		lines:    make([]string, 0, capacity),
		modified: false,
		filePath: "",
	}
}

// NewLineEditorFromLines creates a LineEditor from a string array.
func NewLineEditorFromLines(lines []string) *LineEditor {
	return &LineEditor{
		lines:    lines,
		modified: false,
		filePath: "",
	}
}

// NewLineEditorFromText creates a LineEditor from a text string.
// The text is split by newlines.
func NewLineEditorFromText(text string) *LineEditor {
	if text == "" {
		return NewLineEditor()
	}
	// Split by any newline type (Unix, Windows, or old Mac)
	lines := strings.Split(text, "\n")
	// Handle \r\n and \r
	for i, line := range lines {
		lines[i] = strings.TrimSuffix(line, "\r")
	}
	// Remove trailing empty line if text ended with newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return &LineEditor{
		lines:    lines,
		modified: false,
		filePath: "",
	}
}

// NewLineEditorFromFile opens a file and creates a LineEditor from its contents.
func NewLineEditorFromFile(path string) (*LineEditor, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if lines == nil {
		lines = make([]string, 0)
	}

	return &LineEditor{
		lines:    lines,
		modified: false,
		filePath: path,
	}, nil
}

// ============================================================
// Basic Operations
// ============================================================

// LineCount returns the number of lines.
func (le *LineEditor) LineCount() int {
	return len(le.lines)
}

// IsEmpty checks if the editor is empty.
func (le *LineEditor) IsEmpty() bool {
	return len(le.lines) == 0
}

// IsModified checks if the content has been modified.
func (le *LineEditor) IsModified() bool {
	return le.modified
}

// normalizeIndex converts a 1-based index (including negative) to 0-based.
// Returns -1 if the index is out of range.
func (le *LineEditor) normalizeIndex(n int) int {
	lineCount := len(le.lines)
	if lineCount == 0 {
		return -1
	}

	// Handle negative indices
	if n < 0 {
		n = lineCount + n + 1 // -1 means last line, -2 means second to last
	}

	// Convert to 0-based
	n--

	// Check bounds
	if n < 0 || n >= lineCount {
		return -1
	}

	return n
}

// GetLine returns the line at index n (1-based, supports negative index).
// Returns empty string and false if index is out of range.
func (le *LineEditor) GetLine(n int) (string, bool) {
	idx := le.normalizeIndex(n)
	if idx < 0 {
		return "", false
	}
	return le.lines[idx], true
}

// SetLine sets the line at index n (1-based). Returns false if out of range.
func (le *LineEditor) SetLine(n int, text string) bool {
	idx := le.normalizeIndex(n)
	if idx < 0 {
		return false
	}
	le.lines[idx] = text
	le.modified = true
	return true
}

// AddLine adds a line at the end.
func (le *LineEditor) AddLine(text string) {
	le.lines = append(le.lines, text)
	le.modified = true
}

// InsertLine inserts a line before index n (1-based). Returns false if out of range.
func (le *LineEditor) InsertLine(n int, text string) bool {
	lineCount := len(le.lines)

	// Handle negative indices
	if n < 0 {
		n = lineCount + n + 1
	}

	// Convert to 0-based
	idx := n - 1

	// Allow insertion at the end (idx == lineCount)
	if idx < 0 || idx > lineCount {
		return false
	}

	le.lines = append(le.lines[:idx], append([]string{text}, le.lines[idx:]...)...)
	le.modified = true
	return true
}

// DeleteLine deletes the line at index n (1-based). Returns false if out of range.
func (le *LineEditor) DeleteLine(n int) bool {
	idx := le.normalizeIndex(n)
	if idx < 0 {
		return false
	}
	le.lines = append(le.lines[:idx], le.lines[idx+1:]...)
	le.modified = true
	return true
}

// DeleteLines deletes lines in the range [start, end] (1-based, inclusive).
func (le *LineEditor) DeleteLines(start, end int) bool {
	lineCount := len(le.lines)
	if lineCount == 0 {
		return false
	}

	// Normalize indices
	startIdx := start - 1
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= lineCount {
		return false
	}

	endIdx := end - 1
	if endIdx < 0 {
		endIdx = lineCount + end // Handle negative
	}
	if endIdx >= lineCount {
		endIdx = lineCount - 1
	}

	if startIdx > endIdx {
		return false
	}

	le.lines = append(le.lines[:startIdx], le.lines[endIdx+1:]...)
	le.modified = true
	return true
}

// ============================================================
// Line Range Operations
// ============================================================

// GetLines returns lines in the range [start, end] (1-based, inclusive).
func (le *LineEditor) GetLines(start, end int) []string {
	lineCount := len(le.lines)
	if lineCount == 0 {
		return nil
	}

	// Normalize indices
	startIdx := start - 1
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= lineCount {
		return nil
	}

	endIdx := end - 1
	if endIdx < 0 {
		endIdx = lineCount + end
	}
	if endIdx >= lineCount {
		endIdx = lineCount - 1
	}

	if startIdx > endIdx {
		return nil
	}

	result := make([]string, endIdx-startIdx+1)
	copy(result, le.lines[startIdx:endIdx+1])
	return result
}

// SetLines replaces lines in the range starting from start (1-based).
func (le *LineEditor) SetLines(start int, newLines []string) bool {
	lineCount := len(le.lines)
	startIdx := start - 1

	if startIdx < 0 || startIdx > lineCount {
		return false
	}

	// Replace the lines
	le.lines = append(le.lines[:startIdx], append(newLines, le.lines[startIdx:]...)...)
	le.modified = true
	return true
}

// AppendLines appends multiple lines at the end.
func (le *LineEditor) AppendLines(newLines []string) {
	le.lines = append(le.lines, newLines...)
	le.modified = true
}

// Clear removes all lines.
func (le *LineEditor) Clear() {
	le.lines = make([]string, 0)
	le.modified = true
}

// ============================================================
// Search Operations
// ============================================================

// Find returns line numbers (1-based) containing the text.
func (le *LineEditor) Find(text string) []int {
	var result []int
	for i, line := range le.lines {
		if strings.Contains(line, text) {
			result = append(result, i+1) // 1-based
		}
	}
	return result
}

// FindRegex returns line numbers (1-based) matching the pattern.
func (le *LineEditor) FindRegex(pattern string) ([]int, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	var result []int
	for i, line := range le.lines {
		if re.MatchString(line) {
			result = append(result, i+1)
		}
	}
	return result, nil
}

// FindAll returns all lines containing the text.
func (le *LineEditor) FindAll(text string) []string {
	var result []string
	for _, line := range le.lines {
		if strings.Contains(line, text) {
			result = append(result, line)
		}
	}
	return result
}

// FindFirst returns the first line number (1-based) containing the text, or 0 if not found.
func (le *LineEditor) FindFirst(text string) int {
	for i, line := range le.lines {
		if strings.Contains(line, text) {
			return i + 1
		}
	}
	return 0
}

// FindLast returns the last line number (1-based) containing the text, or 0 if not found.
func (le *LineEditor) FindLast(text string) int {
	for i := len(le.lines) - 1; i >= 0; i-- {
		if strings.Contains(le.lines[i], text) {
			return i + 1
		}
	}
	return 0
}

// Grep filters lines containing text and returns a new LineEditor.
func (le *LineEditor) Grep(text string) *LineEditor {
	var result []string
	for _, line := range le.lines {
		if strings.Contains(line, text) {
			result = append(result, line)
		}
	}
	return &LineEditor{
		lines:    result,
		modified: false,
		filePath: "",
	}
}

// GrepRegex filters lines matching pattern and returns a new LineEditor.
func (le *LineEditor) GrepRegex(pattern string) (*LineEditor, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, line := range le.lines {
		if re.MatchString(line) {
			result = append(result, line)
		}
	}
	return &LineEditor{
		lines:    result,
		modified: false,
		filePath: "",
	}, nil
}

// GrepNot filters lines NOT containing text and returns a new LineEditor.
func (le *LineEditor) GrepNot(text string) *LineEditor {
	var result []string
	for _, line := range le.lines {
		if !strings.Contains(line, text) {
			result = append(result, line)
		}
	}
	return &LineEditor{
		lines:    result,
		modified: false,
		filePath: "",
	}
}

// GrepNotRegex filters lines NOT matching pattern and returns a new LineEditor.
func (le *LineEditor) GrepNotRegex(pattern string) (*LineEditor, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, line := range le.lines {
		if !re.MatchString(line) {
			result = append(result, line)
		}
	}
	return &LineEditor{
		lines:    result,
		modified: false,
		filePath: "",
	}, nil
}

// ============================================================
// Replace Operations
// ============================================================

// Replace replaces all occurrences of old text with new text in all lines.
// Returns the number of replacements made.
func (le *LineEditor) Replace(old, new string) int {
	count := 0
	for i, line := range le.lines {
		n := strings.Count(line, old)
		if n > 0 {
			le.lines[i] = strings.ReplaceAll(line, old, new)
			count += n
		}
	}
	if count > 0 {
		le.modified = true
	}
	return count
}

// ReplaceLine replaces only in the specified line (1-based).
func (le *LineEditor) ReplaceLine(n int, old, new string) int {
	idx := le.normalizeIndex(n)
	if idx < 0 {
		return 0
	}
	count := strings.Count(le.lines[idx], old)
	if count > 0 {
		le.lines[idx] = strings.ReplaceAll(le.lines[idx], old, new)
		le.modified = true
	}
	return count
}

// ReplaceFirst replaces the first occurrence globally.
func (le *LineEditor) ReplaceFirst(old, new string) bool {
	for i, line := range le.lines {
		idx := strings.Index(line, old)
		if idx >= 0 {
			le.lines[i] = line[:idx] + new + line[idx+len(old):]
			le.modified = true
			return true
		}
	}
	return false
}

// ReplaceLast replaces the last occurrence globally.
func (le *LineEditor) ReplaceLast(old, new string) bool {
	for i := len(le.lines) - 1; i >= 0; i-- {
		idx := strings.LastIndex(le.lines[i], old)
		if idx >= 0 {
			le.lines[i] = le.lines[i][:idx] + new + le.lines[i][idx+len(old):]
			le.modified = true
			return true
		}
	}
	return false
}

// ReplaceRegex replaces all matches of pattern with new text.
func (le *LineEditor) ReplaceRegex(pattern, new string) (int, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return 0, err
	}

	count := 0
	for i, line := range le.lines {
		newLine := re.ReplaceAllString(line, new)
		if newLine != line {
			le.lines[i] = newLine
			count++
		}
	}
	if count > 0 {
		le.modified = true
	}
	return count, nil
}

// ReplaceRange replaces within the specified line range [start, end] (1-based).
func (le *LineEditor) ReplaceRange(start, end int, old, new string) int {
	lineCount := len(le.lines)
	if lineCount == 0 {
		return 0
	}

	startIdx := start - 1
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= lineCount {
		return 0
	}

	endIdx := end - 1
	if endIdx < 0 {
		endIdx = lineCount + end
	}
	if endIdx >= lineCount {
		endIdx = lineCount - 1
	}

	if startIdx > endIdx {
		return 0
	}

	count := 0
	for i := startIdx; i <= endIdx; i++ {
		n := strings.Count(le.lines[i], old)
		if n > 0 {
			le.lines[i] = strings.ReplaceAll(le.lines[i], old, new)
			count += n
		}
	}
	if count > 0 {
		le.modified = true
	}
	return count
}

// ============================================================
// Sort and Unique Operations
// ============================================================

// Sort sorts lines alphabetically (ascending).
func (le *LineEditor) Sort() {
	sort.Strings(le.lines)
	le.modified = true
}

// SortDesc sorts lines alphabetically (descending).
func (le *LineEditor) SortDesc() {
	sort.Sort(sort.Reverse(sort.StringSlice(le.lines)))
	le.modified = true
}

// SortNum sorts lines numerically (ascending).
// Non-numeric lines are placed at the end.
func (le *LineEditor) SortNum() {
	sort.Slice(le.lines, func(i, j int) bool {
		numI, errI := strconv.ParseFloat(le.lines[i], 64)
		numJ, errJ := strconv.ParseFloat(le.lines[j], 64)

		// Non-numeric lines go to the end
		if errI != nil && errJ != nil {
			return le.lines[i] < le.lines[j] // Both non-numeric: sort alphabetically
		}
		if errI != nil {
			return false // i is non-numeric, goes after
		}
		if errJ != nil {
			return true // j is non-numeric, goes after
		}
		return numI < numJ
	})
	le.modified = true
}

// SortNumDesc sorts lines numerically (descending).
func (le *LineEditor) SortNumDesc() {
	sort.Slice(le.lines, func(i, j int) bool {
		numI, errI := strconv.ParseFloat(le.lines[i], 64)
		numJ, errJ := strconv.ParseFloat(le.lines[j], 64)

		if errI != nil && errJ != nil {
			return le.lines[i] > le.lines[j]
		}
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}
		return numI > numJ
	})
	le.modified = true
}

// SortByCol sorts by specified column (1-based).
func (le *LineEditor) SortByCol(col int, sep string) {
	if col < 1 {
		return
	}

	sort.Slice(le.lines, func(i, j int) bool {
		colI := getCol(le.lines[i], col, sep)
		colJ := getCol(le.lines[j], col, sep)
		return colI < colJ
	})
	le.modified = true
}

// SortByColNum sorts by specified column numerically (1-based).
func (le *LineEditor) SortByColNum(col int, sep string) {
	if col < 1 {
		return
	}

	sort.Slice(le.lines, func(i, j int) bool {
		colI := getCol(le.lines[i], col, sep)
		colJ := getCol(le.lines[j], col, sep)

		numI, errI := strconv.ParseFloat(colI, 64)
		numJ, errJ := strconv.ParseFloat(colJ, 64)

		if errI != nil && errJ != nil {
			return colI < colJ
		}
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}
		return numI < numJ
	})
	le.modified = true
}

// getCol extracts the nth column (1-based) from a line.
func getCol(line string, col int, sep string) string {
	if sep == "" {
		return line
	}
	parts := strings.Split(line, sep)
	if col > len(parts) {
		return ""
	}
	return parts[col-1]
}

// Reverse reverses the line order.
func (le *LineEditor) Reverse() {
	for i, j := 0, len(le.lines)-1; i < j; i, j = i+1, j-1 {
		le.lines[i], le.lines[j] = le.lines[j], le.lines[i]
	}
	le.modified = true
}

// Shuffle randomizes the line order.
func (le *LineEditor) Shuffle() {
	rand.Shuffle(len(le.lines), func(i, j int) {
		le.lines[i], le.lines[j] = le.lines[j], le.lines[i]
	})
	le.modified = true
}

// Unique removes duplicate lines (keeps first occurrence).
func (le *LineEditor) Unique() {
	seen := make(map[string]bool)
	var result []string

	for _, line := range le.lines {
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}

	le.lines = result
	le.modified = true
}

// UniqueSorted sorts then removes duplicates.
func (le *LineEditor) UniqueSorted() {
	le.Sort()
	le.Unique()
}

// FindDupes returns duplicate lines with counts as a map.
func (le *LineEditor) FindDupes() map[string]int {
	counts := make(map[string]int)
	for _, line := range le.lines {
		counts[line]++
	}

	// Filter to only duplicates
	dupes := make(map[string]int)
	for line, count := range counts {
		if count > 1 {
			dupes[line] = count
		}
	}
	return dupes
}

// RemoveDupes removes all duplicate lines (keeps only unique lines).
func (le *LineEditor) RemoveDupes() {
	counts := make(map[string]int)
	for _, line := range le.lines {
		counts[line]++
	}

	var result []string
	for _, line := range le.lines {
		if counts[line] == 1 {
			result = append(result, line)
		}
	}

	le.lines = result
	le.modified = true
}

// KeepDupes keeps only duplicate lines (removes unique lines).
func (le *LineEditor) KeepDupes() {
	counts := make(map[string]int)
	for _, line := range le.lines {
		counts[line]++
	}

	var result []string
	for _, line := range le.lines {
		if counts[line] > 1 {
			result = append(result, line)
		}
	}

	le.lines = result
	le.modified = true
}

// ============================================================
// Text Processing Operations
// ============================================================

// Trim trims whitespace from each line.
func (le *LineEditor) Trim() {
	for i, line := range le.lines {
		le.lines[i] = strings.TrimSpace(line)
	}
	le.modified = true
}

// TrimLeft trims leading whitespace from each line.
func (le *LineEditor) TrimLeft() {
	for i, line := range le.lines {
		le.lines[i] = strings.TrimLeft(line, " \t")
	}
	le.modified = true
}

// TrimRight trims trailing whitespace from each line.
func (le *LineEditor) TrimRight() {
	for i, line := range le.lines {
		le.lines[i] = strings.TrimRight(line, " \t")
	}
	le.modified = true
}

// RemoveEmpty removes empty lines.
func (le *LineEditor) RemoveEmpty() {
	var result []string
	for _, line := range le.lines {
		if line != "" {
			result = append(result, line)
		}
	}
	le.lines = result
	le.modified = true
}

// RemoveBlank removes blank lines (whitespace only).
func (le *LineEditor) RemoveBlank() {
	var result []string
	for _, line := range le.lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	le.lines = result
	le.modified = true
}

// Dedent removes common indentation from all lines.
func (le *LineEditor) Dedent() {
	if len(le.lines) == 0 {
		return
	}

	// Find minimum indentation
	minIndent := -1
	for _, line := range le.lines {
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}
		indent := 0
		for _, ch := range line {
			if ch == ' ' || ch == '\t' {
				indent++
			} else {
				break
			}
		}
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return
	}

	// Remove common indentation
	for i, line := range le.lines {
		if len(line) >= minIndent {
			le.lines[i] = line[minIndent:]
		}
	}
	le.modified = true
}

// Indent adds prefix to each line.
func (le *LineEditor) Indent(prefix string) {
	for i, line := range le.lines {
		le.lines[i] = prefix + line
	}
	le.modified = true
}

// NumberLines adds line number prefix to each line.
func (le *LineEditor) NumberLines(start int) {
	for i, line := range le.lines {
		le.lines[i] = strconv.Itoa(start+i) + ": " + line
	}
	le.modified = true
}

// Join joins all lines into one with separator.
func (le *LineEditor) Join(sep string) string {
	return strings.Join(le.lines, sep)
}

// SplitLines splits each line by separator into multiple lines.
func (le *LineEditor) SplitLines(sep string) {
	var result []string
	for _, line := range le.lines {
		parts := strings.Split(line, sep)
		result = append(result, parts...)
	}
	le.lines = result
	le.modified = true
}

// Prefix adds prefix to each line.
func (le *LineEditor) Prefix(prefix string) {
	for i, line := range le.lines {
		le.lines[i] = prefix + line
	}
	le.modified = true
}

// Suffix adds suffix to each line.
func (le *LineEditor) Suffix(suffix string) {
	for i, line := range le.lines {
		le.lines[i] = line + suffix
	}
	le.modified = true
}

// ToUpperCase converts all lines to uppercase.
func (le *LineEditor) ToUpperCase() {
	for i, line := range le.lines {
		le.lines[i] = strings.ToUpper(line)
	}
	le.modified = true
}

// ToLowerCase converts all lines to lowercase.
func (le *LineEditor) ToLowerCase() {
	for i, line := range le.lines {
		le.lines[i] = strings.ToLower(line)
	}
	le.modified = true
}

// ============================================================
// Export and Save Operations
// ============================================================

// ToText returns the full text with newlines.
func (le *LineEditor) ToText() string {
	return strings.Join(le.lines, "\n")
}

// ToLines returns the line array.
func (le *LineEditor) ToLines() []string {
	result := make([]string, len(le.lines))
	copy(result, le.lines)
	return result
}

// Save saves to the original file.
func (le *LineEditor) Save() error {
	if le.filePath == "" {
		return errNoFilePath
	}
	return le.SaveAs(le.filePath)
}

// SaveAs saves to a new file.
func (le *LineEditor) SaveAs(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range le.lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	le.modified = false
	le.filePath = path
	return nil
}

// Close releases resources (no-op for LineEditor).
func (le *LineEditor) Close() {
	// No resources to release
}

// GetFilePath returns the associated file path.
func (le *LineEditor) GetFilePath() string {
	return le.filePath
}

// SetFilePath sets the associated file path.
func (le *LineEditor) SetFilePath(path string) {
	le.filePath = path
}

// ============================================================
// Statistics Methods
// ============================================================

// CharCount returns the total number of bytes in all lines.
func (le *LineEditor) CharCount() int {
	total := 0
	for _, line := range le.lines {
		total += len(line)
	}
	// Add newlines
	total += len(le.lines) // Each line has a newline when saved
	if len(le.lines) > 0 {
		total-- // Last line may not have newline
	}
	return total
}

// RuneCount returns the total number of runes (Unicode code points) in all lines.
func (le *LineEditor) RuneCount() int {
	total := 0
	for _, line := range le.lines {
		total += len([]rune(line))
	}
	return total
}

// WordCount returns the total number of words in all lines.
func (le *LineEditor) WordCount() int {
	total := 0
	for _, line := range le.lines {
		total += len(strings.Fields(line))
	}
	return total
}

// Info returns a map with statistics about the editor content.
func (le *LineEditor) Info() map[string]int {
	return map[string]int{
		"lineCount":  le.LineCount(),
		"charCount":  le.CharCount(),
		"runeCount":  le.RuneCount(),
		"wordCount":  le.WordCount(),
		"emptyLines": le.CountEmpty(),
	}
}

// CountEmpty returns the number of empty lines.
func (le *LineEditor) CountEmpty() int {
	count := 0
	for _, line := range le.lines {
		if line == "" {
			count++
		}
	}
	return count
}

// ============================================================
// File Operations
// ============================================================

// AppendToFile appends the content to an existing file.
func (le *LineEditor) AppendToFile(path string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range le.lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// AppendFromFile appends lines from a file to the editor.
func (le *LineEditor) AppendFromFile(path string) error {
	other, err := NewLineEditorFromFile(path)
	if err != nil {
		return err
	}

	le.lines = append(le.lines, other.lines...)
	le.modified = true
	return nil
}

// ============================================================
// View Methods (for printing/displaying)
// ============================================================

// ViewLine returns a formatted string for the specified line (with line number).
func (le *LineEditor) ViewLine(n int, showNumber bool) string {
	line, ok := le.GetLine(n)
	if !ok {
		return ""
	}
	if showNumber {
		return strconv.Itoa(n) + ": " + line
	}
	return line
}

// ViewLines returns formatted strings for a range of lines.
func (le *LineEditor) ViewLines(start, end int, showNumber bool) []string {
	lines := le.GetLines(start, end)
	if lines == nil {
		return nil
	}

	result := make([]string, len(lines))
	for i, line := range lines {
		if showNumber {
			result[i] = strconv.Itoa(start+i) + ": " + line
		} else {
			result[i] = line
		}
	}
	return result
}

// ViewAll returns all lines formatted (with optional line numbers).
func (le *LineEditor) ViewAll(showNumber bool) []string {
	result := make([]string, len(le.lines))
	for i, line := range le.lines {
		if showNumber {
			result[i] = strconv.Itoa(i+1) + ": " + line
		} else {
			result[i] = line
		}
	}
	return result
}

// ============================================================
// Line Operations (Additional)
// ============================================================

// SwapLines swaps two lines at the given indices (1-based).
func (le *LineEditor) SwapLines(n1, n2 int) bool {
	idx1 := le.normalizeIndex(n1)
	idx2 := le.normalizeIndex(n2)

	if idx1 < 0 || idx2 < 0 {
		return false
	}

	le.lines[idx1], le.lines[idx2] = le.lines[idx2], le.lines[idx1]
	le.modified = true
	return true
}

// MoveLine moves a line from one position to another (1-based).
func (le *LineEditor) MoveLine(from, to int) bool {
	lineCount := len(le.lines)
	if lineCount == 0 {
		return false
	}

	// Normalize indices for insertion
	fromIdx := le.normalizeIndex(from)
	if fromIdx < 0 {
		return false
	}

	// Handle 'to' position (can be after the last line)
	toIdx := to - 1
	if toIdx < 0 {
		toIdx = lineCount + to // Handle negative
	}
	if toIdx < 0 {
		toIdx = 0
	}
	if toIdx > lineCount {
		toIdx = lineCount
	}

	// Get the line to move
	line := le.lines[fromIdx]

	// Remove from original position
	le.lines = append(le.lines[:fromIdx], le.lines[fromIdx+1:]...)

	// Adjust toIdx if needed
	if toIdx > fromIdx {
		toIdx--
	}

	// Insert at new position
	le.lines = append(le.lines[:toIdx], append([]string{line}, le.lines[toIdx:]...)...)
	le.modified = true
	return true
}

// DuplicateLine duplicates a line at the given index (1-based).
func (le *LineEditor) DuplicateLine(n int) bool {
	idx := le.normalizeIndex(n)
	if idx < 0 {
		return false
	}

	line := le.lines[idx]
	le.lines = append(le.lines[:idx+1], append([]string{line}, le.lines[idx+1:]...)...)
	le.modified = true
	return true
}

// TakeLines extracts and removes a range of lines, returning them.
func (le *LineEditor) TakeLines(start, end int) []string {
	lines := le.GetLines(start, end)
	if lines == nil {
		return nil
	}

	le.DeleteLines(start, end)
	return lines
}

// ============================================================
// Text Processing (Additional)
// ============================================================

// PadRight pads each line to the specified width with the given character.
func (le *LineEditor) PadRight(width int, padChar string) {
	for i, line := range le.lines {
		runes := []rune(line)
		if len(runes) < width {
			padding := strings.Repeat(padChar, width-len(runes))
			le.lines[i] = line + padding
		}
	}
	le.modified = true
}

// PadLeft pads each line on the left to the specified width with the given character.
func (le *LineEditor) PadLeft(width int, padChar string) {
	for i, line := range le.lines {
		runes := []rune(line)
		if len(runes) < width {
			padding := strings.Repeat(padChar, width-len(runes))
			le.lines[i] = padding + line
		}
	}
	le.modified = true
}

// Truncate truncates each line to the specified maximum length.
func (le *LineEditor) Truncate(maxLen int) {
	for i, line := range le.lines {
		runes := []rune(line)
		if len(runes) > maxLen {
			le.lines[i] = string(runes[:maxLen])
		}
	}
	le.modified = true
}

// TruncateWithEllipsis truncates each line and adds "..." if truncated.
func (le *LineEditor) TruncateWithEllipsis(maxLen int) {
	ellipsis := "..."
	ellipsisLen := len([]rune(ellipsis))
	for i, line := range le.lines {
		runes := []rune(line)
		if len(runes) > maxLen {
			if maxLen > ellipsisLen {
				le.lines[i] = string(runes[:maxLen-ellipsisLen]) + ellipsis
			} else {
				le.lines[i] = string(runes[:maxLen])
			}
		}
	}
	le.modified = true
}

// AlignLeft aligns all lines to the left by trimming and optionally padding.
func (le *LineEditor) AlignLeft(width int) {
	le.TrimLeft()
	if width > 0 {
		le.PadRight(width, " ")
	}
}

// AlignRight aligns all lines to the right by padding.
func (le *LineEditor) AlignRight(width int) {
	for i, line := range le.lines {
		runes := []rune(line)
		if len(runes) < width {
			padding := strings.Repeat(" ", width-len(runes))
			le.lines[i] = padding + line
		}
	}
	le.modified = true
}

// Center centers each line within the specified width.
func (le *LineEditor) Center(width int) {
	for i, line := range le.lines {
		runes := []rune(line)
		if len(runes) < width {
			leftPad := (width - len(runes)) / 2
			rightPad := width - len(runes) - leftPad
			le.lines[i] = strings.Repeat(" ", leftPad) + line + strings.Repeat(" ", rightPad)
		}
	}
	le.modified = true
}

// StripPrefix removes a prefix from each line if present.
func (le *LineEditor) StripPrefix(prefix string) {
	for i, line := range le.lines {
		le.lines[i] = strings.TrimPrefix(line, prefix)
	}
	le.modified = true
}

// StripSuffix removes a suffix from each line if present.
func (le *LineEditor) StripSuffix(suffix string) {
	for i, line := range le.lines {
		le.lines[i] = strings.TrimSuffix(line, suffix)
	}
	le.modified = true
}

// Comment adds a comment prefix to each line.
func (le *LineEditor) Comment(prefix string) {
	le.Prefix(prefix + " ")
}

// Uncomment removes a comment prefix from each line.
func (le *LineEditor) Uncomment(prefix string) {
	le.StripPrefix(prefix + " ")
	le.StripPrefix(prefix)
}

// SelectLines returns a new LineEditor with only the lines at specified indices.
func (le *LineEditor) SelectLines(indices []int) *LineEditor {
	result := NewLineEditor()
	for _, idx := range indices {
		if line, ok := le.GetLine(idx); ok {
			result.AddLine(line)
		}
	}
	return result
}

// Sample returns a new LineEditor with n randomly selected lines.
func (le *LineEditor) Sample(n int) *LineEditor {
	if n <= 0 || len(le.lines) == 0 {
		return NewLineEditor()
	}

	// Copy and shuffle
	shuffled := make([]string, len(le.lines))
	copy(shuffled, le.lines)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Take first n
	if n > len(shuffled) {
		n = len(shuffled)
	}
	return NewLineEditorFromLines(shuffled[:n])
}

// Head returns the first n lines.
func (le *LineEditor) Head(n int) []string {
	if n <= 0 {
		return nil
	}
	if n > len(le.lines) {
		n = len(le.lines)
	}
	return le.GetLines(1, n)
}

// Tail returns the last n lines.
func (le *LineEditor) Tail(n int) []string {
	if n <= 0 {
		return nil
	}
	lineCount := len(le.lines)
	if n > lineCount {
		n = lineCount
	}
	return le.GetLines(lineCount-n+1, lineCount)
}