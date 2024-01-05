package yamlx

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type ListElement struct {
	Value any
}

func NewListElement(value any) ListElement {
	return ListElement{value}
}

type Pair struct {
	Key   string
	Value any
}

func NewPair(key string, value any) Pair {
	return Pair{key, value}
}

func Parse(data string) {
	lines := strings.Split(data, "\n")
	l := processLines(lines, 0)
	jsonString, err := json.Marshal(l)
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
		return
	}
	fmt.Println(string(jsonString))
}

// Recursive function to process lines
func processLines(lines []string, currentLevel int) []any {
	if len(lines) == 0 {
		return nil
	}

	values := make([]any, 0)
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		lineLevel := strings.Count(line, "\t")

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check if line is at the current level
		if lineLevel == currentLevel {
			if i < len(lines)-1 && strings.Count(lines[i+1], "\t") > lineLevel {
				// This line is a map, call recursively for nested lines
				end := findEndOfBlock(lines, i+1, lineLevel)
				fmt.Println("(map)", line)
				l := processLines(lines[i+1:end], lineLevel+1)
				if len(l) > 0 {
					key := strings.TrimSuffix(line, ":")
					key = strings.TrimSpace(key)
					println("woo00", key)
					switch reflect.TypeOf(l[0]) {
					case reflect.TypeOf(ListElement{}):
						println("huh")
						value := make([]any, 0)
						for _, v := range l {
							value = append(value, v.(ListElement).Value)
						}
						values = append(values, Pair{key, value})
					case reflect.TypeOf(Pair{}):
						println("hiiii")
						value := make(map[string]any)
						for _, v := range l {
							fmt.Println("v", v)
							p := v.(Pair)
							value[p.Key] = p.Value
						}
						values = append(values, Pair{key, value})
					}
				}
				i = end - 1 // Skip processed lines
			} else {
				// This line is a key/value pair
				value := processValue(line)
				fmt.Println("(k/v)", value)
				values = append(values, value)
			}
		}
	}
	return values
}

func processValue(line string) any {
	trimmed := strings.TrimSpace(line)

	// Check for list element
	if strings.HasPrefix(trimmed, "- ") {
		value := strings.TrimPrefix(trimmed, "- ")
		return NewListElement(processValue(value))
	}

	// Check for key-value pair
	if strings.Contains(trimmed, ":") {
		split := strings.SplitN(trimmed, ":", 2)
		return NewPair(strings.TrimSpace(split[0]), processValue(strings.TrimSpace(split[1])))
	}

	// Attempt to parse as int, float, bool, or return as string
	if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return f
	}
	if b, err := strconv.ParseBool(trimmed); err == nil {
		return b
	}
	return trimmed
}

// Helper function to find the end of the current block
func findEndOfBlock(lines []string, start int, level int) int {
	for i := start; i < len(lines); i++ {
		if strings.Count(lines[i], "\t") <= level {
			return i
		}
	}
	return len(lines)
}
