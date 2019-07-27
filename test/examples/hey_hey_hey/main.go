package main

import (
	"bytes"
	"fmt"
)

func main() {
	var b bytes.Buffer
	HeyHeyHey(&b)
	fmt.Print(b.String())
}
