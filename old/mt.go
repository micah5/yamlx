package yamlx

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestDetermineLineTypeAndIndent(t *testing.T) {
	testCases := []struct {
		line     string
		indent   int
		lineType LineType
	}{
		{"key: value", 0, LineTypeKeyValue},
		{"  - item", 1, LineTypeList},
		{"unknown", 0, LineTypeUnknown},
	}

	for _, tc := range testCases {
		indent, lineType := determineLineTypeAndIndent(tc.line)
		assert.Equal(t, tc.indent, indent)
		assert.Equal(t, tc.lineType, lineType)
	}
}

func TestParseLine(t *testing.T) {
	testCases := []struct {
		input string
		key   string
		value any
		err   bool
	}{
		{"key: value", "key", "value", false},
		{"name: John Doe", "name", "John Doe", false},
		{"age: 30", "age", 30, false},
		{"temperature: 98.6", "temperature", 98.6, false},
		{"active: true", "active", true, false},
		{"invalid", "", "", true},
	}

	for _, tc := range testCases {
		key, value, err := parseLine(tc.input)
		if tc.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, tc.key, key)
			assert.Equal(t, tc.value, value)
		}
	}
}

func TestParseList(t *testing.T) {
	testCases := []struct {
		name  string
		lines []string
		want  []any
		err   bool
	}{
		{
			name:  "List of Strings",
			lines: []string{"- item1", "- item2", "- item3"},
			want:  []any{"item1", "item2", "item3"},
			err:   false,
		},
		{
			name:  "List of Integers",
			lines: []string{"- 1", "- 2", "- 3"},
			want:  []any{1, 2, 3},
			err:   false,
		},
		{
			name:  "List of Floats",
			lines: []string{"- 1.1", "- 2.2", "- 3.3"},
			want:  []any{1.1, 2.2, 3.3},
			err:   false,
		},
		{
			name:  "List of Bools",
			lines: []string{"- true", "- false", "- true"},
			want:  []any{true, false, true},
			err:   false,
		},
		{
			name:  "List of Mixed Types",
			lines: []string{"- item1", "- 2", "- 3.3", "- true"},
			want:  []any{"item1", 2, 3.3, true},
			err:   false,
		},
		{
			name:  "List of Maps",
			lines: []string{"- key1: value1", "  key2: value2", "- key3: value3"},
			want:  []any{map[string]any{"key1": "value1", "key2": "value2"}, map[string]any{"key3": "value3"}},
			err:   false,
		},
		{
			name:  "Invalid Format",
			lines: []string{"- item1", "item2"},
			want:  nil,
			err:   true,
		},
	}

	for _, tc := range testCases {
		got, err := parseList(tc.lines, 0)
		if (err != nil) != tc.err {
			t.Errorf("TestParseList %s error = %v, wantErr %v", tc.name, err, tc.err)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("TestParseList %s = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestParseMap(t *testing.T) {
	testCases := []struct {
		name  string
		lines []string
		want  map[string]any
		err   bool
	}{
		{
			name:  "Simple Map",
			lines: []string{"key1: value1", "key2: value2"},
			want:  map[string]any{"key1": "value1", "key2": "value2"},
			err:   false,
		},
		{
			name:  "Nested List in Map",
			lines: []string{"key1:", "  - item1", "  - item2", "key2: value2"},
			want:  map[string]any{"key1": []any{"item1", "item2"}, "key2": "value2"},
			err:   false,
		},
		{
			name:  "Nested Map in Map",
			lines: []string{"key1:", "  subkey1: value1", "  subkey2: value2", "key2: value2"},
			want:  map[string]any{"key1": map[string]any{"subkey1": "value1", "subkey2": "value2"}, "key2": "value2"},
			err:   false,
		},
		{
			name:  "Invalid Format",
			lines: []string{"key1 value1", "key2: value2"},
			want:  nil,
			err:   true,
		},
	}

	for _, tc := range testCases {
		got, err := parseMap(tc.lines, 0)
		if (err != nil) != tc.err {
			t.Errorf("TestParseMap %s error = %v, wantErr %v", tc.name, err, tc.err)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("TestParseMap %s = %v, want %v", tc.name, got, tc.want)
		}
	}
}

/*
func TestParseNested(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected map[string]any
		err      bool
	}{
		{
			name: "Simple Nested Map",
			input: []string{
				"key:",
				"  nestedKey: value",
			},
			expected: map[string]any{
				"key": map[string]any{"nestedKey": "value"},
			},
			err: false,
		},
		{
			name: "2D Nested Map",
			input: []string{
				"key:",
				"  nestedKey:",
				"    nestedNestedKey: value",
			},
			expected: map[string]any{
				"key": map[string]any{"nestedKey": map[string]any{"nestedNestedKey": "value"}},
			},
			err: false,
		},
		{
			name: "Nested List of Maps",
			input: []string{
				"key:",
				"  - nestedKey1: value1",
				"  - nestedKey2: value2",
			},
			expected: map[string]any{
				"key": []any{map[string]any{"nestedKey1": "value1"}, map[string]any{"nestedKey2": "value2"}},
			},
			err: false,
		},
		{
			name: "Nested Map with List",
			input: []string{
				"key:",
				"  nestedKey:",
				"    - item1",
				"    - item2",
			},
			expected: map[string]any{
				"key": map[string]any{"nestedKey": []any{"item1", "item2"}},
			},
			err: false,
		},
	}

	for _, tc := range testCases {
		result, err := parseNested(tc.input, 0)
		if tc.err {
			assert.NotNil(t, err, tc.name)
		} else {
			assert.Nil(t, err, tc.name)
			assert.Equal(t, tc.expected, result, tc.name)
		}
	}
}*/
