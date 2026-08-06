package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// loadDotEnv loads KEY=VALUE entries from the given .env file if it exists.
// Missing files are not an error (credentials can come from the environment);
// malformed files are reported.
func loadDotEnv(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if verbose {
			fmt.Fprintf(os.Stderr, "no %s found, using environment\n", path)
		}
		return nil
	}
	if err := godotenv.Load(path); err != nil {
		return fmt.Errorf("load %s: %w", path, err)
	}
	return nil
}
