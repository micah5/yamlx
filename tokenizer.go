package yamlx

import (
	"fmt"
	"regexp"
	"strconv"
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
	LOOP
	LOOP_RANGE
)

type Tokens []*Token

func (t Tokens) Find(typ Type) *Token {
	for _, token := range t {
		if token.Type == typ {
			return token
		}
	}
	return nil
}

// Token represents a lexical token.
type Token struct {
	Type        Type
	Literal     string
	Children    Tokens // To hold nested tokens
	Attachments Tokens // To hold attachments
}

func NewToken(t Type, literal string) *Token {
	return &Token{t, literal, nil, nil}
}

func (t Token) String() string {
	tokenTypes := []string{"KEY", "VALUE", "LIST_ITEM", "ANCHOR", "ALIAS", "MERGE_KEY", "LOOP", "LOOP_RANGE"}
	result := fmt.Sprintf("%s: %s", tokenTypes[t.Type], t.Literal)
	return result
}

func (t Token) Print(prefix string) {
	fmt.Println(prefix + t.String())
	for _, attachment := range t.Attachments {
		attachment.Print(prefix + " ")
	}
	for _, child := range t.Children {
		child.Print(prefix + "\t")
	}
}

func Tokenize(lines []string, currentLevel int) ([]*Token, error) {
	var tokens []*Token
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Remove comments
		if strings.Contains(line, "#") {
			line = strings.Split(line, "#")[0]
		}

		// Trim spaces
		line = strings.TrimSpace(line)

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Determine the token type based on the line
		var parentToken *Token
		if strings.Contains(line, "!for") {
			// for loop
			// format is: !for <var> in <range>:

			// get text between !for and :
			startIndex := strings.Index(line, "!for")
			endIndex := strings.Index(line, ":")
			contents := strings.TrimSpace(line[startIndex+4 : endIndex])
			// split by in
			values := strings.Split(contents, " in ")
			if len(values) != 2 {
				return nil, fmt.Errorf("invalid for loop: %s", line)
			}
			variable := strings.TrimSpace(values[0])
			parentToken = NewToken(LOOP, variable)
			rangeString := strings.TrimSpace(values[1])
			if strings.HasPrefix(rangeString, "[") {
				// get the string between brackets
				rangeString = strings.Trim(strings.TrimSpace(rangeString), "[]")
			}
			parentToken.Attachments = []*Token{NewToken(LOOP_RANGE, rangeString)}
			tokens = append(tokens, parentToken)
		} else if strings.HasPrefix(line, "- ") {
			parts := strings.SplitN(line[2:], ":", 2)
			parentToken = NewToken(LIST_ITEM, strings.TrimSpace(parts[0]))
			if len(parts) > 1 && len(strings.TrimSpace(parts[1])) > 0 {
				attachment := handleKeyValueString(parentToken, strings.TrimSpace(parts[1]))
				if attachment != nil {
					parentToken.Attachments = []*Token{attachment}
				}
			}
			tokens = append(tokens, parentToken)
		} else if strings.HasPrefix(line, "<<: *") {
			parts := strings.SplitN(line, "<<: *", 2)
			tokens = append(tokens, NewToken(MERGE_KEY, strings.TrimSpace(parts[1])))
		} else if strings.Contains(line, ": &") {
			parts := strings.SplitN(line, ": &", 2)
			parentToken = NewToken(KEY, strings.TrimSpace(parts[0]))
			tokens = append(tokens, parentToken)
			value := strings.TrimSpace(parts[1])
			values := strings.Split(value, " ")
			attachments := make([]*Token, 0)
			if len(values) > 1 {
				attachments = append(attachments, NewToken(ANCHOR, values[0]))
				attachment := handleKeyValueString(parentToken, strings.Join(values[1:], " "))
				if attachment != nil {
					attachments = append(attachments, attachment)
				}
			} else {
				attachments = append(attachments, NewToken(ANCHOR, value))
			}
			parentToken.Attachments = attachments
		} else if strings.Contains(line, ": *") {
			parts := strings.SplitN(line, ": *", 2)
			token := NewToken(KEY, strings.TrimSpace(parts[0]))
			token.Attachments = []*Token{NewToken(ALIAS, strings.TrimSpace(parts[1]))}
			tokens = append(tokens, token)
		} else if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			parentToken = NewToken(KEY, strings.TrimSpace(parts[0]))
			if len(parts) > 1 && len(strings.TrimSpace(parts[1])) > 0 {
				attachment := handleKeyValueString(parentToken, strings.TrimSpace(parts[1]))
				if attachment != nil {
					parentToken.Attachments = []*Token{attachment}
				}
			}
			tokens = append(tokens, parentToken)
		} else {
			return nil, fmt.Errorf("invalid line: %s", line)
		}

		if i < len(lines)-1 {
			nextIndent := countLeadingSpaces(lines[i+1])
			if parentToken != nil && nextIndent > currentLevel {
				// Process nested lines
				end := findEndOfBlock(lines, i+1, currentLevel)
				nestedTokens, err := Tokenize(lines[i+1:end], nextIndent)
				if err != nil {
					return nil, err
				}
				parentToken.Children = nestedTokens
				i = end - 1 // Skip processed lines
			}
		}
	}
	return tokens, nil
}

func handleKeyValueString(parentToken *Token, value string) *Token {
	// returns attachment if necessary
	re := regexp.MustCompile(`(\$\{[^}]*\}|"[^"]*")|(\*)`)
	matches := re.FindAllStringSubmatch(value, -1)
	found := false
	for _, match := range matches {
		// match[2] contains the asterisks outside the excluded patterns
		if match[2] != "" {
			found = true
			break
		}
	}
	if found {
		// get the word after *
		parts := strings.SplitN(value, "*", 2)
		attachment := NewToken(ALIAS, strings.TrimSpace(parts[1]))
		return attachment
	} else if strings.HasPrefix(strings.TrimSpace(value), "[") {
		// get the string between brackets
		contents := strings.Trim(strings.TrimSpace(value), "[]")
		// split by ..
		values := strings.Split(contents, "..")
		if len(values) == 2 {
			start, end := strings.TrimSpace(values[0]), strings.TrimSpace(values[1])
			startInt, _ := strconv.ParseInt(start, 10, 64)
			endInt, _ := strconv.ParseInt(end, 10, 64)
			children := make([]*Token, 0)
			for i := startInt; i <= endInt; i++ {
				children = append(children, NewToken(LIST_ITEM, fmt.Sprintf("%d", i)))
			}
			parentToken.Children = children
		} else {
			// split by comma
			values := strings.Split(contents, ",")
			children := make([]*Token, 0)
			for _, value := range values {
				children = append(children, NewToken(LIST_ITEM, strings.TrimSpace(value)))
			}
			parentToken.Children = children
		}
	} else {
		return NewToken(VALUE, strings.TrimSpace(value))
	}
	return nil
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
