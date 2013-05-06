package main

import (
	"bytes"
	"fmt"
)

func main() {
	var b bytes.Buffer
	IntegerInterpolation(&b)
	fmt.Print(b.String())
}
