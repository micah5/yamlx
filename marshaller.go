package yamlx

import (
	"errors"
	"gopkg.in/yaml.v3"
	"reflect"
	"strings"
)

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
