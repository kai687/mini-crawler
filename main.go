package main

import (
	"context"
	"fmt"
	"os"

	"github.com/algolia/mini-crawler/cmd"
)

func main() {
	if err := cmd.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
