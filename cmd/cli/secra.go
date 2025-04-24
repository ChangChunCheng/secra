package main

import (
	"fmt"
	"log"

	"gitlab.com/jacky850509/secra/cmd/cli/root"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func main() {
	fmt.Println("🚀 CLI Running")
	fmt.Printf("Version: %s\nCommit: %s\nBuildDate: %s\n", Version, Commit, BuildDate)

	if err := root.Execute(); err != nil {
		log.Fatalf("CLI error: %v", err)
	}
}
