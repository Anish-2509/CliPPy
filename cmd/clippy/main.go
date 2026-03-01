package main

import (
	"clippy/internal/cli"
	"os"
)

func main() {
	// CLI options; DB initialization happens lazily in commands that need persistence.
	opts := &cli.CLIOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	// Execute the CLI
	cli.Execute(opts)
}
