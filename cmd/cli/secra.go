package main

import (
	"log"

	"gitlab.com/jacky850509/secra/cmd/cli/root"
)

func main() {
	if err := root.Execute(); err != nil {
		log.Fatalf("CLI error: %v", err)
	}
}
