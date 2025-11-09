package main

import (
	"fmt"
	"os"

	"github.com/Snider/Mining/cmd/mining/cmd"
	_ "github.com/Snider/Mining/docs"
)

// @title Mining API
// @version 1.0
// @description This is a sample server for a mining application.
// @host localhost:8080
// @BasePath /api/v1/mining
func main() {
	// If no command is provided, default to "serve"
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "serve")
	}

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
