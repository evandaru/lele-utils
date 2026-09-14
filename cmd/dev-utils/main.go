package main

import (
	"fmt"
	"os"

	"dev-utils/internal/cli"
)

var version = "v0.1.0"

func main() {
	if err := cli.NewRoot(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
