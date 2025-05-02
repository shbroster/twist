package twist

import (
	"errors"
	"testing"
)

func TestStrPartSucess(t *testing.T) {
	type testCase struct {
		name        string
		original    string
		wantString  string
		wantTrimmed string
	}

	testCases := []testCase{
		{
			name:        "Hello, World!",
			original:    "Hello, World!",
			wantString:  "Hello, World!",
			wantTrimmed: "Hello, World!",
		},
		{
			name:        "empty",
			original:    "",
			wantString:  "",
			wantTrimmed: "",
		},
		{
			name:        "only spaces",
			original:    "     ",
			wantString:  "     ",
			wantTrimmed: "",
		},
		{
			name:        "space at start",
			original:    "   Hello, World!",
			wantString:  "   Hello, World!",
			wantTrimmed: "Hello, World!",
		},
		{
			name:        "space at end",
			original:    "Hello, World!   ",
			wantString:  "Hello, World!   ",
			wantTrimmed: "Hello, World!",
		},
		{
			name:        "space at start and end",
			original:    "   Hello, World!   ",
			wantString:  "   Hello, World!   ",
			wantTrimmed: "Hello, World!",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			original := testCase.original
			part, err := NewStrPart(original, 0, len(original))
			if err != nil {
				t.Errorf("Invalid args: %v", err)
				return
			}
			if part.String() != testCase.wantString {
				t.Errorf("Expected '%s', got '%s'", testCase.wantString, part.String())
				return
			}
			if part.TrimSpace().String() != testCase.wantTrimmed {
				t.Errorf("Expected '%s', got '%s'", testCase.wantTrimmed, part.String())
				return
			}
		})
	}
}

func TestStrPartFailure(t *testing.T) {
	type testCase struct {
		name     string
		original string
		start    int
		end      int
	}

	testCases := []testCase{
		{
			name:     "negative start",
			original: "Hello, World!",
			start:    -1,
			end:      13,
		},
		{
			name:     "start greater than end",
			original: "Hello, World!",
			start:    14,
			end:      13,
		},
		{
			name:     "end greater than total length",
			original: "Hello, World!",
			start:    0,
			end:      14,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := NewStrPart(testCase.original, testCase.start, testCase.end)
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}
			if !errors.Is(err, ErrInvalidStrPart) {
				t.Errorf("Expected error %v, got %v", ErrInvalidStrPart, err)
				return
			}
		})
	}
}

func TestStrPartEquals(t *testing.T) {
	type testCase struct {
		name     string
		original StrPart
		other    StrPart
		want     bool
	}

	testCases := []testCase{
		{
			name:     "exactly equal",
			original: MustNewStrPart("Hello, World!", 0, 13),
			other:    MustNewStrPart("Hello, World!", 0, 13),
			want:     true,
		},
		{
			name:     "partially equal",
			original: MustNewStrPart("Hello, World!", 0, 13),
			other:    MustNewStrPart(" Hello, World! ", 1, 14),
			want:     true,
		},
		{
			name:     "empty",
			original: MustNewStrPart("", 0, 0),
			other:    MustNewStrPart("", 0, 0),
			want:     true,
		},
		{
			name:     "empty other",
			original: MustNewStrPart("Hello, World!", 0, 13),
			other:    MustNewStrPart("", 0, 0),
			want:     false,
		},
		{
			name:     "different",
			original: MustNewStrPart("Hello, World!", 0, 13),
			other:    MustNewStrPart("Hello, Universe!", 0, 13),
			want:     false,
		},
		{
			name:     "longer",
			original: MustNewStrPart("Hello, World!", 0, 2),
			other:    MustNewStrPart("Hello, World!", 0, 3),
			want:     false,
		},
		{
			name:     "shorter",
			original: MustNewStrPart("Hello, World!", 0, 3),
			other:    MustNewStrPart("Hello, World!", 0, 2),
			want:     false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.original.Matches(testCase.other)
			if result != testCase.want {
				t.Errorf("Incorrect match result")
				return
			}
		})
	}

}
