package yamlx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type SimpleStruct struct {
	Key1 string `yamlx:"key1"`
	Key2 string `yamlx:"key2"`
}

func TestUnmarshalSimple(t *testing.T) {
	yamlContent := `
key1: value1
key2: value2
`
	var result SimpleStruct
	err := Unmarshal([]byte(yamlContent), &result)

	expected := SimpleStruct{
		Key1: "value1",
		Key2: "value2",
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

type NestedStruct struct {
	MainKey struct {
		NestedKey string `yamlx:"nestedKey"`
	} `yamlx:"mainKey"`
}

func TestUnmarshalNestedStruct(t *testing.T) {
	yamlContent := `
mainKey:
  nestedKey: nestedValue
`
	var result NestedStruct
	err := Unmarshal([]byte(yamlContent), &result)

	expected := NestedStruct{}
	expected.MainKey.NestedKey = "nestedValue"

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

type TypeConversionStruct struct {
	IntegerField int    `yamlx:"integerField"`
	BooleanField bool   `yamlx:"booleanField"`
	StringField  string `yamlx:"stringField"`
}

func TestUnmarshalTypeConversion(t *testing.T) {
	yamlContent := `
integerField: 42
booleanField: true
stringField: "Hello"
`
	var result TypeConversionStruct
	err := Unmarshal([]byte(yamlContent), &result)

	expected := TypeConversionStruct{
		IntegerField: 42,
		BooleanField: true,
		StringField:  "Hello",
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

type ListStruct struct {
	ListField []string `yamlx:"listField"`
}

func TestUnmarshalList(t *testing.T) {
	yamlContent := `
listField:
  - item1
  - item2
    `
	var result ListStruct
	err := Unmarshal([]byte(yamlContent), &result)

	expected := ListStruct{
		ListField: []string{"item1", "item2"},
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

type MissingFieldStruct struct {
	ExistingField string `yamlx:"existingField"`
	MissingField  string `yamlx:"missingField"`
}

func TestUnmarshalMissingField(t *testing.T) {
	yamlContent := `
existingField: value
`
	var result MissingFieldStruct
	err := Unmarshal([]byte(yamlContent), &result)

	expected := MissingFieldStruct{
		ExistingField: "value",
		MissingField:  "",
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
