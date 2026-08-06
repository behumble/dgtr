// Command dgtb (Daily Google Tasks Brief) is a lightweight CLI that prints
// a rule-based briefing of yesterday's completed and in-progress Google Tasks.
//
// It calls the Google Tasks API directly (no LLM, no token cost) and keeps
// everything in a single binary.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// envFile is the path to the .env file (overridable via DGTB_ENV).
	envFile string
	// verbose enables extra logging.
	verbose bool
)

func main() {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "dgtb",
		Short: "Daily Google Tasks Brief - rule-based Google Tasks briefing CLI",
		Long: `dgtb prints a concise briefing of your Google Tasks.

It talks directly to the Google Tasks API - no LLM, no per-request cost,
just a single static binary. Configure your credentials in .env (see
.gitignore and README), run "dgtb login" once to authorize, then
"dgtb brief" whenever you want a briefing.

Run "dgtb brief --date 2026-08-03" to target a specific day.
`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			envFile = getEnv("DGTB_ENV", ".env")
			if err := loadDotEnv(envFile); err != nil {
				return err
			}
			return nil
		},
	}

	root.PersistentFlags().StringVarP(&envFile, "env", "e", getEnv("DGTB_ENV", ".env"), "path to .env file")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")

	root.AddCommand(newLoginCmd())
	root.AddCommand(newBriefCmd())
	root.AddCommand(newVersionCmd())
	return root
}
