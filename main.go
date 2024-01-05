package yamlx

import (
	"fmt"
	"strings"
)

type LineType int

const (
	LineTypeUnknown LineType = iota
	LineTypeKeyValue
	LineTypeKey
	LineTypeListElement
)

type Line struct {
	Type            LineType
	Indent          int
	ProcessedString string
}

func NewLine(line string) *Line {
	// get the indent
	indent := 0
	for _, char := range line {
		if string(char) == "\t" {
			indent++
		} else {
			break
		}
	}

	// determine the line type
	var lineType LineType
	trimmed := strings.TrimSpace(line)
	if strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, " ") {
		lineType = LineTypeKey
	} else if strings.HasPrefix(trimmed, "- ") {
		lineType = LineTypeListElement
		trimmed = strings.TrimPrefix(trimmed, "- ")
	} else if strings.Contains(trimmed, ":") {
		lineType = LineTypeKeyValue
	} else {
		lineType = LineTypeUnknown
	}

	return &Line{
		Type:            lineType,
		Indent:          indent,
		ProcessedString: trimmed,
	}
}

type Lines []*Line

func Parse(data string) (any, error) {
	lines := make(Lines, 0)
	rawLines := strings.Split(data, "\n")

	// parse the lines
	for _, rawLine := range rawLines {
		if rawLine == "" {
			continue
		}

		parsedLine := NewLine(rawLine)
		if parsedLine.Type == LineTypeUnknown {
			return nil, fmt.Errorf("invalid line: %s", rawLine)
		}
		lines = append(lines, parsedLine)
	}

	// TODO: Implement further logic to parse the lines into a structured YAML representation
	return lines, nil
}
