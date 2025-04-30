package twist

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestTemplateExceuteSucess(t *testing.T) {
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
			}
			got, err := tmpl.Execute(tt.data)
			if err != nil {
				t.Errorf("New.Execute() error: %v", err)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want)\n%s", diff)
			}
		})
	}
}

func TestTemplateExecuteError(t *testing.T) {
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
			_, err = tmpl.Execute(tt.data)
			if err == nil {
				t.Errorf("New.Execute() error is nil")
			}
			if !errors.Is(err, tt.errorType) {
				t.Errorf("New.Execute() error type = '%v', want type '%v'", err, tt.errorType)
			}
			if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("New.Execute() error = '%v', want to contain '%v'", err, tt.errorMsg)
			}
		})
	}
}

func TestNewRevTempl(t *testing.T) {
	type testCase struct {
		name            string
		template        string
		expectedFields  []string
		expectedPretext []string
	}

	tests := []testCase{
		{
			name:            "empty",
			template:        "",
			expectedFields:  []string{},
			expectedPretext: []string{""},
		},
		{
			name:            "no fields",
			template:        "Hello, world",
			expectedFields:  []string{},
			expectedPretext: []string{"Hello, world"},
		},
		{
			name:            "only template",
			template:        "{{Name}}",
			expectedFields:  []string{"Name"},
			expectedPretext: []string{"", ""},
		},
		{
			name:            "only template with whitespace",
			template:        "{{  Name  }}",
			expectedFields:  []string{"Name"},
			expectedPretext: []string{"", ""},
		},
		{
			name:            "basic 1",
			template:        "Hello, {{Name}}",
			expectedFields:  []string{"Name"},
			expectedPretext: []string{"Hello, ", ""},
		},
		{
			name:            "basic 2",
			template:        "{{HelloType}} Sam",
			expectedFields:  []string{"HelloType"},
			expectedPretext: []string{"", " Sam"},
		},
		{
			name:            "basic 3",
			template:        "Hello, {{Place}} Sam",
			expectedFields:  []string{"Place"},
			expectedPretext: []string{"Hello, ", " Sam"},
		},
		{
			name:            "whitespace",
			template:        " {{ Name }} ",
			expectedFields:  []string{"Name"},
			expectedPretext: []string{" ", " "},
		},
		{
			name:            "multiple fields 1",
			template:        "Hello, {{ Place }} Sam {{ Age }}. ",
			expectedFields:  []string{"Place", "Age"},
			expectedPretext: []string{"Hello, ", " Sam ", ". "},
		},
		{
			name:            "multiple fields 2",
			template:        "{{ Hello }}{{ Place }}{{ Age }}",
			expectedFields:  []string{"Hello", "Place", "Age"},
			expectedPretext: []string{"", "", "", ""},
		},
		{
			name:            "multiple field 3",
			template:        "Hello, {{ Place }} Sam {{ Age }}",
			expectedFields:  []string{"Place", "Age"},
			expectedPretext: []string{"Hello, ", " Sam ", ""},
		},
		{
			name:            "multiple fields 4",
			template:        "{{ Place }} Sam {{ Age }}. ",
			expectedFields:  []string{"Place", "Age"},
			expectedPretext: []string{"", " Sam ", ". "},
		},
		{
			name:            "contains underscores",
			template:        "{{ Hello_World }}. ",
			expectedFields:  []string{"Hello_World"},
			expectedPretext: []string{"", ". "},
		},
		{
			name:            "contains numbers",
			template:        "{{ Hello123 }}. ",
			expectedFields:  []string{"Hello123"},
			expectedPretext: []string{"", ". "},
		},
		{
			name:            "contains duplicates",
			template:        "{{ Hello }} {{ Hello }} - {{ Hello }}",
			expectedFields:  []string{"Hello", "Hello", "Hello"},
			expectedPretext: []string{"", " ", " - ", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.template)
			if err != nil {
				t.Errorf("New() error = %v", err)
			}
			if diff := cmp.Diff(got.original, tt.template); diff != "" {
				t.Errorf("original mismatch (-got +want)\n%s", diff)
			}
			if diff := cmp.Diff(got.fields, tt.expectedFields); diff != "" {
				t.Errorf("fields mismatch (-got +want)\n%s", diff)
			}
			if diff := cmp.Diff(got.pretext, tt.expectedPretext); diff != "" {
				t.Errorf("pretext mismatch (-got +want)\n%s", diff)
			}
		})
	}
}

func TestNewRevTemplError(t *testing.T) {
	type testCase struct {
		name      string
		template  string
		errorType error
		errorMsg  string
	}

	tests := []testCase{
		{
			name:      "field must not start with lowercase letter",
			template:  "{{invalidField}}",
			errorType: ErrInvalidField,
			errorMsg:  "must start with an uppercase letter",
		},
		{
			name:      "field must not start with a number",
			template:  "{{1InvalidField}}",
			errorType: ErrInvalidField,
			errorMsg:  "must start with an uppercase letter",
		},
		{
			name:      "field must not start with an underscore",
			template:  "{{_InvalidField}}",
			errorType: ErrInvalidField,
			errorMsg:  "must start with an uppercase letter",
		},
		{
			name:      "field must not contain special characters",
			template:  "{{InvalidField@}}",
			errorType: ErrInvalidField,
			errorMsg:  "must contain only letters, digits, and underscores",
		},
		{
			name:      "field must not be empty 1",
			template:  "{{}}",
			errorType: ErrInvalidField,
			errorMsg:  "must not be empty",
		},
		{
			name:      "field must not be empty 2",
			template:  "{{    }}",
			errorType: ErrInvalidField,
			errorMsg:  "must not be empty",
		},
		{
			name:      "missing closing brace 1",
			template:  "{{ Hello",
			errorType: ErrInvalidTemplate,
			errorMsg:  "unmatched delimiters",
		},
		{
			name:      "missing closing brace 2",
			template:  "{{ Hello }",
			errorType: ErrInvalidTemplate,
			errorMsg:  "unmatched delimiters",
		},
		{
			name:      "missing open brace 1",
			template:  "Hello }}",
			errorType: ErrInvalidTemplate,
			errorMsg:  "unmatched delimiters",
		},
		{
			name:      "missing open brace 2",
			template:  "{ Hello }}",
			errorType: ErrInvalidTemplate,
			errorMsg:  "unmatched delimiters",
		},
		{
			name:      "missing closing braces",
			template:  "{{ {{ Hello }}",
			errorType: ErrInvalidTemplate,
			errorMsg:  "nested delimiters",
		},
		{
			name:      "missing open braces",
			template:  "{{ Hello }} }}",
			errorType: ErrInvalidTemplate,
			errorMsg:  "unmatched delimiters",
		},
		{
			name:      "nested braces",
			template:  "{{ {{ Hello }} }}",
			errorType: ErrInvalidTemplate,
			errorMsg:  "nested delimiters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.template)
			if err == nil {
				t.Errorf("New() error is nil")
			}
			if !errors.Is(err, tt.errorType) {
				t.Errorf("New() error type = '%v', want type '%v'", err, tt.errorType)
			}
			if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("New() error = '%v', want to contain '%v'", err, tt.errorMsg)
			}
		})
	}
}

func TestFindFieldIndicies(t *testing.T) {
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
			name:     "mutliple matches",
			template: "{{First}}-{{Seconds}}",
			result:   "----",
			want: [][][2]int{
				{[2]int{0, 0}, [2]int{1, 4}},
				{[2]int{0, 1}, [2]int{2, 4}},
				{[2]int{0, 2}, [2]int{3, 4}},
				{[2]int{0, 3}, [2]int{4, 4}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := New(tt.template)
			if err != nil {
				t.Errorf("New() error = %v", err)
			}
			results, err := tmpl.findFieldIndicies(tt.result)
			if err != nil {
				t.Errorf("template mismatch: %v", err)
				return
			}
			if diff := cmp.Diff(results, tt.want); diff != "" {
				t.Errorf("Parse() mismatch (-got +want)\n%s", diff)
			}
		})
	}
}

func TestParseFields(t *testing.T) {
	type testCase struct {
		name     string
		template string
		result   string
		want     map[string]string
	}

	tests := []testCase{
		{
			name:     "basic",
			template: "Hello, {{Name}}",
			result:   "Hello, World",
			want: map[string]string{
				"Name": "World",
			},
		},
		{
			name:     "multiple fields",
			template: "Hello, {{Name}} {{Age}}",
			result:   "Hello, World 25",
			want: map[string]string{
				"Name": "World",
				"Age":  "25",
			},
		},
		// {
		// 	name:     "ambiguous multiple fields",
		// 	template: "...{{Name}}{{Age}}...",
		// 	result:   "...12345678...",
		// 	want: map[string]string{
		// 		"Name": "",
		// 		"Age":  "12345678",
		// 	},
		// },
		// {
		// 	name:     "ambiguous combination",
		// 	template: "...{{Name}} {{Age}}...",
		// 	result:   "...John Smith 23...",
		// 	want: map[string]string{
		// 		"Name": "John",
		// 		"Age":  "Smith 23",
		// 	},
		// },
		{
			name:     "empty",
			template: "",
			result:   "",
			want:     map[string]string{},
		},
		{
			name:     "no fields",
			template: "Hello, world",
			result:   "Hello, world",
			want:     map[string]string{},
		},
		{
			name:     "duplicates",
			template: "{{Hello}} {{Hello}}",
			result:   "Hi Hi",
			want: map[string]string{
				"Hello": "Hi",
			},
		},
		{
			name:     "complex",
			template: " {{A}} {{B}} {{A}}  {{C}} {{B}} {{G}} ",
			result:   " Apple Banana Apple  Cucumber Banana Grape ",
			want: map[string]string{
				"A": "Apple",
				"B": "Banana",
				"C": "Cucumber",
				"G": "Grape",
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
			out, err := tmpl.ParseFields(tt.result)
			if err != nil {
				t.Errorf("New.Parse() error = %v", err)
				return
			}
			if diff := cmp.Diff(out, tt.want); diff != "" {
				t.Errorf("Parse() mismatch (-got +want)\n%s", diff)
			}
		})
	}
}

// func FuzzParseField(f *testing.F) {
// 	// Add seed values to help the fuzzer
// 	f.Add("Hello {{Name}}", "Name:John")
// 	f.Add("GR#{{Greeting}}#NR#{{Name}}", "Greeting:Hi,Name:John")

// 	f.Fuzz(func(t *testing.T, template string, dataStr string) {
// 		if len(dataStr) < 3 || !strings.Contains(template, "{{") || !strings.Contains(template, "}}") {
// 			t.Skip()
// 		}

// 		// Parse the data string into a map
// 		data := make(map[string]string)
// 		pairs := strings.Split(dataStr, ",")
// 		for _, pair := range pairs {
// 			parts := strings.SplitN(pair, ":", 2)
// 			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
// 				t.Skip() // Skip invalid data format
// 			}
// 			data[parts[0]] = parts[1]
// 		}

// 		tmpl, err := New(template)
// 		if err != nil {
// 			t.Skip()
// 		}
// 		result, err := tmpl.Execute(data)
// 		if err != nil {
// 			t.Skip()
// 		}
// 		out, err := tmpl.ParseFields(result)
// 		if err != nil {
// 			t.Skip()
// 		}
// 		if diff := cmp.Diff(out, data); diff != "" {
// 			t.Errorf("Mismatch\n template: '%s'\n map data: '%s'\n result  : '%s'\n (-got +want)\n%s", template, data, result, diff)
// 		}
// 	})
// }

// func TestParse(t *testing.T) {
// 	type testCase struct {
// 		name     string
// 		template string
// 		result   string
// 		out      any
// 		want     any
// 	}

// 	tests := []testCase{
// 		{
// 			name:     "basic",
// 			template: "Hello, {{Name}}",
// 			result:   "Hello, World",
// 			out: &struct {
// 				Name string
// 			}{},
// 			want: &struct {
// 				Name string
// 			}{
// 				Name: "World",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tmpl, err := New(tt.template)
// 			if err != nil {
// 				t.Errorf("New() error = %v", err)
// 			}
// 			err = tmpl.Parse(tt.result, tt.out)
// 			if err != nil {
// 				t.Errorf("New.Parse() error = %v", err)
// 			}
// 			if diff := cmp.Diff(tt.out, tt.want); diff != "" {
// 				t.Errorf("Parse() mismatch (-got +want)\n%s", diff)
// 			}
// 		})
// 	}
// }
