package main

import (
	"fmt"
	"os"

	"github.com/runetale/runetale-oidc-server/cmd/db/cmd"
)

func main() {
	if err := cmd.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
