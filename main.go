// Command dgtr (Daily Google Tasks Review) is a lightweight CLI that prints
// a rule-based review of Google Tasks modifications since yesterday 00:00.
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
	// envFile is the path to the .env file (overridable via DGTR_ENV or DGTB_ENV).
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
	defaultEnv := getEnv("DGTR_ENV", getEnv("DGTB_ENV", ".env"))
	root := &cobra.Command{
		Use:   "dgtr",
		Short: "Daily Google Tasks Review - rule-based Google Tasks review CLI",
		Long: `dgtr prints a concise review of modifications to your Google Tasks.

It talks directly to the Google Tasks API - no LLM, no per-request cost,
just a single static binary. Configure your credentials in .env (see
.gitignore and README), run "dgtr login" once to authorize, then
"dgtr review" (or "dgtr brief") to see task changes since yesterday 00:00.

Run "dgtr review --date 2026-08-03" to target a specific day.
`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			envFile = getEnv("DGTR_ENV", getEnv("DGTB_ENV", ".env"))
			if err := loadDotEnv(envFile); err != nil {
				return err
			}
			return nil
		},
	}

	root.PersistentFlags().StringVarP(&envFile, "env", "e", defaultEnv, "path to .env file")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")

	root.AddCommand(newLoginCmd())
	root.AddCommand(newReviewCmd())
	root.AddCommand(newOpenCmd())
	root.AddCommand(newVersionCmd())
	return root
}

