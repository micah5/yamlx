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

func TestAnchorsAndAliasesParsing(t *testing.T) {
	yamlContent := `
anchorKey: &anchor value
aliasKey: *anchor
`
	lines := strings.Split(yamlContent, "\n")
	tokens, _ := Tokenize(lines, 0)
	result, err := Parse(tokens)

	expected := map[string]any{
		"anchorKey": "value",
		"aliasKey":  "value", // aliasKey should have the same value as anchorKey
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestMergeKeyParsing(t *testing.T) {
	yamlContent := `
defaults: &defaults
  key1: default1
  key2: default2

custom:
  <<: *defaults
  key2: custom2
`
	lines := strings.Split(yamlContent, "\n")
	tokens, _ := Tokenize(lines, 0)
	result, err := Parse(tokens)

	expected := map[string]any{
		"defaults": map[string]any{
			"key1": "default1",
			"key2": "default2",
		},
		"custom": map[string]any{
			"key1": "default1",
			"key2": "custom2", // Should override the default value
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestComplexStructureWithAnchorsAliasesMergeKeys(t *testing.T) {
	yamlContent := `
base: &base
  id: 1
  name: Base

extended:
  <<: *base
  description: Extended version

aliasTest:
  <<: *base
  id: *base
`
	lines := strings.Split(yamlContent, "\n")
	tokens, _ := Tokenize(lines, 0)
	result, err := Parse(tokens)

	expected := map[string]any{
		"base": map[string]any{
			"id":   int64(1),
			"name": "Base",
		},
		"extended": map[string]any{
			"id":          int64(1),
			"name":        "Base",
			"description": "Extended version",
		},
		"aliasTest": map[string]any{
			"id":   map[string]any{"id": int64(1), "name": "Base"},
			"name": "Base",
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
