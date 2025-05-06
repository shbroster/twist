package twist

import "fmt"

func Example() {
	data := map[string]string{
		"Greeting": "Hello",
		"Subject":  "World",
	}
	twist := MustNew("{{ Greeting }}, {{ Subject }}!")
	message := twist.MustExecute(data)
	fmt.Printf("%#v\n", message)

	fields, _ := twist.ParseToMap(message)
	fmt.Printf("%#v\n", fields)

	// Output:
	// "Hello, World!"
	// map[string]string{"Greeting":"Hello", "Subject":"World"}
}

func Example_with_structs() {
	type Data struct {
		Greeting string
		Person   string
		Age      int
	}

	input := Data{
		Greeting: "Happy Birthday",
		Person:   "Annie",
		Age:      6,
	}
	twist := MustNew("{{ Greeting }}, {{ Person }}. You're {{ Age }} years old today!")
	message := twist.MustExecute(input)
	fmt.Printf("%#v\n", message)

	var output Data
	twist.Parse(message, &output)
	fmt.Printf("%#v\n", output)

	// Output:
	// "Happy Birthday, Annie. You're 6 years old today!"
	// twist.Data{Greeting:"Happy Birthday", Person:"Annie", Age:6}
}

func ExampleNew_error() {
	_, err := New("{{ Greeting }}, {{ Subject!")
	fmt.Println(err)
	// Output: unmatched delimiters: twist error: invalid template
}

func ExampleNew_custom_delimeters() {
	data := map[string]string{
		"Greeting": "Hello",
		"Subject":  "World",
	}
	twist := MustNew("[ Greeting ], [ Subject]!", WithDelimiters([2]string{"[", "]"}))
	message := twist.MustExecute(data)
	fmt.Printf("%#v\n", message)
	// Output: "Hello, World!"
}

func ExampleTwist_Execute_error_missing_field() {
	data := map[string]string{
		"Greeting": "Hello",
	}
	twist, _ := New("{{ Greeting }}, {{ Subject }}!")
	_, err := twist.Execute(data)
	fmt.Println(err)
	// Output: field 'Subject' is missing: twist error: invalid data
}

func ExampleTwist_Execute_error_not_unique() {
	data := map[string]string{
		"Greeting": "Good Night",
		"Subject":  "Mr. Tom",
	}
	twist := MustNew("{{ Greeting }} {{ Subject }}!")
	_, err := twist.Execute(data, WithUnique())
	fmt.Println(err)
	// Output: multiple mathces: twist error: template is ambiguous
}

func ExampleTwist_ParseToMap_error_ambiguous() {
	message := "Good Night Mr. Tom!"
	twist := MustNew("{{ Greeting }} {{ Subject }}!")
	_, err := twist.ParseToMap(message)
	fmt.Println(err)
	// Output: multiple matches: twist error: template is ambiguous
}

func ExampleTwist_ParseToMaps() {
	message := "Good Night Mr. Tom!"
	twist := MustNew("{{ Greeting }} {{ Subject }}!")
	fieldMaps, _ := twist.ParseToMaps(message)
	for _, fieldMap := range fieldMaps {
		fmt.Println(fieldMap)
	}
	// Output:
	// map[Greeting:Good Subject:Night Mr. Tom]
	// map[Greeting:Good Night Subject:Mr. Tom]
	// map[Greeting:Good Night Mr. Subject:Tom]
}
