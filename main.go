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

func NewLine(line, indentationType string) *Line {
	indent := getIndent(line, indentationType)
	trimmed := strings.TrimSpace(line)

	var lineType LineType
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

func getIndent(s, indentationType string) int {
	indent := 0
	for _, char := range s {
		if string(char) == indentationType {
			indent++
		} else {
			break
		}
	}
	return indent
}

type Lines []*Line

func getIndentationType(rawLines []string) (string, error) {
	// find the type of indentation used
	var indentationType string
	for _, rawLine := range rawLines {
		if strings.Contains(rawLine, "\t") {
			indentationType = "\t"
			break
		}
		// check for spaces
		for i := 1; i < 8; i++ {
			_indentationType := strings.Repeat(" ", i)
			if strings.HasPrefix(rawLine, _indentationType) {
				indentationType = _indentationType
				break
			}
		}
	}

	// check that all lines use the same indentation type
	for _, rawLine := range rawLines {
		if !strings.HasPrefix(rawLine, indentationType) {
			return "", fmt.Errorf("inconsistent indentation")
		}
	}
	return indentationType, nil
}

func Parse(data string) (any, error) {
	lines := make(Lines, 0)
	rawLines := strings.Split(data, "\n")

	// find the type of indentation used
	indentationType, err := getIndentationType(rawLines)
	if err != nil {
		return nil, err
	}

	// parse the lines
	for _, rawLine := range rawLines {
		if rawLine == "" {
			continue
		}

		parsedLine := NewLine(rawLine, indentationType)
		if parsedLine.Type == LineTypeUnknown {
			return nil, fmt.Errorf("invalid line: %s", rawLine)
		}
		lines = append(lines, parsedLine)
	}

	// TODO: Implement further logic to parse the lines into a structured YAML representation
	return lines, nil
}
