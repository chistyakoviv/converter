package mapper

import (
	"fmt"
	"reflect"
	"strings"
)

func MapToStruct(input map[string]interface{}, output interface{}) error {
	// Ensure output is a pointer to a struct
	val := reflect.ValueOf(output)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("output must be a pointer to a struct")
	}

	// Dereference the pointer to get the struct value
	structVal := val.Elem()
	structType := structVal.Type()

	// Create a mapping of lowercase field names to actual struct fields
	fieldMap := make(map[string]int)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldMap[strings.ToLower(field.Name)] = i
	}

	// Iterate over the map and assign matching fields
	for key, value := range input {
		lowerKey := strings.ToLower(key)

		// Check if the struct has a field matching the lowercase key
		if fieldIndex, ok := fieldMap[lowerKey]; ok {
			field := structVal.Field(fieldIndex)
			if !field.CanSet() {
				continue // Skip if the field can't be set
			}

			// Convert the value to the field's type
			fieldType := field.Type()
			val := reflect.ValueOf(value)

			// Ignore not convertible types
			if val.Type().ConvertibleTo(fieldType) {
				field.Set(val.Convert(fieldType))
			}
		}
	}

	return nil
}
