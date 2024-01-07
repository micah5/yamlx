package main

import (
	"bufio"
	"fmt"
	"strings"
)

// Token types
const (
	KEY = iota
	VALUE
	LIST_ITEM
	ANCHOR
	ALIAS
	MERGE_KEY
)

// Token represents a lexical token.
type Token struct {
	Type     int
	Literal  string
	Children []Token // To hold nested tokens
}

func (t Token) String() string {
	tokenTypes := []string{"KEY", "VALUE", "LIST_ITEM", "ANCHOR", "ALIAS", "MERGE_KEY"}
	result := fmt.Sprintf("%s: %s", tokenTypes[t.Type], t.Literal)
	for _, child := range t.Children {
		result += "\n\t" + strings.ReplaceAll(child.String(), "\n", "\n\t")
	}
	return result
}

func (t Token) Print(prefix string) {
	fmt.Println(prefix + t.String())
	for _, child := range t.Children {
		child.Print(prefix + "\t")
	}
}

func Tokenize(input string, baseIndent int) []Token {
	var tokens []Token
	scanner := bufio.NewScanner(strings.NewReader(input))
	var currentToken *Token
	var nestedLines string

	for scanner.Scan() {
		line := scanner.Text()
		currentIndent := countLeadingSpaces(line)
		line = strings.TrimSpace(line)

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Check if the line belongs to a nested structure
		if currentToken != nil && currentIndent > baseIndent {
			nestedLines += line + "\n"
			continue
		}

		// Tokenize nested structure
		if nestedLines != "" {
			nestedTokens := Tokenize(nestedLines, baseIndent+1)
			currentToken.Children = nestedTokens
			nestedLines = ""
		}

		// Determine the token type based on the line
		if strings.HasPrefix(line, "- ") {
			tokens = append(tokens, Token{LIST_ITEM, strings.TrimSpace(line[2:]), nil})
		} else if strings.HasPrefix(line, "<<: *") {
			tokens = append(tokens, Token{MERGE_KEY, strings.TrimSpace(line[4:]), nil})
		} else if strings.Contains(line, ": &") {
			parts := strings.SplitN(line, ": &", 2)
			currentToken = &Token{KEY, strings.TrimSpace(parts[0]), nil}
			tokens = append(tokens, *currentToken)
			tokens = append(tokens, Token{ANCHOR, strings.TrimSpace(parts[1]), nil})
		} else if strings.Contains(line, ": *") {
			parts := strings.SplitN(line, ": *", 2)
			currentToken = &Token{KEY, strings.TrimSpace(parts[0]), nil}
			tokens = append(tokens, *currentToken)
			tokens = append(tokens, Token{ALIAS, strings.TrimSpace(parts[1]), nil})
		} else if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			currentToken = &Token{KEY, strings.TrimSpace(parts[0]), nil}
			tokens = append(tokens, *currentToken)
			if len(parts) > 1 && len(strings.TrimSpace(parts[1])) > 0 {
				tokens = append(tokens, Token{VALUE, strings.TrimSpace(parts[1]), nil})
			}
		}
	}

	// Tokenize any remaining nested structure
	if nestedLines != "" {
		nestedTokens := Tokenize(nestedLines, baseIndent+1)
		currentToken.Children = nestedTokens
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

	tokens := Tokenize(yamlContent, 0)
	for _, token := range tokens {
		token.Print("")
	}
}
