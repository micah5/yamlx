package yamlx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLineParsing(t *testing.T) {
	testCases := []struct {
		input           string
		expectedType    LineType
		expectedIndent  int
		expectedContent string
	}{
		{"key: value", LineTypeKeyValue, 0, "key: value"},
		{"- list item", LineTypeListElement, 0, "list item"},
		{"  nestedKey: nestedValue", LineTypeKeyValue, 2, "nestedKey: nestedValue"},
		{"  - nestedList", LineTypeListElement, 2, "nestedList"},
		{"unknown", LineTypeUnknown, 0, "unknown"},
	}

	for _, tc := range testCases {
		line := NewLine(tc.input, " ")
		assert.Equal(t, tc.expectedType, line.Type)
		assert.Equal(t, tc.expectedIndent, line.Indent)
		assert.Equal(t, tc.expectedContent, line.ProcessedString)
	}
}

func TestIndentationConsistency(t *testing.T) {
	testCases := []struct {
		input           string
		indentationType string
		expectedIndent  int
		expectedError   bool
	}{
		{"  key: value", " ", 2, false},
		{"    key: value", " ", 4, false},
		{"\tkey: value", "\t", 1, false},
		{"\t\tkey: value", "\t", 2, false},
		{"  key: value", "\t", 0, true},
		{"\tkey:\n\t\tkey: value", "\t", 1, false},
		{"\tkey:\n  key: value", "\t", 0, true},
	}

	for _, tc := range testCases {
		indentationType, err := getIndentationType([]string{tc.input})
		assert.Equal(t, tc.indentationType, indentationType)
		if tc.expectedError {
			assert.Error(t, err)
		} else {
			indent := getIndent(tc.input, tc.indentationType)
			assert.Equal(t, tc.expectedIndent, indent)
		}
	}
}
