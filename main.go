package main

import (
	"os"

	"github.com/songzhibin97/mini-compiler/repl"
)

func main() {
	repl.Start(os.Stdin, os.Stdout)
}
