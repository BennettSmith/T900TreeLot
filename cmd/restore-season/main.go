package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fmt.Fprintln(os.Stderr, "restore-season is not implemented yet.")
	fmt.Fprintln(os.Stderr, "This offline entry point will validate and restore encrypted season archives in a later increment.")
	fmt.Fprintln(os.Stderr, "It never applies schema migrations to a live application database.")
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "received arguments: %v\n", args)
	}
	return 2
}
