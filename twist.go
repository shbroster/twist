package twist

import (
	"fmt"
	"reflect"
	"strings"
)

type twist struct {
	original     string
	fieldParts   []StrPart
	pretextParts []StrPart
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

func (t twist) execute(data any) (string, error) {
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
				return "", fmt.Errorf("field '%s' is missing: %w", field, ErrInvalidData)
			}
			stringValue, err := toString(value.Interface())
			if err != nil {
				return "", fmt.Errorf("field '%s' is not stringable: %w", field, ErrInvalidData)
			}
			dataMap[field] = stringValue
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			val := v.MapIndex(key)
			stringValue, err := toString(val.Interface())
			if err != nil {
				return "", fmt.Errorf("field '%s' is not stringable: %w", key.String(), ErrInvalidData)
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
			return "", fmt.Errorf("field '%s' is missing: %w", field, ErrInvalidData)
		}
		result += fmt.Sprintf("%s%s", pretext[i], dataField)
	}
	result += pretext[len(pretext)-1]
	return result, nil
}

type result struct {
	val [][2]int
	err error
}

func errResult(msg string) result {
	return result{val: nil, err: fmt.Errorf("%s: %w", msg, ErrTemplateMismatch)}
}

func valResult(val [][2]int) result {
	return result{val: val, err: nil}
}

func (t twist) findFieldIndicies(s string) <-chan result {
	runSearch := true
	ch := make(chan result, 3)
	pretext := t.pretext()
	resultCount := 0
	var sEnd int
	var lastPretext string

	// If there are any fields, these will be at least 2 pretexts
	if len(pretext) <= 1 {
		if s == pretext[0] {
			ch <- valResult([][2]int{})
		} else {
			ch <- errResult("strings do not match")
		}
		runSearch = false
	} else {
		// Verify the first pretexts match and then they can be
		firstPretext := pretext[0]
		if firstPretext != s[:len(firstPretext)] {
			ch <- errResult("string start does not match template")
			runSearch = false
		} else {
			// The last pretext can never be part of the match so check that it matches
			// and then exclude from all searches.
			lastPretext = pretext[len(pretext)-1]
			sEnd = len(s) - len(lastPretext)
			if s[sEnd:] != lastPretext {
				ch <- errResult("string end does not match template")
				runSearch = false
			}
		}
	}

	// Function to recursively search for possible pretext matches.
	var search func(start, pretext int, result [][2]int)
	search = func(start, pretextIdx int, result [][2]int) {
		// If we've matched all pretexts, send the result and then return so that we can
		// look for other potential matches.
		if pretextIdx == len(pretext)-1 {
			result[pretextIdx-1][1] = sEnd
			var resultCopy = make([][2]int, len(result))
			copy(resultCopy, result)
			ch <- valResult(resultCopy)
			resultCount++
			return
		}

		for i := 0; i+start <= sEnd; {
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

			// The first match is fixed so don't consider other options
			if pretextIdx == 0 {
				return
			}
		}
	}

	if runSearch {
		go func() {
			search(0, 0, [][2]int{})
			if resultCount < 1 {
				ch <- errResult("string does not match template")
			}
			close(ch)
		}()
	} else {
		close(ch)
	}

	return ch
}
