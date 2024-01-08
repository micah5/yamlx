package yamlx

import (
	"fmt"
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
	tokenTypes := []string{"KEY", "VALUE", "LIST_ITEM", "ANCHOR", "ALIAS", "MERGE_KEY"}
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

func (t Token) Parse(anchors map[string]any) (any, error) {
	switch t.Type {
	case KEY:
		anchor := t.Attachments.Find(ANCHOR)
		var returnValue any
		var err error
		if len(t.Children) > 0 {
			returnValue, err = parseChildren(t.Children, anchors)
		} else if len(t.Attachments) > 0 {
			value := t.Attachments.Find(VALUE)
			alias := t.Attachments.Find(ALIAS)
			if value != nil {
				returnValue = parseValue(value.Literal)
			} else if alias != nil {
				returnValue = anchors[alias.Literal]
			} else {
				return nil, fmt.Errorf("key has no value: %s", t)
			}
		} else {
			return nil, fmt.Errorf("key has no value: %s", t)
		}
		if anchor != nil {
			anchors[anchor.Literal] = returnValue
		}
		return returnValue, err
	case VALUE:
		return parseValue(t.Literal), nil
	case LIST_ITEM:
		value := t.Attachments.Find(VALUE)
		if value != nil {
			return map[string]any{t.Literal: parseValue(value.Literal)}, nil
		} else if len(t.Children) > 0 {
			returnValue, err := parseChildren(t.Children, anchors)
			return map[string]any{t.Literal: returnValue}, err
		} else {
			return parseValue(t.Literal), nil
		}
	case MERGE_KEY:
		return anchors[t.Literal], nil
	default:
		return nil, fmt.Errorf("unknown token type: %s", t)
	}
	return nil, nil
}

func parseChildren(tokens []*Token, anchors map[string]any) (any, error) {
	var returnValue any
	var err error
	if tokens[0].Type == LIST_ITEM {
		l := make([]any, len(tokens))
		for i, child := range tokens {
			l[i], err = child.Parse(anchors)
		}
		returnValue = l
	} else {
		m := make(map[string]any)
		var value any
		for _, child := range tokens {
			value, err = child.Parse(anchors)
			if value != nil {
				m[child.Literal] = value
			}
		}
		returnValue = m
	}
	return returnValue, err
}

func parseValue(literal string) any {
	if i, err := strconv.ParseInt(literal, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(literal, 64); err == nil {
		return f
	}
	if b, err := strconv.ParseBool(literal); err == nil {
		return b
	}
	return literal
}

func Tokenize(lines []string, currentLevel int) ([]*Token, error) {
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
			parts := strings.SplitN(line[2:], ":", 2)
			parentToken = NewToken(LIST_ITEM, strings.TrimSpace(parts[0]))
			if len(parts) > 1 && len(strings.TrimSpace(parts[1])) > 0 {
				parentToken.Attachments = []*Token{NewToken(VALUE, strings.TrimSpace(parts[1]))}
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
				attachments = append(attachments, NewToken(VALUE, strings.Join(values[1:], " ")))
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
			tokens = append(tokens, parentToken)
			attachments := make([]*Token, 0)
			if len(parts) > 1 && len(strings.TrimSpace(parts[1])) > 0 {
				attachments = append(attachments, NewToken(VALUE, strings.TrimSpace(parts[1])))
			}
			parentToken.Attachments = attachments
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

func Parse(tokens []*Token) (map[string]any, error) {
	result := make(map[string]any)
	anchors := make(map[string]any)
	for _, token := range tokens {
		value, err := token.Parse(anchors)
		if err != nil {
			return nil, err
		}
		if value != nil {
			result[token.Literal] = value
		}
	}
	return result, nil
}

/*func main() {
	const yamlContent = `
key1: value1
key2: &key2
  - list item 1
  - list item 2
key3:
  key3_1: &key3_1 value3_1
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

	result, err := Parse(tokens)
	fmt.Println(result, err)
}
*/
