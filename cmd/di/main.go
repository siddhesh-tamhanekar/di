package main

import (
	"flag"

	"github.com/siddhesh-tamhanekar/di/lib"
)

func main() {
	dir := flag.String("path", ".", "Path of the source code directory.")
	mod := flag.String("module", "", "name of the module optional")
	flag.Parse()
	// lib.Debug = true
	lib.Run(*dir, *mod)
}
