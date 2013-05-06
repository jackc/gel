package main

import (
	"bytes"
	"fmt"
)

func main() {
	var b bytes.Buffer
	Parameters(&b, "Jack", 3)
	fmt.Print(b.String())
}
