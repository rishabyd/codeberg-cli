package main

import (
	"fmt"
	"os"

	"github.com/rishabyd/codeberg-cli/internal/commands"
)

var version = "0.1.0"

func main() {
	exitCode, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCode)
	}
}

func run() (int, error) {
	err := commands.Execute(version, os.Args[1:])
	if err != nil {
		if commands.IsUsageError(err) {
			return 2, err
		}
		return 1, err
	}
	return 0, nil
}
