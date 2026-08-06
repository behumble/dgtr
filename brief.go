package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"
)

func newBriefCmd() *cobra.Command {
	var dateStr string
	var all bool
	cmd := &cobra.Command{
		Use:   "brief",
		Short: "Print a briefing of yesterday's (or a given day's) tasks",
		Long: `brief prints a concise, rule-based summary of Google Tasks.

By default it reports on yesterday's completed tasks plus any currently open
(not-yet-completed) tasks with a due date on or before today. Pass --date
YYYY-MM-DD to target a specific day, or --all to list every open task
regardless of due date.

Requires GOOGLE_TASKS_CLIENT_ID, GOOGLE_TASKS_CLIENT_SECRET and
GOOGLE_TASKS_REFRESH_TOKEN in .env (run "dgtb login" first).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			refresh := getEnv(envRefreshToken, "")
			clientID := getEnv(envClientID, "")
			clientSecret := getEnv(envClientSecret, "")
			if refresh == "" || clientID == "" || clientSecret == "" {
				return fmt.Errorf("missing credentials: ensure %s, %s, %s are set in .env (run dgtb login)", envRefreshToken, envClientID, envClientSecret)
			}

			// Determine the target day.
			target := time.Now()
			if dateStr != "" {
				t, err := time.Parse("2006-01-02", dateStr)
				if err != nil {
					return fmt.Errorf("invalid --date %q (use YYYY-MM-DD): %w", dateStr, err)
				}
				target = t
			} else if !all {
				target = target.AddDate(0, 0, -1) // yesterday
			}

			srv, err := tasksClient(clientID, clientSecret, refresh)
			if err != nil {
				return err
			}

			b, err := buildBrief(context.Background(), srv, target, all)
			if err != nil {
				return err
			}
			fmt.Print(b)
			return nil
		},
	}
	cmd.Flags().StringVar(&dateStr, "date", "", "target date as YYYY-MM-DD (default: yesterday)")
	cmd.Flags().BoolVar(&all, "all", false, "list every open task regardless of due date")
	return cmd
}

// tasksClient builds a *tasks.Service authenticated with the stored refresh token.
func tasksClient(clientID, clientSecret, refresh string) (*tasks.Service, error) {
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     oauth2.Endpoint{TokenURL: "https://oauth2.googleapis.com/token"},
		Scopes:       []string{tasksScope},
	}
	tok := &oauth2.Token{RefreshToken: refresh}
	src := conf.TokenSource(ctx, tok)
	client := &http.Client{Transport: &authTransport{src: src}}
	return tasks.NewService(ctx, option.WithHTTPClient(client))
}

// authTransport injects the OAuth access token into each request.
type authTransport struct {
	src oauth2.TokenSource
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	tok, err := t.src.Token()
	if err != nil {
		return nil, err
	}
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	return http.DefaultTransport.RoundTrip(req)
}

// buildBrief queries tasks and renders the briefing text.
func buildBrief(ctx context.Context, srv *tasks.Service, target time.Time, all bool) (string, error) {
	day := target.Format("2006-01-02")

	var open, completed []*tasks.Task
	lists, err := srv.Tasklists.List().Do()
	if err != nil {
		return "", fmt.Errorf("list task lists: %w", err)
	}
	for _, tl := range lists.Items {
		tasksResp, err := srv.Tasks.List(tl.Id).MaxResults(100).ShowCompleted(true).Do()
		if err != nil {
			return "", fmt.Errorf("list tasks for %q: %w", tl.Title, err)
		}
		for _, t := range tasksResp.Items {
			if t == nil || t.Deleted {
				continue
			}
			if t.Completed != nil && dueOn(t.Completed, day) {
				completed = append(completed, t)
			} else if t.Status != "completed" {
				if all || dueOnOrBefore(t.Due, day) {
					open = append(open, t)
				}
			}
		}
	}

	var sb strings.Builder
	title := "Today's briefing"
	if !all {
		title = fmt.Sprintf("Task briefing for %s", day)
		if !isTodayOrFuture(target) {
			title = fmt.Sprintf("Task briefing for %s (yesterday)", day)
		}
	}
	fmt.Fprintf(&sb, "## %s\n\n", title)

	listList := func(header string, items []*tasks.Task) {
		if len(items) == 0 {
			fmt.Fprintf(&sb, "**%s:** none\n\n", header)
			return
		}
		fmt.Fprintf(&sb, "**%s:**\n", header)
		for _, t := range items {
			name := strings.TrimSpace(t.Title)
			if name == "" {
				name = "(untitled)"
			}
			due := ""
			if t.Due != "" {
				if d, err := time.Parse("2006-01-02T15:04:05.000Z", t.Due); err == nil {
					due = " (due " + d.Format("2006-01-02") + ")"
				} else if len(t.Due) >= 10 {
					due = " (due " + t.Due[:10] + ")"
				}
			}
			if t.Completed != nil {
				done := ""
				if d, err := time.Parse(time.RFC3339, *t.Completed); err == nil {
					done = " ✓ completed " + d.Format("2006-01-02")
				} else {
					done = " ✓"
				}
				fmt.Fprintf(&sb, "- [x] %s%s%s\n", name, due, done)
			} else {
				fmt.Fprintf(&sb, "- [ ] %s%s\n", name, due)
			}
		}
		sb.WriteString("\n")
	}

	if all {
		listList(fmt.Sprintf("Open tasks (as of %s)", time.Now().Format("2006-01-02")), open)
	} else {
		listList("Completed", completed)
		listList("Open (due on/before today)", open)
	}

	if len(completed) == 0 && len(open) == 0 {
		fmt.Fprintln(&sb, "_No tasks recorded for this period._")
	}
	return sb.String(), nil
}

// dueOn reports whether a timestamp falls on the given day.
func dueOn(ts *string, day string) bool {
	if ts == nil {
		return false
	}
	if d, err := time.Parse(time.RFC3339, *ts); err == nil {
		return d.Format("2006-01-02") == day
	}
	if len(*ts) >= 10 {
		return (*ts)[:10] == day
	}
	return false
}

// dueOnOrBefore reports whether a due timestamp falls on or before the given day.
func dueOnOrBefore(due, day string) bool {
	if due == "" {
		return false
	}
	target, _ := time.Parse("2006-01-02", day)
	if len(due) >= 10 {
		d, err := time.Parse("2006-01-02", due[:10])
		if err == nil {
			return !d.After(target)
		}
	}
	return false
}

func isTodayOrFuture(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.YearDay() >= now.YearDay()
}
