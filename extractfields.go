package twist

import (
	"fmt"
	"strings"
	"unicode"
)

func extractFields(s string, delimiters [2]string) ([2][]strPart, error) {
	var fields []strPart = []strPart{}
	var pretext []strPart = []strPart{}
	delimitStart := delimiters[0]
	delimitEnd := delimiters[1]
	nilResult := [2][]strPart{{}, {}}
	currentString := s

	offset := 0
	for {
		start := strings.Index(currentString, delimitStart)
		end := strings.Index(currentString, delimitEnd)
		nextStart := -1
		if end != -1 {
			offset := start + len(delimitStart)
			if index := strings.Index(currentString[offset:], delimitStart); index == -1 {
				nextStart = index
			} else {
				nextStart = index + offset
			}
		}
		if start == -1 && end == -1 {
			break
		} else if start == -1 || end < start {
			return nilResult, fmt.Errorf("unmatched delimiters: %w", ErrInvalidTemplate)
		} else if nextStart != -1 && nextStart < end {
			return nilResult, fmt.Errorf("nested delimiters: %w", ErrInvalidTemplate)
		}

		field := mustNewStrPart(s, start+len(delimitStart)+offset, end+offset).TrimSpace()
		if valid, reason := isValidField(field.String()); !valid {
			return nilResult, fmt.Errorf("%s: %w", reason, ErrInvalidTemplate)
		}
		fields = append(fields, field)
		pretext = append(pretext, mustNewStrPart(s, offset, offset+start))

		offset += end + len(delimitEnd)
		currentString = currentString[end+len(delimitEnd):]
	}
	pretext = append(pretext, mustNewStrPart(s, offset, len(s)))
	return [2][]strPart{fields, pretext}, nil
}

func isValidField(field string) (bool, string) {
	if len(field) == 0 {
		return false, "field must not be empty"
	}
	r := rune(field[0])
	if !(unicode.IsUpper(r)) {
		return false, "field must start with an uppercase letter"
	}
	for _, r := range field {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return false, "field must contain only letters, digits, and underscores"
		}
	}
	return true, ""
}
