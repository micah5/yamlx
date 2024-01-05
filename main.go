package yamlx

import (
	"fmt"
	"strconv"
	"strings"
)

type Pair struct {
	Key   string
	Value any
}

func NewPair(key string, value any) *Pair {
	return &Pair{key, value}
}

func Parse(data string) {
	lines := strings.Split(data, "\n")
	processLines(lines, 0)
}

// Recursive function to process lines
func processLines(lines []string, currentLevel int) {
	if len(lines) == 0 {
		return
	}

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
				fmt.Println("(map)", line)
				end := findEndOfBlock(lines, i+1, lineLevel)
				processLines(lines[i+1:end], lineLevel+1)
				i = end - 1 // Skip processed lines
			} else {
				// This line is a key/value pair
				value := processValue(line)
				fmt.Println("(k/v)", value)
			}
		}
	}
}

func processValue(line string) any {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "- ") {
		value := strings.TrimPrefix(trimmed, "- ")
		return processValue(value)
	} else if strings.Contains(trimmed, ":") {
		split := strings.Split(trimmed, ":")
		return NewPair(processValue(split[0]).(string), processValue(split[1]))
	} else {
		if intValue, err := strconv.ParseInt(trimmed, 10, 0); err == nil {
			return int(intValue)
		}
		if floatValue, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return floatValue
		}
		if boolValue, err := strconv.ParseBool(trimmed); err == nil {
			return boolValue
		}
		return trimmed
	}
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
