package twist

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewSuccess(t *testing.T) {
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
			if diff := cmp.Diff(got.fields(), tt.expectedFields); diff != "" {
				t.Errorf("fields mismatch (-got +want)\n%s", diff)
			}
			if diff := cmp.Diff(got.pretext(), tt.expectedPretext); diff != "" {
				t.Errorf("pretext mismatch (-got +want)\n%s", diff)
			}
		})
	}
}

func TestNewError(t *testing.T) {
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

func TestParseToMapSuccess(t *testing.T) {
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
			out, err := tmpl.ParseToMap(tt.result)
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

func TestParseToMapError(t *testing.T) {
	type testCase struct {
		name      string
		template  string
		result    string
		errorType error
		errorMsg  string
	}

	tests := []testCase{
		{
			name:      "ambiguous multiple fields",
			template:  "...{{Name}}{{Age}}...",
			result:    "...12345678...",
			errorType: ErrAmbiguousTemplate,
			errorMsg:  "multiple matches",
		},
		{
			name:      "ambiguous combination",
			template:  "...{{Name}} {{Age}}...",
			result:    "...John Smith 23...",
			errorType: ErrAmbiguousTemplate,
			errorMsg:  "multiple matches",
		},
		{
			name:      "mismatch",
			template:  "a",
			result:    "b",
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
			_, err = tmpl.ParseToMap(tt.result)
			if err == nil {
				t.Errorf("ParseToMap() error is nil")
				return
			}
			if !errors.Is(err, tt.errorType) {
				t.Errorf("ParseToMap() error type = %v, want type %v", err, tt.errorType)
				return
			}
			if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("ParseToMap() error = %v, want to contain %v", err, tt.errorMsg)
				return
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

func TestParse(t *testing.T) {
	type testCase struct {
		name     string
		template string
		result   string
		out      any
		want     any
	}

	tests := []testCase{
		{
			name:     "basic",
			template: "Hello, {{Name}}",
			result:   "Hello, World",
			out: &struct {
				Name string
			}{},
			want: &struct {
				Name string
			}{
				Name: "World",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := New(tt.template)
			if err != nil {
				t.Errorf("New() error = %v", err)
			}
			err = tmpl.Parse(tt.result, tt.out)
			if err != nil {
				t.Errorf("New.Parse() error = %v", err)
			}
			if diff := cmp.Diff(tt.out, tt.want); diff != "" {
				t.Errorf("Parse() mismatch (-got +want)\n%s", diff)
			}
		})
	}
}
