package main

import (
	"os"

	"github.com/rishabyd/codeberg-cli/internal/commands"
	"github.com/rishabyd/codeberg-cli/internal/output"
)

var version = "dev"

func main() {
	exitCode, err := run()
	if err != nil {
		output.PrintErr(err.Error())
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
