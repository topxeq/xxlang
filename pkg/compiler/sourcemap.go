// pkg/compiler/sourcemap.go
// Source map for mapping bytecode positions to source locations
package compiler

import (
	"encoding/binary"
	"fmt"
	"io"
)

// SourceLocation represents a position in source code
type SourceLocation struct {
	Line   int
	Column int
}

// SourceMap maps bytecode instruction positions to source locations
type SourceMap struct {
	// Maps instruction offset -> source location
	Locations map[int]SourceLocation
	// Source file path (if available)
	SourceFile string
	// Source code lines (for error display)
	SourceLines []string
}

// NewSourceMap creates a new empty source map
func NewSourceMap() *SourceMap {
	return &SourceMap{
		Locations: make(map[int]SourceLocation),
	}
}

// Add maps an instruction offset to a source location
func (sm *SourceMap) Add(offset int, loc SourceLocation) {
	sm.Locations[offset] = loc
}

// Get returns the source location for an instruction offset
// Returns the closest location if exact match not found
func (sm *SourceMap) Get(offset int) (SourceLocation, bool) {
	// Try exact match first
	if loc, ok := sm.Locations[offset]; ok {
		return loc, true
	}

	// Find closest location before this offset
	var closest SourceLocation
	var found bool
	var closestOffset int = -1

	for off, loc := range sm.Locations {
		if off <= offset && (closestOffset == -1 || off > closestOffset) {
			closest = loc
			closestOffset = off
			found = true
		}
	}

	return closest, found
}

// SetSourceFile sets the source file path and loads source lines
func (sm *SourceMap) SetSourceFile(path string, source string) {
	sm.SourceFile = path
	sm.SourceLines = splitLines(source)
}

// splitLines splits source code into lines
func splitLines(source string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(source); i++ {
		if source[i] == '\n' {
			lines = append(lines, source[start:i])
			start = i + 1
		}
	}
	if start < len(source) {
		lines = append(lines, source[start:])
	}
	return lines
}

// GetLine returns a specific source line (1-indexed)
func (sm *SourceMap) GetLine(lineNum int) string {
	if lineNum < 1 || lineNum > len(sm.SourceLines) {
		return ""
	}
	return sm.SourceLines[lineNum-1]
}

// FormatError creates a formatted error message with source context
func (sm *SourceMap) FormatError(offset int, message string) string {
	loc, ok := sm.Get(offset)
	if !ok {
		return message
	}

	var result string
	if sm.SourceFile != "" {
		result = fmt.Sprintf("%s:%d:%d: %s", sm.SourceFile, loc.Line, loc.Column, message)
	} else {
		result = fmt.Sprintf("line %d:%d: %s", loc.Line, loc.Column, message)
	}

	// Add source context
	if line := sm.GetLine(loc.Line); line != "" {
		result += fmt.Sprintf("\n\n  %d | %s", loc.Line, line)

		// Add pointer to the specific column
		if loc.Column > 0 {
			pointer := ""
			for i := 0; i < loc.Column-1 && i < len(line); i++ {
				if line[i] == '\t' {
					pointer += "\t"
				} else {
					pointer += " "
				}
			}
			pointer += "^"
			result += fmt.Sprintf("\n      %s", pointer)
		}
		result += "\n"
	}

	return result
}

// Serialize writes the source map to a binary format
func (sm *SourceMap) Serialize(w io.Writer) error {
	// Write source file
	if err := binary.Write(w, binary.BigEndian, uint32(len(sm.SourceFile))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, sm.SourceFile); err != nil {
		return err
	}

	// Write number of source lines
	if err := binary.Write(w, binary.BigEndian, uint32(len(sm.SourceLines))); err != nil {
		return err
	}

	// Write each source line
	for _, line := range sm.SourceLines {
		if err := binary.Write(w, binary.BigEndian, uint32(len(line))); err != nil {
			return err
		}
		if _, err := io.WriteString(w, line); err != nil {
			return err
		}
	}

	// Write number of location mappings
	if err := binary.Write(w, binary.BigEndian, uint32(len(sm.Locations))); err != nil {
		return err
	}

	// Write each mapping
	for offset, loc := range sm.Locations {
		if err := binary.Write(w, binary.BigEndian, uint32(offset)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, uint32(loc.Line)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, uint32(loc.Column)); err != nil {
			return err
		}
	}

	return nil
}

// Deserialize reads the source map from a binary format
func DeserializeSourceMap(r io.Reader) (*SourceMap, error) {
	sm := NewSourceMap()

	// Read source file
	var fileLen uint32
	if err := binary.Read(r, binary.BigEndian, &fileLen); err != nil {
		return nil, err
	}
	fileBytes := make([]byte, fileLen)
	if _, err := io.ReadFull(r, fileBytes); err != nil {
		return nil, err
	}
	sm.SourceFile = string(fileBytes)

	// Read number of source lines
	var numLines uint32
	if err := binary.Read(r, binary.BigEndian, &numLines); err != nil {
		return nil, err
	}

	// Read each source line
	sm.SourceLines = make([]string, numLines)
	for i := uint32(0); i < numLines; i++ {
		var lineLen uint32
		if err := binary.Read(r, binary.BigEndian, &lineLen); err != nil {
			return nil, err
		}
		lineBytes := make([]byte, lineLen)
		if _, err := io.ReadFull(r, lineBytes); err != nil {
			return nil, err
		}
		sm.SourceLines[i] = string(lineBytes)
	}

	// Read number of location mappings
	var numMappings uint32
	if err := binary.Read(r, binary.BigEndian, &numMappings); err != nil {
		return nil, err
	}

	// Read each mapping
	for i := uint32(0); i < numMappings; i++ {
		var offset, line, column uint32
		if err := binary.Read(r, binary.BigEndian, &offset); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.BigEndian, &line); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.BigEndian, &column); err != nil {
			return nil, err
		}
		sm.Locations[int(offset)] = SourceLocation{Line: int(line), Column: int(column)}
	}

	return sm, nil
}
