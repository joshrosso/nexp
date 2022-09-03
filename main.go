package main

import (
	"fmt"
	"os"

	"github.com/joshrosso/nexp/cmd"
)

const (
	tokenEnvVarName = "NOTION_TOKEN"
)

func main() {
	c := cmd.SetupCommands()
	if err := c.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
