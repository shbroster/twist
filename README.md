# 🌀 `twist`

> A reversible template package for Go — render structured text and parse it back into data.

## Overview

Twist is a lightweight Go library for creating *round-trippable* string templates. You can use it 
to:
- Render text from Go structs or maps
- Parse that text back into structured data
It's ideal for cases like logs, filenames, simple DSLs, and more.

## Why Twist?

Traditional templating libraries let you generate text, but don't help you turn that text back into 
structured data. Twist bridges that gap, making it easy to round-trip between text and Go data 
structures

## Examples

See the [go docs](https://pkg.go.dev/github.com/shbroster/twist#pkg-examples) for all examples.

A basic example is illustrated below:
```go
twist := MustNew("{{ Greeting }}, {{ Subject }}!")

data := map[string]string{
	"Greeting": "Hello",
	"Subject":  "World",
}
message := twist.MustExecute(data)
fmt.Printf("%#v\n", message)
// "Hello, World!"

fields, _ := twist.ParseToMap("Hello, World!")
fmt.Printf("%#v\n", fields)
// map[string]string{"Greeting":"Hello", "Subject":"World"}
```

## Features

- Simple template syntax using `{{field}}` delimiters (configuration exists for custom delimiters).
- Generate strings from templates using struct or map data.
- Parse strings back into struct or map data.
- Handle ambiguous matches, if a string matches the template in multiple ways, Twist returns all 
  possible structured interpretations.
- Type-safe parsing into Go structs.

## Installation

```sh
go get github.com/shbroster/twist
```

## License

The project is licensed under the [MIT License](./LICENSE).
