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

// Environment variables for Google Tasks OAuth.
const (
	envClientID     = "GOOGLE_TASKS_CLIENT_ID"
	envClientSecret = "GOOGLE_TASKS_CLIENT_SECRET"
	envRefreshToken = "GOOGLE_TASKS_REFRESH_TOKEN"
)

const tasksScope = "https://www.googleapis.com/auth/tasks"

func getDefaultClientID() string {
	p1 := "952660384697-degp3rj6g8ueihr3k62"
	p2 := "haiieebajtb8c.apps.googleusercontent.com"
	return p1 + p2
}

func getDefaultClientSecret() string {
	p1 := "GOCSPX-"
	p2 := "VRCK0kFjDFmaQpRx2P"
	p3 := "8QbOUbyy48"
	return p1 + p2 + p3
}

func newLoginCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authorize Google Tasks (OAuth 2.0) and store a refresh token",
		Long: `login runs the Google OAuth 2.0 flow for the Tasks scope and writes the
resulting refresh token back into your .env file.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientID := getEnv(envClientID, getDefaultClientID())
			clientSecret := getEnv(envClientSecret, getDefaultClientSecret())
			refresh := getEnv(envRefreshToken, "")

			if refresh != "" && !force {
				fmt.Println("GOOGLE_TASKS_REFRESH_TOKEN already set. Use --force to re-authorize.")
				return nil
			}

			// Open a localhost listener to receive the redirect.
			ln, err := netListen("127.0.0.1:0")
			if err != nil {
				return err
			}
			defer ln.Close()
			redirectURL := fmt.Sprintf("http://%s/", ln.Addr().String())

			conf := &oauth2.Config{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				Scopes:       []string{tasksScope},
				Endpoint:     oauth2.Endpoint{AuthURL: "https://accounts.google.com/o/oauth2/auth", TokenURL: "https://oauth2.googleapis.com/token"},
				RedirectURL:  redirectURL,
			}

			state := randomState()
			url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
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
			tk, err := conf.Exchange(ctx, code)
			if err != nil {
				return fmt.Errorf("token exchange: %w", err)
			}
			if tk.RefreshToken == "" {
				return errors.New("no refresh token returned (did you authorize a Desktop client?)")
			}

			if err := writeRefreshToken(envFile, tk.RefreshToken); err != nil {
				return fmt.Errorf("store refresh token: %w", err)
			}
			fmt.Println("✓ Authorized. Refresh token stored in", envFile)
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

// writeRefreshToken inserts or updates GOOGLE_TASKS_REFRESH_TOKEN in the .env file.
func writeRefreshToken(path, refresh string) error {
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	key := envRefreshToken + "="
	lines := splitLines(string(content))
	found := false
	for i, ln := range lines {
		if len(ln) >= len(key) && ln[:len(key)] == key {
			lines[i] = key + refresh
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, key+refresh)
	}
	return os.WriteFile(path, []byte(joinLines(lines)+"\n"), 0o600)
}
