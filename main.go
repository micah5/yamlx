package main

import (
	"fmt"
	"strings"
)

// Token types
type Type int

const (
	KEY Type = iota
	VALUE
	LIST_ITEM
	ANCHOR
	ALIAS
	MERGE_KEY
)

// Token represents a lexical token.
type Token struct {
	Type     Type
	Literal  string
	Children []*Token // To hold nested tokens
}

func NewToken(t Type, literal string) *Token {
	return &Token{t, literal, nil}
}

func (t Token) String() string {
	tokenTypes := []string{"KEY", "VALUE", "LIST_ITEM", "ANCHOR", "ALIAS", "MERGE_KEY"}
	result := fmt.Sprintf("%s: %s", tokenTypes[t.Type], t.Literal)
	return result
}

func (t Token) Print(prefix string) {
	fmt.Println(prefix + t.String())
	for _, child := range t.Children {
		child.Print(prefix + "\t")
	}
}

func Tokenize(lines []string, currentLevel int) []*Token {
	if len(lines) == 0 {
		return nil
	}

	var tokens []*Token
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		line = strings.TrimSpace(line)

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Remove comments
		if strings.Contains(line, "#") {
			line = strings.Split(line, "#")[0]
		}

		// Determine the token type based on the line
		var parentToken *Token
		if strings.HasPrefix(line, "- ") {
			tokens = append(tokens, NewToken(LIST_ITEM, strings.TrimSpace(line[2:])))
		} else if strings.HasPrefix(line, "<<: *") {
			tokens = append(tokens, NewToken(MERGE_KEY, strings.TrimSpace(line[4:])))
		} else if strings.Contains(line, ": &") {
			parts := strings.SplitN(line, ": &", 2)
			parentToken = NewToken(KEY, strings.TrimSpace(parts[0]))
			tokens = append(tokens, parentToken)
			tokens = append(tokens, NewToken(ANCHOR, strings.TrimSpace(parts[1])))
		} else if strings.Contains(line, ": *") {
			parts := strings.SplitN(line, ": *", 2)
			tokens = append(tokens, NewToken(KEY, strings.TrimSpace(parts[0])))
			tokens = append(tokens, NewToken(ALIAS, strings.TrimSpace(parts[1])))
		} else if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			parentToken = NewToken(KEY, strings.TrimSpace(parts[0]))
			tokens = append(tokens, parentToken)
			if len(parts) > 1 && len(strings.TrimSpace(parts[1])) > 0 {
				tokens = append(tokens, NewToken(VALUE, strings.TrimSpace(parts[1])))
			}
		}

		if i < len(lines)-1 {
			nextIndent := countLeadingSpaces(lines[i+1])
			if parentToken != nil && nextIndent > currentLevel {
				// Process nested lines
				end := findEndOfBlock(lines, i+1, currentLevel)
				nestedTokens := Tokenize(lines[i+1:end], nextIndent)
				parentToken.Children = nestedTokens
				i = end - 1 // Skip processed lines
			}
		}
	}
	return tokens
}

// countLeadingSpaces counts the number of leading spaces in a string.
func countLeadingSpaces(str string) int {
	count := 0
	for _, ch := range str {
		if ch == ' ' {
			count++
		} else {
			break
		}
	}
	return count
}

// findEndOfBlock finds the end index of a block at a given indentation level.
func findEndOfBlock(lines []string, start, level int) int {
	for i := start; i < len(lines); i++ {
		if countLeadingSpaces(lines[i]) <= level {
			return i
		}
	}
	return len(lines)
}

func main() {
	const yamlContent = `
key1: value1
key2: &key2
  - list item 1
  - list item 2
key3:
  key3_1: value3_1
  key3_2: value3_2
  key3_3:
    key3_3_1: value3_3_1
key4:
  key4_1: &key4_1 value4_1
  key4_2: *key4_1
key5:
  <<: *key2
  key5_1: value5_1
    `

	lines := strings.Split(yamlContent, "\n")
	tokens := Tokenize(lines, 0)
	for _, token := range tokens {
		token.Print("")
	}
}
