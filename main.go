package yamlx

import (
	"errors"
	"fmt"
	"github.com/Knetic/govaluate"
	"gopkg.in/yaml.v3"
	"math/rand"
	"reflect"
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
				returnValue, err = parseValue(value.Literal, anchors)
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
			if m, ok := returnValue.(map[string]any); ok {
				for k, v := range createAnchorMap(m, anchor.Literal) {
					anchors[k] = v
				}
			}
		}
		return returnValue, err
	case VALUE:
		return parseValue(t.Literal, anchors)
	case LIST_ITEM:
		value := t.Attachments.Find(VALUE)
		alias := t.Attachments.Find(ALIAS)
		if value != nil || alias != nil {
			var returnValue any
			var err error
			if value != nil {
				returnValue, err = parseValue(value.Literal, anchors)
				if err != nil {
					return nil, err
				}
			} else {
				returnValue = anchors[alias.Literal]
			}
			newMap := map[string]any{t.Literal: returnValue}
			for _, child := range t.Children {
				if child.Type == KEY {
					childValue, err := child.Parse(anchors)
					if err != nil {
						return newMap, err
					}
					newMap[child.Literal] = childValue
				} else {
					return newMap, fmt.Errorf("invalid child type: %s", child)
				}
			}
			return newMap, nil
		} else if len(t.Children) > 0 {
			returnValue, err := parseChildren(t.Children, anchors)
			return map[string]any{t.Literal: returnValue}, err
		} else {
			return parseValue(t.Literal, anchors)
		}
	case MERGE_KEY:
		anchorValue := anchors[t.Literal]
		if anchorValue == nil {
			return nil, fmt.Errorf("anchor not found: %s", t.Literal)
		}
		return anchorValue, nil
	default:
		return nil, fmt.Errorf("unknown token type: %s", t)
	}
	return nil, nil
}

func createAnchorMap(value map[string]any, prefix string) map[string]any {
	returnMap := make(map[string]any)
	for k, v := range value {
		if m, ok := v.(map[string]any); ok {
			for k2, v2 := range createAnchorMap(m, prefix+"."+k) {
				returnMap[k2] = v2
			}
		} else {
			returnMap[prefix+"."+k] = v
		}
	}
	return returnMap
}

func parseChildren(tokens []*Token, anchors map[string]any) (any, error) {
	var returnValue any
	var err error
	isList := tokens[0].Type == LIST_ITEM
	if tokens[0].Type == LOOP {
		isList = tokens[0].Children[0].Type == LIST_ITEM
	}
	if isList {
		l := make([]any, 0)
		for _, child := range tokens {
			if child.Type == LOOP {
				r := child.Attachments.Find(LOOP_RANGE)
				rangeString := r.Literal
				arr := make([]any, 0)
				if strings.HasPrefix(rangeString, "*") {
					parts := strings.SplitN(rangeString, "*", 2)
					anchorKey := strings.TrimSpace(parts[1])
					arr = anchors[anchorKey].([]any)
				} else if strings.Contains(rangeString, "..") {
					parts := strings.SplitN(rangeString, "..", 2)
					start, _ := strconv.ParseInt(parts[0], 10, 64)
					end, _ := strconv.ParseInt(parts[1], 10, 64)
					for i := start; i <= end; i++ {
						arr = append(arr, i)
					}
				} else {
					parts := strings.Split(rangeString, ",")
					for _, part := range parts {
						arr = append(arr, part)
					}
				}
				keys := strings.Split(child.Literal, ",")
				variableKey := keys[0]
				var indexKey string
				if len(keys) == 2 {
					variableKey = strings.TrimSpace(keys[1])
					indexKey = strings.TrimSpace(keys[0])
				}
				for _, nestedChild := range child.Children {
					for i, v := range arr {
						anchors[variableKey] = v
						if indexKey != "" {
							anchors[indexKey] = i
						}
						elem, _ := nestedChild.Parse(anchors)
						l = append(l, elem)
					}
				}
			} else {
				v, _ := child.Parse(anchors)
				l = append(l, v)
			}
		}
		returnValue = l
	} else {
		m := make(map[string]any)
		var value any
		for _, child := range tokens {
			value, err = child.Parse(anchors)
			if value != nil {
				if valueMap, ok := value.(map[string]any); ok && child.Type == MERGE_KEY {
					for k, v := range valueMap {
						m[k] = v
					}
				} else {
					m[child.Literal] = value
				}
			}
		}
		returnValue = m
	}
	return returnValue, err
}

func parseValue(literal string, anchors map[string]any) (any, error) {
	literal, err := replaceWithMap(literal, anchors)
	if err != nil {
		return nil, err
	}

	if i, err := strconv.ParseInt(literal, 10, 64); err == nil {
		return i, nil
	}
	if f, err := strconv.ParseFloat(literal, 64); err == nil {
		return f, nil
	}
	if b, err := strconv.ParseBool(literal); err == nil {
		return b, nil
	}
	if strings.HasPrefix(literal, "\"") && strings.HasSuffix(literal, "\"") {
		return literal[1 : len(literal)-1], nil
	}
	return literal, nil
}

var functions = map[string]govaluate.ExpressionFunction{
	"len": func(args ...any) (any, error) {
		if strval, ok := args[0].(string); ok {
			length := len(strval)
			return (float64)(length), nil
		} else {
			return len(args), nil
		}
	},
	"contains": func(args ...any) (any, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("contains function requires 2 arguments")
		}
		if strval, ok := args[0].(string); ok {
			if substrval, ok := args[1].(string); ok {
				return strings.Contains(strval, substrval), nil
			}
		} else {
			found := false
			for _, v := range args[:len(args)-1] {
				expression, err := govaluate.NewEvaluableExpression(fmt.Sprintf("%v == %v", v, args[len(args)-1]))
				if err != nil {
					return nil, err
				}
				result, err := expression.Evaluate(nil)
				if err != nil {
					return nil, err
				}
				if result.(bool) {
					found = true
					break
				}
			}
			return found, nil
		}
		return nil, fmt.Errorf("contains function requires string arguments")
	},
	"rand": func(args ...any) (any, error) {
		if min, ok := args[0].(float64); ok && len(args) == 2 {
			if max, ok := args[1].(float64); ok {
				return rand.Int63n(int64(max-min)) + int64(min), nil
			} else {
				return nil, fmt.Errorf("rand function requires 2 integer arguments")
			}
		} else {
			// select random option in args
			return args[rand.Intn(len(args))], nil
		}
		return nil, fmt.Errorf("invalid arguments for rand function")
	},
	"max": func(args ...any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("max function requires at least 1 argument")
		}
		max := args[0]
		for _, v := range args {
			if i, ok := v.(int64); ok {
				if i > max.(int64) {
					max = i
				}
			} else if f, ok := v.(float64); ok {
				if f > max.(float64) {
					max = f
				}
			} else {
				return nil, fmt.Errorf("max function requires numeric arguments")
			}
		}
		return max, nil
	},
	"min": func(args ...any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("min function requires at least 1 argument")
		}
		min := args[0]
		for _, v := range args {
			if i, ok := v.(int64); ok {
				if i < min.(int64) {
					min = i
				}
			} else if f, ok := v.(float64); ok {
				if f < min.(float64) {
					min = f
				}
			} else {
				return nil, fmt.Errorf("min function requires numeric arguments")
			}
		}
		return min, nil
	},
	"upper": func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("upper function requires 1 argument")
		}
		if strval, ok := args[0].(string); ok {
			return strings.ToUpper(strval), nil
		}
		return nil, fmt.Errorf("upper function requires string argument")
	},
	"lower": func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("lower function requires 1 argument")
		}
		if strval, ok := args[0].(string); ok {
			return strings.ToLower(strval), nil
		}
		return nil, fmt.Errorf("lower function requires string argument")
	},
	"title": func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("title function requires 1 argument")
		}
		if strval, ok := args[0].(string); ok {
			return strings.Title(strval), nil
		}
		return nil, fmt.Errorf("title function requires string argument")
	},
	"trim": func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("trim function requires 1 argument")
		}
		if strval, ok := args[0].(string); ok {
			return strings.TrimSpace(strval), nil
		}
		return nil, fmt.Errorf("trim function requires string argument")
	},
	"join": func(args ...any) (any, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("join function requires 2 arguments")
		}
		if strval, ok := args[0].(string); ok {
			if slice, ok := args[1].([]any); ok {
				strs := make([]string, len(slice))
				for i, v := range slice {
					strs[i] = fmt.Sprintf("%v", v)
				}
				return strings.Join(strs, strval), nil
			} else {
				// join remaining arguments
				strs := make([]string, len(args)-1)
				for i, v := range args[1:] {
					strs[i] = fmt.Sprintf("%v", v)
				}
				return strings.Join(strs, strval), nil
			}
		}
		return nil, fmt.Errorf("join function requires string and slice arguments")
	},
	"replace": func(args ...any) (any, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("replace function requires 3 arguments")
		}
		if strval, ok := args[0].(string); ok {
			if oldstrval, ok := args[1].(string); ok {
				if newstrval, ok := args[2].(string); ok {
					return strings.Replace(strval, oldstrval, newstrval, -1), nil
				}
			}
		}
		return nil, fmt.Errorf("replace function requires string arguments")
	},
	"substr": func(args ...any) (any, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("substr function requires 3 arguments")
		}
		if strval, ok := args[0].(string); ok {
			if startval, ok := args[1].(float64); ok {
				if endval, ok := args[2].(float64); ok {
					return strval[int(startval):int(endval)], nil
				}
			}
		}
		return nil, fmt.Errorf("substr function requires string and integer arguments")
	},
	"strrev": func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("strrev function requires 1 argument")
		}
		if strval, ok := args[0].(string); ok {
			runes := []rune(strval)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			return string(runes), nil
		}
		return nil, fmt.Errorf("strrev function requires string argument")
	},
	"startswith": func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("startswith function requires 2 arguments")
		}
		if strval, ok := args[0].(string); ok {
			if substrval, ok := args[1].(string); ok {
				return strings.HasPrefix(strval, substrval), nil
			}
		}
		return nil, fmt.Errorf("startswith function requires string arguments")
	},
	"endswith": func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("endswith function requires 2 arguments")
		}
		if strval, ok := args[0].(string); ok {
			if substrval, ok := args[1].(string); ok {
				return strings.HasSuffix(strval, substrval), nil
			}
		}
		return nil, fmt.Errorf("endswith function requires string arguments")
	},
	"alltrue": func(args ...any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("alltrue function requires at least 1 argument")
		}
		for _, v := range args {
			if b, ok := v.(bool); ok {
				if !b {
					return false, nil
				}
			} else {
				return nil, fmt.Errorf("alltrue function requires boolean arguments")
			}
		}
		return true, nil
	},
	"anytrue": func(args ...any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("anytrue function requires at least 1 argument")
		}
		for _, v := range args {
			if b, ok := v.(bool); ok {
				if b {
					return true, nil
				}
			} else {
				return nil, fmt.Errorf("anytrue function requires boolean arguments")
			}
		}
		return false, nil
	},
}

func replaceWithMap(input string, anchors map[string]any) (string, error) {
	// Regular expression to find ${} patterns
	re := regexp.MustCompile(`\$\{([^\}]+)\}`)
	matches := re.FindAllStringSubmatch(input, -1)

	outputString := input
	for _, m := range matches {
		expressionString := m[1]

		// Wrap any anchors in square brackets
		for k, _ := range anchors {
			if expressionString == k {
				expressionString = fmt.Sprintf("[%v]", k)
			} else {
				index := strings.Index(expressionString, k)
				if index >= 0 && (index+len(k) < len(expressionString) && expressionString[index+len(k)] == ' ' || index+len(k) == len(expressionString)) {
					expressionString = strings.Replace(expressionString, k, fmt.Sprintf("[%v]", k), -1)
				}
			}
		}

		// Evaluate the expression
		expression, err := govaluate.NewEvaluableExpressionWithFunctions(expressionString, functions)
		if err != nil {
			return "", err
		}

		result, err := expression.Evaluate(anchors)
		if err != nil {
			return "", err
		}

		resultString := fmt.Sprintf("%v", result)
		outputString = strings.Replace(outputString, m[0], resultString, -1)
	}

	return outputString, nil
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

// Unmarshals YAMLX data into a Go struct
func Unmarshal(data []byte, v interface{}) error {
	lines := strings.Split(string(data), "\n")
	tokens, err := Tokenize(lines, 0)
	if err != nil {
		return err
	}

	parsedData, err := Parse(tokens)
	if err != nil {
		return err
	}

	return mapToStruct(parsedData, v)
}

// Marshal just returns the data as yaml
func Marshal(v interface{}) ([]byte, error) {
	noTagType := createModifiedTagType(reflect.TypeOf(v))
	noTagValue := reflect.New(noTagType).Elem()
	copyToModifiedTagStruct(reflect.ValueOf(v), noTagValue)
	newVal := noTagValue.Interface()
	return yaml.Marshal(newVal)
}

// createModifiedTagType creates a parallel type for the given type with modified tags.
func createModifiedTagType(t reflect.Type) reflect.Type {
	switch t.Kind() {
	case reflect.Struct:
		fields := make([]reflect.StructField, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tag := string(field.Tag)
			modifiedTag := strings.Replace(tag, "yamlx", "yaml", -1)
			fields[i] = reflect.StructField{
				Name: field.Name,
				Type: createModifiedTagType(field.Type),
				Tag:  reflect.StructTag(modifiedTag),
			}
		}
		return reflect.StructOf(fields)

	case reflect.Slice:
		elemType := createModifiedTagType(t.Elem())
		return reflect.SliceOf(elemType)

	default:
		return t
	}
}

// copyToModifiedTagStruct copies data from the original struct to the modified tag struct.
func copyToModifiedTagStruct(src, dest reflect.Value) {
	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		destField := dest.Field(i)

		if srcField.Kind() == reflect.Struct {
			newDestField := reflect.New(destField.Type()).Elem()
			copyToModifiedTagStruct(srcField, newDestField)
			destField.Set(newDestField)
		} else if srcField.Kind() == reflect.Slice {
			newSlice := reflect.MakeSlice(destField.Type(), srcField.Len(), srcField.Cap())
			for j := 0; j < srcField.Len(); j++ {
				newElem := reflect.New(destField.Type().Elem()).Elem()
				copyToModifiedTagStruct(srcField.Index(j), newElem)
				newSlice.Index(j).Set(newElem)
			}
			destField.Set(newSlice)
		} else {
			destField.Set(srcField)
		}
	}
}

// mapToStruct maps a generic map[string]any to a struct
func mapToStruct(m map[string]any, s interface{}) error {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("yamlx: Unmarshal requires a non-nil pointer to a struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("yamlx: Unmarshal requires a pointer to a struct")
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("yamlx")
		if tag != "" {
			tagParts := strings.Split(tag, ",")
			tag = tagParts[0] // Only use the first part of the tag
		} else {
			tag = field.Name
		}

		value, ok := m[tag]
		if !ok {
			continue
		}

		fieldVal := val.Field(i)
		if fieldVal.IsValid() && fieldVal.CanSet() {
			setField(value, fieldVal)
		}
	}

	return nil
}

// setField sets a field of a struct based on its type
func setField(value any, fieldVal reflect.Value) {
	if value == nil {
		return
	}

	switch fieldVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, ok := value.(int64); ok {
			fieldVal.SetInt(val)
		}
	case reflect.Float32, reflect.Float64:
		if val, ok := value.(float64); ok {
			fieldVal.SetFloat(val)
		}
	case reflect.String:
		if val, ok := value.(string); ok {
			fieldVal.SetString(val)
		}
	case reflect.Bool:
		if val, ok := value.(bool); ok {
			fieldVal.SetBool(val)
		}
	case reflect.Slice:
		if val, ok := value.([]any); ok {
			slice := reflect.MakeSlice(fieldVal.Type(), len(val), len(val))
			for i := 0; i < len(val); i++ {
				setField(val[i], slice.Index(i))
			}
			fieldVal.Set(slice)
		}
	case reflect.Map:
		if val, ok := value.(map[string]any); ok {
			m := reflect.MakeMap(fieldVal.Type())
			for k, v := range val {
				mapVal := reflect.New(fieldVal.Type().Elem()).Elem()
				setField(v, mapVal)
				m.SetMapIndex(reflect.ValueOf(k), mapVal)
			}
			fieldVal.Set(m)
		}
	case reflect.Struct:
		if val, ok := value.(map[string]any); ok {
			mapToStruct(val, fieldVal.Addr().Interface())
		}
	}
}
