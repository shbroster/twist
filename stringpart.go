package twist

import (
	"errors"
	"fmt"
	"unicode"
)

// A reference to a subset of a string
type strPart struct {
	original string
	Start    int
	End      int
}

var (
	errInvalidStrPart = errors.New("invalid strPart")
)

// Construct a new strPart
func newStrPart(s string, start, end int) (strPart, error) {
	if start < 0 {
		return strPart{}, fmt.Errorf("start out of bounds: %w", errInvalidStrPart)
	} else if start > end {
		return strPart{}, fmt.Errorf("start greater than end: %w", errInvalidStrPart)
	} else if end > len(s) {
		return strPart{}, fmt.Errorf("end out of bounds: %w", errInvalidStrPart)
	}

	return strPart{original: s, Start: start, End: end}, nil
}

// Construct a new strPart. Panics if arguments are invalid.
func mustNewStrPart(s string, start, end int) strPart {
	new, err := newStrPart(s, start, end)
	if err != nil {
		panic(err)
	}
	return new
}

// Return the substring that strPart refers to
func (p strPart) String() string {
	return p.original[p.Start:p.End]
}

// Construct a new strPart that has it's whitespace trimmed
func (p strPart) TrimSpace() strPart {
	start := p.Start
	end := p.End
	for start < end && unicode.IsSpace(rune(p.original[start])) {
		start++
	}
	for end > start && unicode.IsSpace(rune(p.original[end-1])) {
		end--
	}

	return mustNewStrPart(p.original, start, end)
}

// Compare two strParts to see if their strings are equal
func (p strPart) Matches(other strPart) bool {
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
