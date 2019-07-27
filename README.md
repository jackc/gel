# Gel - Embedded Go templates

Gel is a templating library that compiles templates into Go functions.

## Installation

```
go get -u github.com/jackc/gel
```

## Usage

```
gel < users_index.html | goimports > users_index.go
```

## Example

```
package main

func main() {
  t(os.Stdout)
}

func t(w io.Writer) error
---
Hello, <%= "Jack" %>!
```

All text above the `---` is emitted directly into the output file. The last line must be a function signature that has a `io.Writer`
named `w` and returns an `error`. The template below the `---` will be converted into the function body.
