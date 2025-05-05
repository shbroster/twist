package twist

import (
	"fmt"
	"reflect"
	"strconv"
)

func decode(input map[string]string, out any) error {
	// Validate that 'out' is a pointer to a struct
	outVal, err := validateOut(out)
	if err != nil {
		return err
	}

	// Write fields to the data struct and convert to the correct type
	for key, value := range input {
		field := outVal.FieldByName(key)
		if !field.IsValid() {
			return fmt.Errorf("field '%s' is missing: %w", key, ErrInvalidData)
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(value)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			intValue, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("field '%s' cannot be converted to supplied type: %w", key, ErrInvalidData)
			}
			switch field.Kind() {
			case reflect.Int:
				field.SetInt(int64(intValue))
			case reflect.Int8:
				field.SetInt(int64(intValue))
			case reflect.Int16:
				field.SetInt(int64(intValue))
			case reflect.Int32:
				field.SetInt(int64(intValue))
			case reflect.Int64:
				field.SetInt(int64(intValue))
			case reflect.Uint:
				field.SetUint(uint64(intValue))
			case reflect.Uint8:
				field.SetUint(uint64(intValue))
			case reflect.Uint16:
				field.SetUint(uint64(intValue))
			case reflect.Uint32:
				field.SetUint(uint64(intValue))
			case reflect.Uint64:
				field.SetUint(uint64(intValue))
			}

		case reflect.Bool:
			boolValue, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("field '%s' cannot be converted to supplied type: %w", key, ErrInvalidData)
			}
			field.SetBool(boolValue)

		case reflect.Float64, reflect.Float32:
			bitSize := 64
			if field.Kind() == reflect.Float32 {
				bitSize = 32
			}
			floatValue, err := strconv.ParseFloat(value, bitSize)
			if err != nil {
				return fmt.Errorf("field '%s' cannot be converted to supplied type: %w", key, ErrInvalidData)
			}
			field.SetFloat(floatValue)

		default:
			return fmt.Errorf("field '%s' is not a supported type: %w", key, ErrInvalidData)
		}
	}
	return nil
}

func validateOut(out any) (reflect.Value, error) {
	outVal := reflect.ValueOf(out)
	if kind := outVal.Kind(); kind != reflect.Ptr || outVal.IsNil() {
		return reflect.Value{}, fmt.Errorf("out must be a non-nil pointer but got '%s': %w", kind, ErrInvalidData)
	}
	for outVal.Kind() == reflect.Ptr || outVal.Kind() == reflect.Interface {
		outVal = outVal.Elem()
		if !outVal.IsValid() {
			return reflect.Value{}, fmt.Errorf("out must point to a valid struct: %w", ErrInvalidData)
		}
	}
	if outVal.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("out must point to a struct but got '%s': %w", outVal.Kind(), ErrInvalidData)
	}
	return outVal, nil
}
