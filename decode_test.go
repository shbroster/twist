package twist

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeSucess(t *testing.T) {
	type testCase struct {
		name  string
		input string
		out   any
		want  any
	}

	tests := []testCase{
		{
			name:  "string",
			input: "123",
			out:   &struct{ Field string }{},
			want:  "123",
		},
		{
			name:  "int",
			input: "123",
			out:   &struct{ Field int }{},
			want:  123,
		},
		{
			name:  "int8",
			input: "123",
			out:   &struct{ Field int8 }{},
			want:  int8(123),
		},
		{
			name:  "int16",
			input: "123",
			out:   &struct{ Field int16 }{},
			want:  int16(123),
		},
		{
			name:  "int32",
			input: "123",
			out:   &struct{ Field int32 }{},
			want:  int32(123),
		},
		{
			name:  "int64",
			input: "123",
			out:   &struct{ Field int64 }{},
			want:  int64(123),
		},
		{
			name:  "uint",
			input: "123",
			out:   &struct{ Field uint }{},
			want:  uint(123),
		},
		{
			name:  "uint8",
			input: "123",
			out:   &struct{ Field uint8 }{},
			want:  uint8(123),
		},
		{
			name:  "uint16",
			input: "123",
			out:   &struct{ Field uint16 }{},
			want:  uint16(123),
		},
		{
			name:  "uint32",
			input: "123",
			out:   &struct{ Field uint32 }{},
			want:  uint32(123),
		},
		{
			name:  "uint64",
			input: "123",
			out:   &struct{ Field uint64 }{},
			want:  uint64(123),
		},
		{
			name:  "float",
			input: "123",
			out:   &struct{ Field float64 }{},
			want:  123.0,
		},
		{
			name:  "float64",
			input: "123",
			out:   &struct{ Field float64 }{},
			want:  float64(123.0),
		},
		{
			name:  "float32",
			input: "123",
			out:   &struct{ Field float32 }{},
			want:  float32(123.0),
		},
		{
			name:  "bool true",
			input: "true",
			out:   &struct{ Field bool }{},
			want:  true,
		},
		{
			name:  "bool false",
			input: "false",
			out:   &struct{ Field bool }{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]string{"Field": tt.input}
			err := decode(input, tt.out)
			if err != nil {
				t.Errorf("decode() error = %v", err)
				return
			}
			v := reflect.ValueOf(tt.out).Elem().FieldByName("Field").Interface()
			if diff := cmp.Diff(v, tt.want); diff != "" {
				t.Errorf("deocde() mismatch (-got +want)\n%s", diff)
				return
			}
		})
	}
}

func TestDecodeMultipleSuccess(t *testing.T) {

	var out struct {
		Field1 string
		Field2 int
		Field3 bool
		Field4 float64
		Field5 uint
	}
	want := &struct {
		Field1 string
		Field2 int
		Field3 bool
		Field4 float64
		Field5 uint
	}{
		Field1: "123",
		Field2: -345,
		Field3: true,
		Field4: 19.0,
		Field5: 98,
	}

	multipleInput := map[string]string{
		"Field1": "123",
		"Field2": "-345",
		"Field3": "true",
		"Field4": "19.0",
		"Field5": "98",
	}

	err := decode(multipleInput, &out)
	if err != nil {
		t.Errorf("decode() error = %v", err)
		return
	}
	if diff := cmp.Diff(out, *want); diff != "" {
		t.Errorf("deocde() mismatch (-got +want)\n%s", diff)
		return
	}
}

func TestDecodeInterfaceSuccess(t *testing.T) {
	type empty interface{}
	type nested interface{ empty }
	type testStruct struct {
		Field1 string
	}

	var emptyOut empty = &testStruct{
		Field1: "123",
	}
	var nestedOut nested = &testStruct{
		Field1: "123",
	}

	input := map[string]string{
		"Field1": "123",
	}
	want := &testStruct{
		Field1: "123",
	}

	tests := []struct {
		name string
		out  any
	}{
		{
			name: "empty",
			out:  emptyOut,
		},
		{
			name: "nested",
			out:  nestedOut,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decode(input, tt.out)
			if err != nil {
				t.Errorf("decode() error = %v", err)
				return
			}
			if diff := cmp.Diff(tt.out, want); diff != "" {
				t.Errorf("deocde() mismatch (-got +want)\n%s", diff)
				return
			}
		})
	}
}

func TestDecodeError(t *testing.T) {
	type testCase struct {
		name      string
		input     map[string]string
		out       any
		errorType error
		errorMsg  string
	}

	testString := "test"

	tests := []testCase{
		{
			name:      "invalid int",
			input:     map[string]string{"Field": "str"},
			out:       &struct{ Field int }{},
			errorType: ErrInvalidData,
			errorMsg:  "field 'Field' cannot be converted to supplied type",
		},
		{
			name:      "invalid float",
			input:     map[string]string{"Field": "str"},
			out:       &struct{ Field float64 }{},
			errorType: ErrInvalidData,
			errorMsg:  "field 'Field' cannot be converted to supplied type",
		},
		{
			name:      "invalid bool",
			input:     map[string]string{"Field": "str"},
			out:       &struct{ Field bool }{},
			errorType: ErrInvalidData,
			errorMsg:  "field 'Field' cannot be converted to supplied type",
		},
		{
			name:      "unsupported type",
			input:     map[string]string{"Field": "str"},
			out:       &struct{ Field complex128 }{},
			errorType: ErrInvalidData,
			errorMsg:  "field 'Field' is not a supported type",
		},
		{
			name:      "string not a pointer",
			input:     map[string]string{},
			out:       struct{ Field int }{},
			errorType: ErrInvalidData,
			errorMsg:  "out must be a non-nil pointer but got 'struct'",
		},
		{
			name:      "string not a pointer to a struct",
			input:     map[string]string{},
			out:       &testString,
			errorType: ErrInvalidData,
			errorMsg:  "out must point to a struct but got 'string'",
		},
		{
			name:      "missing field",
			input:     map[string]string{"Field1": "test", "Field2": "test"},
			out:       &struct{ Field1 string }{},
			errorType: ErrInvalidData,
			errorMsg:  "field 'Field2' is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decode(tt.input, tt.out)
			if err == nil {
				t.Errorf("decode() did not return an error")
				return
			}
			if !errors.Is(err, tt.errorType) {
				t.Errorf("decode() error = '%v', want '%v'", err, tt.errorType)
				return
			}
			if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("decode() error = '%v', want to contain '%v'", err, tt.errorMsg)
				return
			}
		})
	}
}
