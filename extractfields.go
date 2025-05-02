package twist

import (
	"fmt"
	"strings"
	"unicode"
)

func extractFields(s string) ([2][]StrPart, error) {
	var fields []StrPart = []StrPart{}
	var pretext []StrPart = []StrPart{}
	nilResult := [2][]StrPart{{}, {}}
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

		field := MustNewStrPart(s, start+len(delimitStart)+offset, end+offset).TrimSpace()
		if valid, reason := isValidField(field.String()); !valid {
			return nilResult, fmt.Errorf("%s: %w", reason, ErrInvalidField)
		}
		fields = append(fields, field)
		pretext = append(pretext, MustNewStrPart(s, offset, offset+start))

		offset += end + len(delimitEnd)
		currentString = currentString[end+len(delimitEnd):]
	}
	pretext = append(pretext, MustNewStrPart(s, offset, len(s)))
	return [2][]StrPart{fields, pretext}, nil
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
