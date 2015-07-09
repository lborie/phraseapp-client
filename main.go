package main

import (
	"fmt"
	"github.com/dynport/dgtk/cli"
	"os"
)

func main() {
	Run()
}

func Run() {
	callArgs, _ := ConfigCallArgs()
	err := router(callArgs).RunWithArgs()
	switch err {
	case cli.ErrorHelpRequested, cli.ErrorNoRoute:
		os.Exit(1)
	case nil:
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
