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
		expectedAnchor  *Anchor
		expectedError   bool
	}{
		{"key:", LineTypeKey, 0, "key", nil, false},
		{"key: value", LineTypeKeyValue, 0, "key: value", nil, false},
		{"\t- list item", LineTypeListElement, 1, "list item", nil, false},
		{"\tnestedKey: nestedValue", LineTypeKeyValue, 1, "nestedKey: nestedValue", nil, false},
		{"\t\t- nestedList", LineTypeListElement, 2, "nestedList", nil, false},
		{"key: &anchor_name value", LineTypeKeyValue, 0, "key: value", &Anchor{Name: "anchor_name"}, false},
		{"\t- &anchor_name list item", LineTypeListElement, 1, "list item", &Anchor{Name: "anchor_name"}, false},
		{"key: &anchor_name", LineTypeKey, 0, "key", &Anchor{Name: "anchor_name"}, false},
		{"unknown", LineTypeKey, 0, "", nil, true},
	}

	for _, tc := range testCases {
		line, err := NewLine(tc.input)
		if tc.expectedError {
			assert.Error(t, err)
		} else {
			assert.Equal(t, tc.expectedType, line.Type)
			assert.Equal(t, tc.expectedIndent, line.Indent)
			assert.Equal(t, tc.expectedContent, line.ProcessedString)
			assert.Equal(t, tc.expectedAnchor, line.Anchor)
		}
	}
}
