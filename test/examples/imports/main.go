package main

import (
	"bytes"
	"fmt"
)

func main() {
	var b bytes.Buffer
	Imports(&b)
	fmt.Print(b.String())
}
