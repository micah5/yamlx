package yamlx

import (
	"fmt"
	"strings"
)

type Type int

const (
	Unknown LineType = iota
	KeyValue
	ListElement
)

type Meta struct {
	Indent int
	Type   Type
	Anchor string
}

func getMeta(line string) (Meta, error) {
	m := Meta{}

	// Determine indent
	for _, char := range line {
		if string(char) == "\t" {
			m.Indent++
		} else {
			break
		}
	}

	// Trim line
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
		m.Anchor := trimmed[start+1 : end]
		trimmed = strings.Replace(trimmed, " &"+m.Anchor, "", 1)
	}

	// Determine line type
	switch {
	case strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, " "):
		m.Type = Unknown
	case strings.HasPrefix(trimmed, "- "):
		m.Type = ListElement
	case strings.Contains(trimmed, ":"):
		m.Type = KeyValue
	default:
		return nil, fmt.Errorf("unknown line type: %s", line)
	}

	return m, nil
}

func parseLines(lines []string) {
	for i, line := range lines {
		meta, err := getMeta(line)
		if err != nil {
			return nil, err
		}

		var value any
		if meta.Type == Unknown {
			// check the type of the next line
			if len(lines) <= i+1 {
				return nil, fmt.Errorf("unexpected end of file")
			}
			nextMeta, err := getMeta(lines[i+1])
			switch nextMeta.Type {
			case KeyValue:
				value = make(map[string]any)
			case ListElement:
				value = make([]any)
			case Unknown:

	}
}

func Parse(data string) {
	lines := strings.Split(data, "\n")

	// Remove empty lines
	for i, line := range lines {
		if line == "" {
			lines = append(lines[:i], lines[i+1:]...)
		}
	}

	// Parse each line
}
