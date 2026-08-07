package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

const (
	envClientID     = "GOOGLE_TASKS_CLIENT_ID"
	envClientSecret = "GOOGLE_TASKS_CLIENT_SECRET"
	envRefreshToken = "GOOGLE_TASKS_REFRESH_TOKEN"
	tasksScope      = "https://www.googleapis.com/auth/tasks"
)

func newLoginCmd() *cobra.Command {
	var force bool
	var flagClientID string
	var flagClientSecret string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authorize Google Tasks (OAuth 2.0 PKCE) and store credentials",
		Long: `login runs the Google OAuth 2.0 PKCE authorization flow for Google Tasks
and saves the client_id, client_secret, and refresh token into ~/.dgtr/config.json.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(configFile)
			if err != nil {
				return err
			}

			creds, err := loadCredentials("")
			if err != nil {
				return err
			}

			// Check environment variable fallback as well
			envToken := getEnv(envRefreshToken, "")
			if (creds.RefreshToken != "" || envToken != "") && !force {
				fmt.Println("Refresh token already set. Use --force to re-authorize.")
				return nil
			}

			// Priority: Flag > Env > Config
			clientID := flagClientID
			if clientID == "" {
				clientID = getEnv(envClientID, cfg.ClientID)
			}

			clientSecret := flagClientSecret
			if clientSecret == "" {
				clientSecret = getEnv(envClientSecret, cfg.ClientSecret)
			}

			// Prompt if credentials are missing
			if clientID == "" {
				val, err := promptInput("Enter your GCP OAuth Client ID: ")
				if err != nil {
					return fmt.Errorf("read Client ID: %w", err)
				}
				clientID = val
			}

			if clientSecret == "" {
				val, err := promptInput("Enter your GCP OAuth Client Secret: ")
				if err != nil {
					return fmt.Errorf("read Client Secret: %w", err)
				}
				clientSecret = val
			}

			if clientID == "" || clientSecret == "" {
				return errors.New("GCP OAuth Client ID and Client Secret are required")
			}

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
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://accounts.google.com/o/oauth2/auth",
					TokenURL: "https://oauth2.googleapis.com/token",
				},
				RedirectURL: redirectURL,
			}

			verifier := oauth2.GenerateVerifier()
			state := randomState()
			url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce, oauth2.S256ChallengeOption(verifier))

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

			cfg.ClientID = clientID
			cfg.ClientSecret = clientSecret
			creds.RefreshToken = tk.RefreshToken

			cfgPath := configFile
			if cfgPath == "" {
				cfgPath, _ = defaultConfigPath()
			}
			credsPath, _ := defaultCredentialsPath()

			if err := saveConfig(configFile, cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}
			if err := saveCredentials("", creds); err != nil {
				return fmt.Errorf("save credentials: %w", err)
			}

			fmt.Println("✓ Authorized successfully.")
			fmt.Println("  Config saved to:", cfgPath)
			fmt.Println("  Token saved to: ", credsPath)
			fmt.Println("Now run: dgtr review")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "re-run the OAuth flow even if a token exists")
	cmd.Flags().StringVar(&flagClientID, "client-id", "", "GCP OAuth Client ID")
	cmd.Flags().StringVar(&flagClientSecret, "client-secret", "", "GCP OAuth Client Secret")
	return cmd
}

func promptInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
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
