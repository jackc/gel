package main

import (
	"bytes"
	"fmt"
)

func main() {
	var b bytes.Buffer
	EscapeHtml(&b)
	fmt.Print(b.String())
}
