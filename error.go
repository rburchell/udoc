package main

import (
	"fmt"
)

func docError(f File, line int, text estring) {
	if f == nil {
		return
	}

	fmt.Printf("%s:%d: %s\n", f.Name(), line, text)
}
