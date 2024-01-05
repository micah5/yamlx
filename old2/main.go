package yamlx

import (
	"fmt"
	"strings"
)

type LineType int

const (
	LineTypeKeyValue LineType = iota
	LineTypeKey
	LineTypeListElement
)

type Anchor struct {
	Name  string
	Lines []*Line
}

type Line struct {
	Type            LineType
	Indent          int
	Value	   any
	Anchor          *Anchor
}

// NewLine creates and initializes a new Line based on the input string.
func NewLine(line string) (*Line, error) {
	l := &Line{}
	if err := l.parseIndent(line); err != nil {
		return nil, err
	}
	if err := l.parseType(line); err != nil {
		return nil, err
	}
	return l, nil
}

// parseIndent parses the indentation of a line.
func (l *Line) parseIndent(line string) error {
	for _, char := range line {
		if string(char) == "\t" {
			l.Indent++
		} else {
			break
		}
	}
	return nil
}

// parseType determines the type of the line and parses it accordingly.
func (l *Line) parseType(line string) error {
	trimmed := strings.TrimSpace(line)

	// Check for anchor in the line
	start := strings.Index(trimmed, "&")
	if start != -1 {
		end := strings.Index(trimmed[start:], " ")
		if end == -1 {
			end = len(trimmed)
		} else {
			end += start
		}
		anchorName := trimmed[start+1 : end]
		l.Anchor = &Anchor{Name: anchorName, Lines: []*Line{l}}
		trimmed = strings.Replace(trimmed, " &"+anchorName, "", 1)
	}

	// Determine line type and parse accordingly
	switch {
	case strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, " "):
		l.Type = LineTypeKey
		trimmed = strings.TrimSuffix(trimmed, ":")
	case strings.HasPrefix(trimmed, "- "):
		l.Type = LineTypeListElement
		trimmed = strings.TrimPrefix(trimmed, "- ")
	case strings.Contains(trimmed, ":"):
		l.Type = LineTypeKeyValue
	default:
		return fmt.Errorf("unknown line type: %s", line)
	}

	return nil
}

func (l *Line) ParseValue() any {
	var val string
	switch l.Type {
	case LineTypeKeyValue:
		val = strings.Split(l.ProcessedString, ":")[1]
	case LineTypeListElement:
		val = l.ProcessedString
	}

	if val != "" {
		if intValue, err := strconv.ParseInt(val, 10, 0); err == nil {
			return int(intValue)
		}
		if floatValue, err := strconv.ParseFloat(val, 64); err == nil {
			return floatValue
		}
		if boolValue, err := strconv.ParseBool(val); err == nil {
			return boolValue
		}
	}
}

func Parse(data string) ([]*Line, error) {
	var lines []*Line
	rawLines := strings.Split(data, "\n")

	// Parse each line
	anchors := make(map[string]*Anchor)
	for _, rawLine := range rawLines {
		if rawLine == "" {
			continue
		}

		// Add lines
		parsedLine, err := NewLine(rawLine)
		if err != nil {
			return nil, err
		}
		lines = append(lines, parsedLine)

		// Add anchors
		if parsedLine.Anchor != nil {
			anchors[parsedLine.Anchor.Name] = parsedLine.Anchor
		}
	}

	for _, line := range lines {
		if line.Type == LineTypeListElement {


	return lines, nil
}
