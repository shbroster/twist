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
	ErrInvalidTemplate  = errors.New("invalid template")
	ErrInvalidField     = errors.New("invalid field")
	ErrInvalidData      = errors.New("invalid data")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrTemplateMismatch = errors.New("template mismatch")
)

// twists are reversable templates that can be used to create basic string template
// using {{ and }} as delimeters.
//
// Supports the following operations;
// - Execute; create a string template given the supplied data,
// - Parse; parse a populated string template and extract data into a struct
// - TODO...
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
	pretext = append(pretext, s)
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
	result += t.pretext[len(t.pretext)-1]
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

func (t twist) findFieldIndicies(s string) ([][][2]int, error) {
	results := [][][2]int{}

	// If there are any fields, these will be at least 2 pretexts
	if len(t.pretext) <= 1 {
		if s == t.pretext[0] {
			return [][][2]int{{}}, nil
		}
		return nil, fmt.Errorf("string does not match template: %w", ErrTemplateMismatch)
	}

	// The last pretext can never be part of the match so check that it matches
	// and then it can be excluded from all searches.
	lastPretext := t.pretext[len(t.pretext)-1]
	sEnd := len(s) - len(lastPretext)
	if s[sEnd:] != lastPretext {
		return nil, fmt.Errorf("string does not match template: %w", ErrTemplateMismatch)
	}

	// Function to recursively search for possible pretext mathces.
	var search func(start, pretext int, result [][2]int)
	search = func(start, pretext int, result [][2]int) {
		// If we've matches all pretexts, save the result and then return so that we can
		// look for other potential matches.
		if pretext == len(t.pretext)-1 {
			result[pretext-1][1] = sEnd
			var resultCopy = make([][2]int, len(result))
			copy(resultCopy, result)
			results = append(results, resultCopy)
			return
		}

		for i := 0; ; {
			pretextStr := t.pretext[pretext]
			offset := start + i
			match := strings.Index(s[offset:sEnd], pretextStr)
			i = i + match + 1

			// Stop searching this path
			if match == -1 {
				return
			}

			// Get start of this match
			indexStart := match + offset + len(pretextStr)
			result = append(result, [2]int{indexStart, 0})

			// Update the end for the previous match
			if pretext > 0 {
				result[pretext-1][1] = match + offset
			}

			// Search for the next pretext
			search(indexStart, pretext+1, result)
			result = result[:len(result)-1]

			// The first match is fixed so don't conisder other options
			if pretext == 0 {
				return
			}
		}
	}

	search(0, 0, [][2]int{})

	if len(results) == 0 {
		return nil, fmt.Errorf("string does not match template: %w", ErrTemplateMismatch)
	}
	return results, nil
}

// ParseFields takes a string generated by executing a template and returns the original data
// that was used when executing the template.
//
// It is possible to construct templates where there are multiple possible ways that the data
// could be extraced, the parsing algorithm extracts data in a non-greedy fashion. For example
// the template '{{Name}} {Age}' and the string 'John Smith 23' would be parsed as Name="John"
// & Age="Smith 23" which is not necesarily the result you would expect.
func (t twist) ParseFields(s string) (map[string]string, error) {
	indicies, err := t.findFieldIndicies(s)
	if err != nil {
		return nil, err
	}

	if len(indicies) != 1 {
		return nil, fmt.Errorf("multiple matches: %w", ErrTemplateMismatch)
	}

	result := map[string]string{}
	for i, field := range t.fields {
		result[field] = s[indicies[0][i][0]:indicies[0][i][1]]
	}
	return result, nil
}

// 	// find the indices of the fields in the string
// 	indices := make([][2]int, len(t.fields))
// 	lastIndex := 0
// 	for i, text := range t.pretext {
// 		match := strings.Index(s[lastIndex:], text)
// 		if match == -1 {
// 			return nil, fmt.Errorf("string does not match template: %w", ErrTemplateMismatch)
// 		}
// 		lastIndex = match + len(text) + lastIndex

// 		if i < len(indices) {
// 			indices[i][0] = lastIndex
// 		}
// 		if i > 0 {
// 			var index int
// 			if i == len(t.pretext)-1 && text == "" {
// 				index = len(s)
// 			} else {
// 				index = lastIndex - len(text)
// 			}
// 			indices[i-1][1] = index
// 		}
// 	}
// }

// Parse takes a string generated by executing a template and returns the original data
// that was used when executing the template.
//
// The parsed data is cast to the appropriate type and stored in the provided struct
// func (t twist) Parse(s string, out any) error {
// 	// Validate data is a pointer to a struct
// 	outVal := reflect.ValueOf(out)
// 	if kind := outVal.Kind(); kind != reflect.Ptr {
// 		return fmt.Errorf("out must be a pointer but got '%s': %w", kind, ErrInvalidArgument)
// 	}
// 	outVal = outVal.Elem()
// 	if outVal.Kind() == reflect.Interface {
// 		outVal = outVal.Elem()
// 	}
// 	if outVal.Kind() != reflect.Struct {
// 		return fmt.Errorf("out must point to a struct but got: %w", ErrInvalidArgument)
// 	}

// 	result, err := t.ParseFields(s)
// 	if err != nil {
// 		return err
// 	}

// 	// Write fields to the data struct and convert to the correct type
// 	for key, value := range result {
// 		val := reflect.ValueOf(out)
// 		fmt.Printf("reflect.ValueOf(out): Kind=%s, Type=%s\n", val.Kind(), val.Type())
// 		if val.Kind() == reflect.Ptr {
// 			val = val.Elem()
// 			fmt.Printf("After .Elem(): Kind=%s, Type=%s\n", val.Kind(), val.Type())
// 		}
// 		for val.Kind() == reflect.Interface {
// 			val = val.Elem()
// 		}
// 		field := val.FieldByName(key)
// 		fmt.Printf("FieldByName(%q): IsValid=%v, CanSet=%v, Kind=%s\n", key, field.IsValid(), field.CanSet(), field.Kind())

// 		if !field.IsValid() {
// 			return fmt.Errorf("field '%s' is missing: %w", key, ErrInvalidData)
// 		}

// 		switch field.Kind() {
// 		case reflect.String:
// 			field.SetString(value)
// 		case reflect.Int:
// 			intValue, err := strconv.Atoi(value)
// 			if err != nil {
// 				return fmt.Errorf("field '%s' cannot be converted to supplied type: %w", key, ErrInvalidField)
// 			}
// 			field.SetInt(int64(intValue))
// 		case reflect.Bool:
// 			boolValue, err := strconv.ParseBool(value)
// 			if err != nil {
// 				return fmt.Errorf("field '%s' cannot be converted to supplied type: %w", key, ErrInvalidField)
// 			}
// 			field.SetBool(boolValue)
// 		case reflect.Float64:
// 			boolValue, err := strconv.ParseFloat(value, 10)
// 			if err != nil {
// 				return fmt.Errorf("field '%s' cannot be converted to supplied type: %w", key, ErrInvalidField)
// 			}
// 			field.SetFloat(boolValue)
// 		default:
// 			return fmt.Errorf("field '%s' is not a supported type: %w", key, ErrInvalidData)
// 		}
// 	}
// 	return nil
// }
