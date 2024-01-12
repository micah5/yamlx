package yamlx

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"regexp"
	"strconv"
	"strings"
)

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
	"len":        length,
	"contains":   contains,
	"rand":       calcRand,
	"max":        calcMax,
	"min":        calcMin,
	"upper":      upper,
	"lower":      lower,
	"title":      title,
	"trim":       trim,
	"join":       join,
	"replace":    replace,
	"substr":     substr,
	"strrev":     strrev,
	"startswith": startswith,
	"endswith":   endswith,
	"alltrue":    alltrue,
	"anytrue":    anytrue,
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
