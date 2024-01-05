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
		{"\t- list item", LineTypeListElement, 1, "list item"},
		{"\tnestedKey: nestedValue", LineTypeKeyValue, 1, "nestedKey: nestedValue"},
		{"\t\t- nestedList", LineTypeListElement, 2, "nestedList"},
		{"unknown", LineTypeUnknown, 0, "unknown"},
	}

	for _, tc := range testCases {
		line := NewLine(tc.input)
		assert.Equal(t, tc.expectedType, line.Type)
		assert.Equal(t, tc.expectedIndent, line.Indent)
		assert.Equal(t, tc.expectedContent, line.ProcessedString)
	}
}
