package twist

import (
	"errors"
	"fmt"
	"unicode"
)

// A reference to a subset of a string
type StrPart struct {
	original string
	Start    int
	End      int
}

var (
	ErrInvalidStrPart = errors.New("Invalid StrPart")
)

// Construct a new StrPart
func NewStrPart(s string, start, end int) (StrPart, error) {
	if start < 0 {
		return StrPart{}, fmt.Errorf("start out of bounds: %w", ErrInvalidStrPart)
	} else if start > end {
		return StrPart{}, fmt.Errorf("start greater than end: %w", ErrInvalidStrPart)
	} else if end > len(s) {
		return StrPart{}, fmt.Errorf("end out of bounds: %w", ErrInvalidStrPart)
	}

	return StrPart{original: s, Start: start, End: end}, nil
}

// Construct a new StrPart. Panics if arguments are invalid.
func MustNewStrPart(s string, start, end int) StrPart {
	new, err := NewStrPart(s, start, end)
	if err != nil {
		panic(err)
	}
	return new
}

// Return the substring that StrPart refers to
func (p StrPart) String() string {
	return p.original[p.Start:p.End]
}

// Construct a new StrPart that has it's whitespace trimmed
func (p StrPart) TrimSpace() StrPart {
	start := p.Start
	end := p.End
	for start < end && unicode.IsSpace(rune(p.original[start])) {
		start++
	}
	for end > start && unicode.IsSpace(rune(p.original[end-1])) {
		end--
	}

	return MustNewStrPart(p.original, start, end)
}

// Compare two StrParts to see if their strings are equal
func (p StrPart) Matches(other StrPart) bool {
	if p.End-p.Start != other.End-other.Start {
		return false
	}
	for i := 0; i < p.End-p.Start; i++ {
		if p.original[p.Start+i] != other.original[other.Start+i] {
			return false
		}
	}
	return true
}
