package yamlx

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type LineType int

const (
	LineTypeUnknown LineType = iota
	LineTypeKeyValue
	LineTypeList
	LineTypeMap
)

func determineLineTypeAndIndent(line string) (int, LineType) {
	indent := strings.Cont(line, "  ")
	trimmed := strings.TrimSpace(line)

	if strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, " ") {
		return indent, LineTypeMap // Line is a key for a nested map
	} else if strings.HasPrefix(trimmed, "- ") {
		return indent, LineTypeList
	} else if strings.Contains(trimmed, ":") {
		return indent, LineTypeKeyValue
	} else {
		return indent, LineTypeUnknown
	}
}

// ParseValue attempts to convert a string value to a more specific type: int, float, bool, or string.
func parseValue(value string) any {
	if intValue, err := strconv.ParseInt(value, 10, 0); err == nil {
		return int(intValue) // Convert to int
	}
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}
	if boolValue, err := strconv.ParseBool(value); err == nil {
		return boolValue
	}
	if mapValue, err := parseMap([]string{value}, 0); err == nil {
		return mapValue
	}
	return value
}

// ParseLine parses a single line of YAML-like syntax into a key and a dynamically typed value.
func parseLine(line string) (key string, value any, err error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", nil, errors.New("invalid format")
	}
	return strings.TrimSpace(parts[0]), parseValue(strings.TrimSpace(parts[1])), nil
}

func parseList(lines []string, currentIndent int) ([]any, error) {
	var list []any
	i := 0

	for i < len(lines) {
		line := lines[i]
		indent, lineType := determineLineTypeAndIndent(line)

		if indent < currentIndent {
			// This line is less indented than the current list level, so it's not part of the list
			break
		} else if indent > currentIndent {
			// This line is more indented, so it must be part of a nested structure in the previous list item
			if len(list) == 0 {
				return nil, fmt.Errorf("invalid indentation at line: %s", line)
			}

			// Determine if the last item in the list is a map or a list and parse accordingly
			switch lastItem := list[len(list)-1].(type) {
			case map[string]any:
				// Parse nested map
				nestedMap, err := parseMap(lines[i:], indent)
				if err != nil {
					return nil, err
				}
				for k, v := range nestedMap {
					lastItem[k] = v
				}
				list[len(list)-1] = lastItem
			case []any:
				// Parse nested list
				nestedList, err := parseList(lines[i:], indent)
				if err != nil {
					return nil, err
				}
				list[len(list)-1] = append(lastItem, nestedList...)
			default:
				return nil, fmt.Errorf("unexpected type in list: %s", line)
			}

			// Skip lines that have been processed as part of the nested structure
			for i < len(lines) && strings.Count(lines[i], "  ") >= indent {
				i++
			}
			continue
		}

		if lineType == LineTypeList {
			trimmedLine := strings.TrimPrefix(strings.TrimSpace(line), "- ")
			itemValue := parseValue(trimmedLine)
			list = append(list, itemValue)
		} else if lineType == LineTypeMap {
			// Start of a new map within the list
			nestedMap, err := parseMap(lines[i:], indent+1)
			if err != nil {
				return nil, err
			}
			list = append(list, nestedMap)
			for i < len(lines) && strings.Count(lines[i], "  ") > indent {
				i++
			}
			continue
		} else {
			return nil, fmt.Errorf("invalid list item type: %s", line)
		}

		i++
	}

	return list, nil
}

func parseMap(lines []string, currentIndent int) (map[string]any, error) {
	m := make(map[string]any)
	i := 0

	for i < len(lines) {
		line := lines[i]
		indent, lineType := determineLineTypeAndIndent(line)

		if indent < currentIndent {
			// This line is less indented than the current map level, so it's not part of the map
			break
		} else if indent > currentIndent {
			// This line is more indented, so it must be part of a nested structure in the last map item
			if len(m) == 0 {
				return nil, fmt.Errorf("invalid indentation at line: %s", line)
			}

			// Handle nested structure in the last map item
			lastKey := getLastKey(m)
			switch lastItem := m[lastKey].(type) {
			case map[string]any:
				// Parse nested map
				nestedMap, err := parseMap(lines[i:], indent)
				if err != nil {
					return nil, err
				}
				for k, v := range nestedMap {
					lastItem[k] = v
				}
				m[lastKey] = lastItem
			case []any:
				// Parse nested list
				nestedList, err := parseList(lines[i:], indent)
				if err != nil {
					return nil, err
				}
				m[lastKey] = append(lastItem, nestedList...)
			default:
				// Assign a nested map or list if the last item was not already one
				if _, isMap := lastItem.(map[string]any); !isMap {
					m[lastKey] = map[string]any{}
				}
				if _, isList := lastItem.([]any); !isList {
					m[lastKey] = []any{}
				}
			}

			// Skip lines that have been processed as part of the nested structure
			for i < len(lines) && strings.Count(lines[i], "  ") >= indent {
				i++
			}
			continue
		}

		if lineType == LineTypeKeyValue {
			key, value, err := parseLine(line)
			if err != nil {
				return nil, err
			}
			m[key] = value
		} else if lineType == LineTypeList || lineType == LineTypeMap {
			// Prepare for a nested list or map
			key, _, err := parseLine(line)
			if err != nil {
				return nil, err
			}
			m[key] = nil
		} else {
			return nil, fmt.Errorf("invalid map item type: %s", line)
		}

		i++
	}

	return m, nil
}

// getLastKey returns the last key added to the map. This is used when handling nested structures.
func getLastKey(m map[string]any) string {
	var lastKey string
	for k := range m {
		lastKey = k
	}
	return lastKey
}

/*
func parseNested(lines []string, currentIndentation int) (map[string]any, error) {
	result := make(map[string]any)
	var currentKey string
	var nestedLines []string
	var isNested, isList bool

	for _, line := range lines {
		indent, lineType := determineLineTypeAndIndent(line)

		// Check for the start of a nested structure or a list
		if indent == currentIndentation {
			if isNested {
				// Process the nested structure
				if isList {
					result[currentKey], _ = parseList(nestedLines, currentIndentation+1)
				} else {
					result[currentKey], _ = parseNested(nestedLines, currentIndentation+1)
				}
				nestedLines = nil
				isNested = false
			}
			if lineType == LineTypeMap {
				currentKey = strings.TrimSpace(strings.TrimSuffix(line, ":"))
				isNested = true
			} else if lineType == LineTypeKeyValue {
				key, value, _ := parseLine(line)
				result[key] = value
			}
		}

		if isNested && indent > currentIndentation {
			nestedLines = append(nestedLines, line)
		} else if lineType == LineTypeKeyValue && indent == currentIndentation {
			key, value, _ := parseLine(line)
			result[key] = value
		}
	}

	// Final processing for the last nested structure
	if isNested {
		if isList {
			fmt.Println("nested list", nestedLines)
			result[currentKey], _ = parseList(nestedLines, currentIndentation+1)
		} else {
			result[currentKey], _ = parseNested(nestedLines, currentIndentation+1)
		}
	}

	return result, nil
}
*/
