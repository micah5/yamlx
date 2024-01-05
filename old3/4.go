package yamlx

import (
	"fmt"
	"strings"
)

func Parse(data string) {
	lines := strings.Split(data, "\n")
	anchors := make(map[string]any) // Store anchors
	processLines(lines, 0, anchors)
}

func processLines(lines []string, currentLevel int, anchors map[string]any) {
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

		// Evaluate line

		// Extract anchors
		anchor := extractAnchor(line)
		if anchor != "" {
			// Store anchor and its content
			anchors[anchor] = trimmedLine
		}

		// Replace alias with its anchor content
		trimmedLine = replaceAlias(trimmedLine, anchors)

		// Check if line is at the current level
		if lineLevel == currentLevel {
			if i < len(lines)-1 && strings.Count(lines[i+1], "\t") > lineLevel {
				// This line is a map, call recursively for nested lines
				fmt.Println("(map)", trimmedLine)
				end := findEndOfBlock(lines, i+1, lineLevel)
				processLines(lines[i+1:end], lineLevel+1, anchors)
				i = end - 1 // Skip processed lines
			} else {
				// This line is a key/value pair
				fmt.Println("(k/v)", trimmedLine)
			}
		}
	}
}

func extractAnchor(line string) string {
	start := strings.Index(line, "&")
	if start != -1 {
		end := strings.Index(line[start:], " ")
		if end == -1 {
			end = len(line)
		} else {
			end += start
		}
		anchor := line[start+1 : end]
		line = strings.Replace(line, "&"+anchor, "", 1)
		return anchor
	}
	return ""
}

func replaceAlias(line string, anchors map[string]any) string {
	for alias, content := range anchors {
		aliasMarker := "*" + alias
		if strings.Contains(line, aliasMarker) {
			line = strings.Replace(line, aliasMarker, content, 1)
		}
	}
	return line
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
