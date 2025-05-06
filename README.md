# 🌀 `twist`

> A reversible template package for Go — render structured text and parse it back into data.

## Overview

Twist is a Go library for creating reversible string templates. It allows you to both generate strings from templates and parse strings back into structured data.

## Features

- Simple template syntax using `{{field}}` delimiters (with configuration options for custom delimeters)
- Generate strings from templates using struct or map data
- Parse strings back into struct or map data
- Handle ambiguous matches by returning all possible interpretations
- Type-safe parsing into Go structs

## Installation

```sh
go get github.com/shbroster/twist
```

## Examples

See the [go docs](https://pkg.go.dev/github.com/shbroster/twist) for examples.  

## License

The project is licensed under the [MIT License](./LICENSE).