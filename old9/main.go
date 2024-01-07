package yamlx

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Type int

const (
	Primitive Type = iota
	KeyValue
	ListElement
	List
	Map
)

// ListElement represents a YAML list element.
type Part struct {
	Type  Type
	Key   string
	Value any
}

func NewPart(line string) (*Part, error) {
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "- ") { // List element
		value := strings.TrimPrefix(trimmed, "- ")
		return &Part{Type: ListElement, Value: NewPart(value)}, nil
	}

	if strings.Contains(trimmed, ":") { // Key-value pair
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format: %s", line)
		}
		// Check if value is empty
		key := strings.TrimSpace(parts[0])
		if strings.TrimSpace(parts[1]) == "" {
			// Map
			return &Part{Type: Map, Key: key}, nil
		}
		return &Part{Type: KeyValue, Key: key, Value: NewPart(parts[1])}, nil
	}

	// Try parsing as int, float, bool, or return as string
	if i, err := strconv.ParseInt(p.Value, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(p.Value, 64); err == nil {
		return f
	}
	if b, err := strconv.ParseBool(p.Value); err == nil {
		return b
	}

	return trimmed
}

func (p Part) Process() (any, error) {
	// Handle collections
	if p.Type == List || p.Type == Map {
		for _, part := range p.Value {
			if (p.Type == List && part.Type != ListElement) || (p.Type == Map && part.Type != KeyValue) {
				return nil, fmt.Errorf("invalid format: %s", p.Value)
			}
			part.Parse()
		}
	}

	// Handle generic types

}

// *** Parsing ***
// Parse parses the given YAML data into JSON format.
func Parse(data string) {
	lines := strings.Split(data, "\n")
	processedLines := processLines(lines, 0, make(map[string]any))

	jsonString, err := json.Marshal(processedLines)
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
		return
	}
	fmt.Println(string(jsonString))
}

// processLines recursively processes YAML lines into structured data.
func processLines(lines []string, currentLevel int, anchors map[string]any) []any {
	if len(lines) == 0 {
		return nil
	}

	var values []any
	for i, line := range lines {
		lineLevel := strings.Count(line, "\t")

		if line == "" { // Skip empty lines
			continue
		}

		anchor, newLine := extractAnchor(line, "&")
		line = newLine

		if strings.Contains(line, "#") { // Remove comments
			line = strings.Split(line, "#")[0]
		}

		if lineLevel == currentLevel {
			if i < len(lines)-1 && strings.Count(lines[i+1], "\t") > lineLevel {
				// Process nested lines
				end := findEndOfBlock(lines, i+1, lineLevel)
				nestedLines := processLines(lines[i+1:end], lineLevel+1, anchors)
				i = end - 1 // Skip processed lines
			} else {
				// Process key/value pair
			}
		}
	}
	return values
}

// *** Utils ***
// findEndOfBlock finds the end index of a block at a given indentation level.
func findEndOfBlock(lines []string, start, level int) int {
	for i := start; i < len(lines); i++ {
		if strings.Count(lines[i], "\t") <= level {
			return i
		}
	}
	return len(lines)
}
