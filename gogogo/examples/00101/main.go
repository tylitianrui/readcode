package main

import (
	_ "00101/pkg/internal"
	_ "unsafe"
)

//go:linkname printTime  00101/pkg/internal.printTime
func printTime()

func main() {
	printTime()
}
