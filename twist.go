package twist

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

var delimitStart = "{{"
var delimitEnd = "}}"

var (
	ErrInvalidTemplate = errors.New("invalid template")
	ErrInvalidField    = errors.New("invalid field")
	ErrInvalidData     = errors.New("invalid data")
)

// twists are reversable templates that can be used to create basic string template
// using {{ and }} as delimeters.
//
// Supports the following operations;
// - Execute; create a string template given the supplied data,
// - Parse; parse a populated string template and extract data into a struct
type twist struct {
	original string
	fields   []string
	pretext  []string
}

// New creates a 'twist', errors if the template is invald.
func New(s string) (twist, error) {
	result, err := extractFields(s)
	if err != nil {
		return twist{}, err
	}
	return twist{
		original: s,
		fields:   result[0],
		pretext:  result[1],
	}, nil
}

func extractFields(s string) ([2][]string, error) {
	var fields []string = []string{}
	var pretext []string = []string{}

	for {
		start := strings.Index(s, delimitStart)
		end := strings.Index(s, delimitEnd)
		nextStart := -1
		if end != -1 {
			offset := start + len(delimitStart)
			if index := strings.Index(s[offset:], delimitStart); index == -1 {
				nextStart = index
			} else {
				nextStart = index + offset
			}
		}
		if start == -1 && end == -1 {
			break
		} else if start == -1 || end < start {
			return [2][]string{{}, {}}, fmt.Errorf("unmatched delimiters: %w", ErrInvalidTemplate)
		} else if nextStart != -1 && nextStart < end {
			return [2][]string{{}, {}}, fmt.Errorf("nested delimiters: %w", ErrInvalidTemplate)
		}

		field := strings.TrimSpace(s[start+len(delimitStart) : end])
		if valid, reason := isValidField(field); !valid {
			return [2][]string{{}, {}}, fmt.Errorf("%s: %w", reason, ErrInvalidField)
		}
		fields = append(fields, field)
		pretext = append(pretext, s[:start])
		s = s[end+len(delimitEnd):]
	}
	return [2][]string{fields, pretext}, nil
}

func isValidField(field string) (bool, string) {
	if len(field) == 0 {
		return false, "must not be empty"
	}
	r := rune(field[0])
	if !(unicode.IsUpper(r)) {
		return false, "must start with an uppercase letter"
	}
	for _, r := range field {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return false, "must contain only letters, digits, and underscores"
		}
	}
	return true, ""
}

// Execute executes the template with the given data and returns the generated string.
func (t twist) Execute(data any) (string, error) {

	// Convert data to a map[string]string
	dataMap := make(map[string]string)

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		for _, field := range t.fields {
			value := v.FieldByName(field)
			if !value.IsValid() {
				return "", fmt.Errorf("field '%s' is missing: %w", field, ErrInvalidField)
			}
			stringValue, err := toString(value.Interface())
			if err != nil {
				return "", fmt.Errorf("field '%s' is not stringable: %w", field, ErrInvalidField)
			}
			dataMap[field] = stringValue
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			val := v.MapIndex(key)
			stringValue, err := toString(val.Interface())
			if err != nil {
				return "", fmt.Errorf("field '%s' is not stringable: %w", key.String(), ErrInvalidField)
			}
			dataMap[key.String()] = stringValue
		}
	default:
		return "", fmt.Errorf("data is not a struct or map: %w", ErrInvalidData)
	}

	// Construct the result string
	var result string
	for i, field := range t.fields {
		// access a variable dynamically from any object of type any
		dataField, ok := dataMap[field]
		if !ok {
			return "", fmt.Errorf("field '%s' is missing: %w", field, ErrInvalidField)
		}
		result += fmt.Sprintf("%s%s", t.pretext[i], dataField)
	}

	return result, nil
}

func toString(v interface{}) (string, error) {
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		v = reflect.ValueOf(v).Elem().Interface()
	}
	switch val := v.(type) {
	case string:
		return val, nil
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		bool:
		return fmt.Sprintf("%v", v), nil
	case fmt.Stringer:
		return val.String(), nil

	default:
		return "", errors.New("value cannot be converted to string")
	}
}
