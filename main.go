package main

import (
	"github.com/justmiles/go-markdown2confluence/cmd"
)

// Version of markdown2confluence. Overwritten during build
var Version = "0.0.0"

func main() {
	cmd.Execute(Version)
}
