package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is set via -ldflags at build time (see README).
var version = "0.1.0"

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("dgtr %s\n", version)
		},
	}
}
