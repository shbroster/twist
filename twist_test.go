package twist

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestTwistExceuteSuccess(t *testing.T) {
	type testCase struct {
		name     string
		template string
		data     any
		want     string
	}

	timeNow := time.Now()
	hello := "hello"
	data := map[string]any{"Name": "World"}

	tests := []testCase{
		{
			name:     "basic",
			template: "Hello, {{Name}}",
			data:     data,
			want:     "Hello, World",
		},
		{
			name:     "basic 2",
			template: "Hello, {{Name}}",
			data:     &data,
			want:     "Hello, World",
		},
		{
			name:     "multiple",
			template: "Hello, {{ Name}} {{ Age}}",
			data:     map[string]any{"Name": "World", "Age": 25},
			want:     "Hello, World 25",
		},
		{
			name:     "empty",
			template: "",
			data:     map[string]any{},
			want:     "",
		},
		{
			name:     "struct",
			template: "Hello, {{ Name}} {{ Age}}",
			data: struct {
				Name string
				Age  int
			}{Name: "World", Age: 25},
			want: "Hello, World 25",
		},
		{
			name:     "numbers",
			template: "{{Int}} {{Int8}} {{Int16}} {{Int32}} {{Int64}} {{Uint}} {{Uint8}} {{Uint16}} {{Uint32}} {{Uint64}} {{Float}}",
			data: struct {
				Int    int
				Int8   int8
				Int16  int16
				Int32  int32
				Int64  int64
				Uint   uint
				Uint8  uint8
				Uint16 uint16
				Uint32 uint32
				Uint64 uint64
				Float  float64
			}{
				Int:    1,
				Int8:   2,
				Int16:  3,
				Int32:  4,
				Int64:  5,
				Uint:   6,
				Uint8:  7,
				Uint16: 8,
				Uint32: 9,
				Uint64: 10,
				Float:  11.1,
			},
			want: "1 2 3 4 5 6 7 8 9 10 11.1",
		},
		{
			name:     "stringer",
			template: "{{Stringer}}",
			data: struct {
				Stringer fmt.Stringer
			}{
				Stringer: timeNow,
			},
			want: timeNow.String(),
		},
		{
			name:     "pointer",
			template: "{{Pointer}}",
			data: struct {
				Pointer *string
			}{
				Pointer: &hello,
			},
			want: "hello",
		},
		{
			name:     "text at end",
			template: "{{Greeting}} World",
			data:     map[string]string{"Greeting": "Hello"},
			want:     "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := New(tt.template)
			if err != nil {
				t.Errorf("New() error: %v", err)
				return
			}
			got, err := tmpl.execute(tt.data)
			if err != nil {
				t.Errorf("execute() error: %v", err)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want)\n%s", diff)
				return
			}
		})
	}
}

func TestTwistExecuteError(t *testing.T) {
	type testCase struct {
		name      string
		template  string
		data      any
		errorType error
		errorMsg  string
	}

	tests := []testCase{
		{
			name:      "wrong data type",
			template:  "Hello, {{Name}}",
			data:      "Some string",
			errorType: ErrInvalidData,
			errorMsg:  "data is not a struct or map",
		},
		{
			name:      "missing field (map)",
			template:  "Hello, {{Name}}",
			data:      map[string]any{},
			errorType: ErrInvalidField,
			errorMsg:  "field 'Name' is missing",
		},
		{
			name:      "not stringable (map)",
			template:  "Hello, {{NotStringable}}",
			data:      map[string]any{"NotStringable": errors.New("Not Stringable")},
			errorType: ErrInvalidField,
			errorMsg:  "field 'NotStringable' is not stringable",
		},
		{
			name:     "missing field (struct)",
			template: "Hello, {{Name}}",
			data: struct {
				OtherName string
			}{},
			errorType: ErrInvalidField,
			errorMsg:  "field 'Name' is missing",
		},
		{
			name:      "not stringable (struct)",
			template:  "Hello, {{NotStringable}}",
			data:      struct{ NotStringable error }{NotStringable: errors.New("Not Stringable")},
			errorType: ErrInvalidField,
			errorMsg:  "field 'NotStringable' is not stringable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := New(tt.template)
			if err != nil {
				t.Errorf("New() error = %v", err)
			}
			_, err = tmpl.execute(tt.data)
			if err == nil {
				t.Errorf("execute() error is nil")
				return
			}
			if !errors.Is(err, tt.errorType) {
				t.Errorf("execute() error type = '%v', want type '%v'", err, tt.errorType)
				return
			}
			if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("execute() error = '%v', want to contain '%v'", err, tt.errorMsg)
				return
			}
		})
	}
}

func TestFindFieldIndiciesSucess(t *testing.T) {
	type testCase struct {
		name     string
		template string
		result   string
		want     [][][2]int
	}

	tests := []testCase{
		{
			name:     "single match at end",
			template: "Hello, {{Name}}",
			result:   "Hello, World",
			want:     [][][2]int{{[2]int{7, 12}}},
		},
		{
			name:     "single match at start",
			template: "{{Hello}}, World",
			result:   "Hello, World",
			want:     [][][2]int{{[2]int{0, 5}}},
		},
		{
			name:     "single match in middle",
			template: "Hello{{Seperator}}World",
			result:   "Hello, World",
			want:     [][][2]int{{[2]int{5, 7}}},
		},
		{
			name:     "single match; nothing else",
			template: "{{Greeting}}",
			result:   "Hello, World",
			want:     [][][2]int{{[2]int{0, 12}}},
		},
		{
			name:     "multiple fields",
			template: "Hello, {{Name}} {{Age}}",
			result:   "Hello, World 25",
			want:     [][][2]int{{[2]int{7, 12}, [2]int{13, 15}}},
		},
		{
			name:     "empty",
			template: "",
			result:   "",
			want:     [][][2]int{{}},
		},
		{
			name:     "no fields",
			template: "template with no fields",
			result:   "template with no fields",
			want:     [][][2]int{{}},
		},
		{
			name:     "mutliple matches, common seperator",
			template: "{{First}}-{{Seconds}}",
			result:   "----",
			want: [][][2]int{
				{[2]int{0, 0}, [2]int{1, 4}},
				{[2]int{0, 1}, [2]int{2, 4}},
				{[2]int{0, 2}, [2]int{3, 4}},
				{[2]int{0, 3}, [2]int{4, 4}},
			},
		},
		{
			name:     "multiple matches, no seperator",
			template: "...{{Name}}{{Age}}...",
			result:   "...12345678...",
			want: [][][2]int{
				{{3, 3}, {3, 11}},
				{{3, 4}, {4, 11}},
				{{3, 5}, {5, 11}},
				{{3, 6}, {6, 11}},
				{{3, 7}, {7, 11}},
				{{3, 8}, {8, 11}},
				{{3, 9}, {9, 11}},
				{{3, 10}, {10, 11}},
				{{3, 11}, {11, 11}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := New(tt.template)
			if err != nil {
				t.Errorf("New() error = %v", err)
				return
			}
			results, err := tmpl.findFieldIndicies(tt.result)
			if err != nil {
				t.Errorf("template mismatch: %v", err)
				return
			}
			if diff := cmp.Diff(results, tt.want); diff != "" {
				t.Errorf("Parse() mismatch (-got +want)\n%s", diff)
				return
			}
		})
	}
}

func TestFindFieldIndiciesError(t *testing.T) {
	type testCase struct {
		name      string
		template  string
		result    string
		errorType error
		errorMsg  string
	}

	tests := []testCase{
		{
			name:      "no match (no fields)",
			template:  "X",
			result:    "a",
			errorType: ErrTemplateMismatch,
			errorMsg:  "strings do not match",
		},
		{
			name:      "no match at start",
			template:  "X{{Field}}Y",
			result:    "YaY",
			errorType: ErrTemplateMismatch,
			errorMsg:  "string start does not match template",
		},
		{
			name:      "no match at end",
			template:  "X{{Field}}Y",
			result:    "XaX",
			errorType: ErrTemplateMismatch,
			errorMsg:  "string end does not match template",
		},
		{
			name:      "no match in middle",
			template:  "X{{Field}}-{{Field}}Y",
			result:    "XaaaY",
			errorType: ErrTemplateMismatch,
			errorMsg:  "string does not match template",
		},
		{
			name:      "empty template",
			template:  "",
			result:    "123",
			errorType: ErrTemplateMismatch,
			errorMsg:  "strings do not match",
		},
		{
			name:      "empty result",
			template:  "123",
			result:    "",
			errorType: ErrTemplateMismatch,
			errorMsg:  "strings do not match",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := New(tt.template)
			if err != nil {
				t.Errorf("New() error = %v", err)
				return
			}
			_, err = tmpl.findFieldIndicies(tt.result)
			if err == nil {
				t.Errorf("findFieldIndicies() error is nil")
				return
			}
			if !errors.Is(err, tt.errorType) {
				t.Errorf("findFieldIndicies() error type = '%v', want type '%v'", err, tt.errorType)
				return
			}
			if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("findFieldIndicies() error = '%v', want to contain '%v'", err, tt.errorMsg)
				return
			}
		})
	}
}
