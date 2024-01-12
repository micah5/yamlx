package yamlx

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestFunctionCallParsing(t *testing.T) {
	yamlContent := `
numbers: &numbers [1, 2, 3, 4, 5]
len:
  - ${len("hello")}
  - ${len(numbers)}
contains:
  - ${contains("hello world", "hello")}
  - ${contains(numbers, 3)}
  - ${contains("hello world", "foo")}
  - ${contains(numbers, 6)}
max:
  - ${max(1, 2, 3, 4)}
  - ${max(numbers)}
min:
  - ${min(0, 1, 2, 3, 4)}
  - ${min(numbers)}
upper: ${upper("hello")}
lower: ${lower("HELLO")}
title: ${title("hello world")}
trim: ${trim("  hello  ")}
join:
  - ${join("-", "foo", "bar", "baz")}
  - ${join(",", numbers)}
replace: ${replace("hello world", "world", "universe")}
substr: ${substr("hello world", 0, 5)}
strrev: ${strrev("hello world")}
startswith:
  - ${startswith("hello world", "hello")}
  - ${startswith("hello world", "world")}
endswith:
  - ${endswith("hello world", "hello")}
  - ${endswith("hello world", "world")}
alltrue:
  - ${alltrue(true, true, true)}
  - ${alltrue(true, false, false)}
anytrue:
  - ${anytrue(true, false, true)}
  - ${anytrue(false, false, false)}
`
	lines := strings.Split(yamlContent, "\n")
	tokens, _ := Tokenize(lines, 0)
	result, err := Parse(tokens)

	expected := map[string]any{
		"numbers": []any{int64(1), int64(2), int64(3), int64(4), int64(5)},
		"len": []any{
			int64(5),
			int64(5),
		},
		"contains": []any{
			true,
			true,
			false,
			false,
		},
		"max": []any{
			int64(4),
			int64(5),
		},
		"min": []any{
			int64(0),
			int64(1),
		},
		"upper": "HELLO",
		"lower": "hello",
		"title": "Hello World",
		"trim":  "hello",
		"join": []any{
			"foo-bar-baz",
			"1,2,3,4,5",
		},
		"replace":    "hello universe",
		"substr":     "hello",
		"strrev":     "dlrow olleh",
		"startswith": []any{true, false},
		"endswith":   []any{false, true},
		"alltrue":    []any{true, false},
		"anytrue":    []any{true, false},
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestRandomFunctionParsing(t *testing.T) {
	yamlContent := `
numbers: &numbers [1, 2, 3]
names: &names [foo, bar, baz]
random:
  - ${rand(1, 10)}
  - ${rand(numbers)}
  - ${rand(names)}
`
	lines := strings.Split(yamlContent, "\n")
	tokens, _ := Tokenize(lines, 0)
	result, err := Parse(tokens)

	l := result["random"].([]any)
	assert.GreaterOrEqual(t, l[0].(int64), int64(1))
	assert.LessOrEqual(t, l[0].(int64), int64(10))
	assert.Contains(t, []any{int64(1), int64(2), int64(3)}, l[1].(int64))
	assert.Contains(t, []any{"foo", "bar", "baz"}, l[2].(string))

	assert.NoError(t, err)
}
