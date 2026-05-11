package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kai687/mini-crawler/cmd"
)

func main() {
	if err := cmd.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
