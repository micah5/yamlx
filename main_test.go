package yamlx

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestSimpleKeyValueParsing(t *testing.T) {
	yamlContent := `
key1: value1
key2: value2
`
	lines := strings.Split(yamlContent, "\n")
	tokens, _ := Tokenize(lines, 0)
	result, err := Parse(tokens)

	expected := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestNestedStructuresParsing(t *testing.T) {
	yamlContent := `
parent:
  child1: value1
  child2: value2
  child3:
    grandchild1: value3
    grandchild2: value4
`
	lines := strings.Split(yamlContent, "\n")
	tokens, _ := Tokenize(lines, 0)
	result, err := Parse(tokens)

	expected := map[string]any{
		"parent": map[string]any{
			"child1": "value1",
			"child2": "value2",
			"child3": map[string]any{
				"grandchild1": "value3",
				"grandchild2": "value4",
			},
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestListParsing(t *testing.T) {
	yamlContent := `
items:
  - item1
  - item2
  - item3: 12
  - item4:
    - item4_1
    - item4_2
`
	lines := strings.Split(yamlContent, "\n")
	tokens, _ := Tokenize(lines, 0)
	result, err := Parse(tokens)

	expected := map[string]any{
		"items": []any{"item1", "item2", map[string]any{"item3": int64(12)}, map[string]any{"item4": []any{"item4_1", "item4_2"}}},
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestComplexStructureParsing(t *testing.T) {
	yamlContent := `
key1: value1
key2:
  - item1
  - item2
key3:
  nestedKey: 123
`
	lines := strings.Split(yamlContent, "\n")
	tokens, _ := Tokenize(lines, 0)
	result, err := Parse(tokens)

	expected := map[string]any{
		"key1": "value1",
		"key2": []any{"item1", "item2"},
		"key3": map[string]any{
			"nestedKey": int64(123),
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTokenizeErrorHandling(t *testing.T) {
	yamlContent := `
key1: value1
key2
  - item1
`
	lines := strings.Split(yamlContent, "\n")
	_, err := Tokenize(lines, 0)
	assert.Error(t, err)
}

func TestParseErrorHandling(t *testing.T) {
	yamlContent := `
key1: value1
key2:
`
	lines := strings.Split(yamlContent, "\n")
	tokens, _ := Tokenize(lines, 0)
	_, err := Parse(tokens)
	assert.Error(t, err)
}
