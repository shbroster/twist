package twist

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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
	original     string
	fieldParts   []StrPart
	pretextParts []StrPart
}

// New creates a 'twist', errors if the template is invald.
func New(s string) (twist, error) {
	result, err := extractFields(s)
	if err != nil {
		return twist{}, err
	}
	return twist{
		original:     s,
		fieldParts:   result[0],
		pretextParts: result[1],
	}, nil
}

func (t twist) fields() []string {
	result := make([]string, len(t.fieldParts))
	for i, p := range t.fieldParts {
		result[i] = p.String()
	}
	return result
}

func (t twist) pretext() []string {
	result := make([]string, len(t.pretextParts))
	for i, p := range t.pretextParts {
		result[i] = p.String()
	}
	return result
}

// Execute executes the template with the given data and returns the generated string.
func (t twist) Execute(data any) (string, error) {
	fields := t.fields()
	pretext := t.pretext()

	// Convert data to a map[string]string
	dataMap := make(map[string]string)

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		for _, field := range fields {
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
	for i, field := range fields {
		// access a variable dynamically from any object of type any
		dataField, ok := dataMap[field]
		if !ok {
			return "", fmt.Errorf("field '%s' is missing: %w", field, ErrInvalidField)
		}
		result += fmt.Sprintf("%s%s", pretext[i], dataField)
	}
	result += pretext[len(pretext)-1]
	// TODO: error if not reversible!!
	return result, nil
}

func (t twist) findFieldIndicies(s string) ([][][2]int, error) {
	results := [][][2]int{}
	pretext := t.pretext()

	// If there are any fields, these will be at least 2 pretexts
	if len(pretext) <= 1 {
		if s == pretext[0] {
			return [][][2]int{{}}, nil
		}
		return nil, fmt.Errorf("string start does not match template: %w", ErrTemplateMismatch)
	}

	// The last pretext can never be part of the match so check that it matches
	// and then it can be excluded from all searches.
	lastPretext := pretext[len(pretext)-1]
	sEnd := len(s) - len(lastPretext)
	if s[sEnd:] != lastPretext {
		return nil, fmt.Errorf("string end does not match template: %w", ErrTemplateMismatch)
	}

	// Function to recursively search for possible pretext mathces.
	var search func(start, pretext int, result [][2]int)
	search = func(start, pretextIdx int, result [][2]int) {
		// If we've matches all pretexts, save the result and then return so that we can
		// look for other potential matches.
		if pretextIdx == len(pretext)-1 {
			result[pretextIdx-1][1] = sEnd
			var resultCopy = make([][2]int, len(result))
			copy(resultCopy, result)
			results = append(results, resultCopy)
			return
		}

		for i := 0; ; {
			pretextStr := pretext[pretextIdx]
			offset := start + i
			match := strings.Index(s[offset:sEnd], pretextStr)
			i = i + match + 1

			// Stop searching this path
			if match == -1 {
				return
			}

			// Store the start of this match
			indexStart := match + offset + len(pretextStr)
			result = append(result, [2]int{indexStart, 0})

			// Store the end for the previous match
			if pretextIdx > 0 {
				result[pretextIdx-1][1] = match + offset
			}

			// Search for the next pretext
			search(indexStart, pretextIdx+1, result)
			result = result[:len(result)-1]

			// The first match is fixed so don't conisder other options
			if pretextIdx == 0 {
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
	for i, field := range t.fields() {
		result[field] = s[indicies[0][i][0]:indicies[0][i][1]]
	}
	return result, nil
}

// Parse takes a string generated by executing a template and returns the original data
// that was used when executing the template.
//
// The parsed data is cast to the appropriate type and stored in the provided struct
func (t twist) Parse(s string, out any) error {
	// Validate data is a pointer to a struct
	outVal := reflect.ValueOf(out)
	if kind := outVal.Kind(); kind != reflect.Ptr {
		return fmt.Errorf("out must be a pointer but got '%s': %w", kind, ErrInvalidArgument)
	}
	outVal = outVal.Elem()
	if outVal.Kind() == reflect.Interface {
		outVal = outVal.Elem()
	}
	if outVal.Kind() != reflect.Struct {
		return fmt.Errorf("out must point to a struct but got: %w", ErrInvalidArgument)
	}

	result, err := t.ParseFields(s)
	if err != nil {
		return err
	}

	// Write fields to the data struct and convert to the correct type
	for key, value := range result {
		val := reflect.ValueOf(out)
		fmt.Printf("reflect.ValueOf(out): Kind=%s, Type=%s\n", val.Kind(), val.Type())
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
			fmt.Printf("After .Elem(): Kind=%s, Type=%s\n", val.Kind(), val.Type())
		}
		for val.Kind() == reflect.Interface {
			val = val.Elem()
		}
		field := val.FieldByName(key)
		fmt.Printf("FieldByName(%q): IsValid=%v, CanSet=%v, Kind=%s\n", key, field.IsValid(), field.CanSet(), field.Kind())

		if !field.IsValid() {
			return fmt.Errorf("field '%s' is missing: %w", key, ErrInvalidData)
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(value)
		case reflect.Int:
			intValue, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("field '%s' cannot be converted to supplied type: %w", key, ErrInvalidField)
			}
			field.SetInt(int64(intValue))
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("field '%s' cannot be converted to supplied type: %w", key, ErrInvalidField)
			}
			field.SetBool(boolValue)
		case reflect.Float64:
			boolValue, err := strconv.ParseFloat(value, 10)
			if err != nil {
				return fmt.Errorf("field '%s' cannot be converted to supplied type: %w", key, ErrInvalidField)
			}
			field.SetFloat(boolValue)
		default:
			return fmt.Errorf("field '%s' is not a supported type: %w", key, ErrInvalidData)
		}
	}
	return nil
}
