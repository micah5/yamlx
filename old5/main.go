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

func (p Pair) ToMap() map[string]any {
	return map[string]any{p.Key: p.Value}
}

func Parse(data string) {
	lines := strings.Split(data, "\n")
	l := processLines(lines, 0, make(map[string]any))
	l2 := processResults(l, "").(Pair).Value
	jsonString, err := json.Marshal(l2)
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
		return
	}
	fmt.Println(string(jsonString))
}

// Recursive function to process lines
func processLines(lines []string, currentLevel int, anchors map[string]any) []any {
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

		// Extract anchors
		anchor, newLine := extractAnchor(line, "&")
		line = newLine

		// Remove comments
		if strings.Contains(line, "#") {
			line = strings.Split(line, "#")[0]
		}

		// Check if line is at the current level
		if lineLevel == currentLevel {
			println(line)
			if i < len(lines)-1 && strings.Count(lines[i+1], "\t") > lineLevel {
				println("in1")
				// This line is a map, call recursively for nested lines
				end := findEndOfBlock(lines, i+1, lineLevel)
				l := processLines(lines[i+1:end], lineLevel+1, anchors)
				fmt.Println("got back", l)
				processedL := processResults(l, line)
				values = addValue(values, processedL, anchor, anchors)
				i = end - 1 // Skip processed lines
			} else {
				println("in2")
				// This line is a key/value pair
				value := processValue(line)
				anchor2, _ := extractAnchor(line, "*")
				values = handleAnchor(values, value, anchor, anchor2, anchors)
				fmt.Println("added", value)
			}
		}
	}
	return values
}

func addValue(values []any, value any, anchor string, anchors map[string]any) []any {
	values = append(values, value)
	if anchor != "" {
		anchors[anchor] = value
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
		fmt.Println("This is a bool ->", trimmed, b)
		return b
	}
	return trimmed
}

func processResults(l []any, line string) any {
	if len(l) > 0 {
		key := strings.TrimSuffix(line, ":")
		key = strings.TrimSpace(key)
		fmt.Println("KEY", key, reflect.TypeOf(l[0]))
		switch reflect.TypeOf(l[0]) {
		case reflect.TypeOf(ListElement{}):
			value := make([]any, 0)
			var curMap map[string]any
			creatingMap := false
			for _, v := range l {
				if creatingMap {
					if reflect.TypeOf(v) == reflect.TypeOf(Pair{}) {
						// Add to map
						p := v.(Pair)
						curMap[p.Key] = p.Value
					} else {
						// End of map
						value = append(value, curMap)
						creatingMap = false
					}
				}
				if !creatingMap {
					fmt.Println("HELLO", v, reflect.TypeOf(v))
					listElementValue := v.(ListElement).Value
					if reflect.TypeOf(listElementValue) == reflect.TypeOf(Pair{}) {
						curMap = listElementValue.(Pair).ToMap()
						creatingMap = true
					} else {
						fmt.Println("adding", listElementValue)
						value = append(value, listElementValue)
					}
				}
			}
			if creatingMap {
				value = append(value, curMap)
			}
			if strings.HasPrefix(key, "- ") {
				cleanKey := strings.TrimPrefix(key, "- ")
				return NewListElement(NewPair(cleanKey, value))
			}
			return NewPair(key, value)
		case reflect.TypeOf(Pair{}):
			value := make(map[string]any)
			for _, v := range l {
				p := v.(Pair)
				value[p.Key] = p.Value
			}
			if strings.HasPrefix(key, "- ") {
				cleanKey := strings.TrimPrefix(key, "- ")
				return NewListElement(NewPair(cleanKey, value))
			}
			return NewPair(key, value)
		}
	}
	return nil
}

func extractAnchor(line, prefix string) (string, string) {
	start := strings.Index(line, prefix)
	if start != -1 {
		end := strings.Index(line[start:], " ")
		if end == -1 {
			end = len(line)
		} else {
			end += start
		}
		anchor := line[start+1 : end]
		line = strings.Replace(line, prefix+anchor, "", 1)
		return anchor, line
	}
	return "", line
}

func handleAnchor(values []any, value any, anchor, anchor2 string, anchors map[string]any) []any {
	add := true
	if anchor2 != "" {
		replaceValue := anchors[anchor2]
		if replaceValue != nil {
			switch reflect.TypeOf(value) {
			case reflect.TypeOf(ListElement{}):
				fmt.Println("REPLACE0", replaceValue, value, reflect.TypeOf(value))
				if reflect.TypeOf(replaceValue) == reflect.TypeOf(Pair{}) {
					replaceValue = replaceValue.(Pair).Value
				}
				firstElement := value.(ListElement).Value
				if reflect.TypeOf(firstElement) == reflect.TypeOf(Pair{}) {
					value = NewListElement(NewPair(firstElement.(Pair).Key, replaceValue))
				} else {
					value = NewListElement(replaceValue)
				}
			case reflect.TypeOf(Pair{}):
				if value.(Pair).Key == "<<" {
					add = false
					if reflect.TypeOf(replaceValue) == reflect.TypeOf(Pair{}) {
						for k, v := range replaceValue.(Pair).Value.(map[string]any) {
							values = addValue(values, NewPair(k, v), anchor, anchors)
						}
					}
				} else {
					if reflect.TypeOf(replaceValue) == reflect.TypeOf(ListElement{}) {
						replaceValue = replaceValue.(ListElement).Value
					} else if reflect.TypeOf(replaceValue) == reflect.TypeOf(Pair{}) {
						replaceValue = replaceValue.(Pair).ToMap()
					}
					fmt.Println("REPLACE", replaceValue, value)
					value = NewPair(value.(Pair).Key, replaceValue)
				}
			}
		}
	}
	if add {
		values = addValue(values, value, anchor, anchors)
	}
	return values
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
