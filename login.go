package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

const (
	envClientID     = "GOOGLE_TASKS_CLIENT_ID"
	envRefreshToken = "GOOGLE_TASKS_REFRESH_TOKEN"
	tasksScope      = "https://www.googleapis.com/auth/tasks"
)

func getDefaultClientID() string {
	p1 := "952660384697-nbav4brvoolkri7sd2gcdvrpfahqd8fe"
	p2 := ".apps.googleusercontent.com"
	return p1 + p2
}

func newLoginCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authorize Google Tasks (OAuth 2.0 PKCE) and store refresh token",
		Long: `login runs the Google OAuth 2.0 PKCE authorization flow for Google Tasks
and saves the resulting refresh token into ~/.dgtr/config.json.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(configFile)
			if err != nil {
				return err
			}

			// Check environment variable fallback as well
			envToken := getEnv(envRefreshToken, "")
			if (cfg.RefreshToken != "" || envToken != "") && !force {
				fmt.Println("Refresh token already set. Use --force to re-authorize.")
				return nil
			}

			clientID := getEnv(envClientID, getDefaultClientID())

			ln, err := netListen("127.0.0.1:0")
			if err != nil {
				return err
			}
			defer ln.Close()
			redirectURL := fmt.Sprintf("http://%s/", ln.Addr().String())

			conf := &oauth2.Config{
				ClientID:     clientID,
				ClientSecret: "", // PKCE mode - zero client secret
				Scopes:       []string{tasksScope},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://accounts.google.com/o/oauth2/auth",
					TokenURL: "https://oauth2.googleapis.com/token",
				},
				RedirectURL: redirectURL,
			}

			verifier := oauth2.GenerateVerifier()
			state := randomState()
			url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce, oauth2.VerifierOption(verifier))

			fmt.Println("\nOpen this URL in your browser and sign in with your Google account:")
			fmt.Println(url)
			if err := openBrowser(url); err != nil && verbose {
				fmt.Fprintf(os.Stderr, "could not auto-open browser: %v\n", err)
			}

			code, err := receiveCode(ln, state)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			tk, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
			if err != nil {
				return fmt.Errorf("token exchange: %w", err)
			}
			if tk.RefreshToken == "" {
				return errors.New("no refresh token returned from Google OAuth")
			}

			cfg.RefreshToken = tk.RefreshToken
			targetPath := configFile
			if targetPath == "" {
				targetPath, _ = defaultConfigPath()
			}
			if err := saveConfig(configFile, cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Println("✓ Authorized successfully. Config saved to", targetPath)
			fmt.Println("Now run: dgtr review")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "re-run the OAuth flow even if a token exists")
	return cmd
}

// openBrowser attempts to open url in the default browser.
func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
