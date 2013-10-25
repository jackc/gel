package main

import (
	"bytes"
	"fmt"
)

func main() {
	var b bytes.Buffer
	HelloWorld(&b)
	EscapeHtml(&b)
	fmt.Print(b.String())
}
