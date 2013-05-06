package main

import (
	"bytes"
	"fmt"
)

func main() {
	var b bytes.Buffer
	StringInterpolation(&b)
	fmt.Print(b.String())
}
