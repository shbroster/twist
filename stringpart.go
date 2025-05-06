package twist

import (
	"errors"
	"fmt"
	"unicode"
)

// A reference to a subset of a string
type strPart struct {
	original string
	start    int
	end      int
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

	return strPart{original: s, start: start, end: end}, nil
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
	return p.original[p.start:p.end]
}

// Construct a new strPart that has it's whitespace trimmed
func (p strPart) TrimSpace() strPart {
	start := p.start
	end := p.end
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
	if p.end-p.start != other.end-other.start {
		return false
	}
	for i := 0; i < p.end-p.start; i++ {
		if p.original[p.start+i] != other.original[other.start+i] {
			return false
		}
	}
	return true
}
