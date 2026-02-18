package main

import (
	"emoji-roguelike/internal/game"
	"fmt"
	"os"
)

func main() {
	g, err := game.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	g.Run()
}
