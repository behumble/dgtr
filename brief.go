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

func resolveCredentials() (clientID, clientSecret, refreshToken string, err error) {
	cfg, loadErr := loadConfig(configFile)
	if loadErr == nil {
		clientID = cfg.ClientID
		clientSecret = cfg.ClientSecret
		refreshToken = cfg.RefreshToken
	}

	if clientID == "" {
		clientID = getEnv(envClientID, "")
	}
	if clientSecret == "" {
		clientSecret = getEnv(envClientSecret, "")
	}
	if refreshToken == "" {
		refreshToken = getEnv(envRefreshToken, "")
	}

	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return "", "", "", fmt.Errorf("missing GCP credentials or refresh token in ~/.dgtr/config.json (please run \"dgtr login\")")
	}

	return clientID, clientSecret, refreshToken, nil
}

func newReviewCmd() *cobra.Command {
	var dateStr string
	var all bool
	var modified bool
	cmd := &cobra.Command{
		Use:     "review",
		Aliases: []string{"brief"},
		Short:   "Print a review of task modifications (default: yesterday 00:00 until now)",
		Long: `review prints a concise, rule-based summary of modified Google Tasks.

By default it reports on tasks modified between yesterday 00:00:00 and now.
Pass --date YYYY-MM-DD to target a specific day, or --all to list every open task regardless of modification date.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientID, clientSecret, refresh, err := resolveCredentials()
			if err != nil {
				return err
			}

			srv, err := tasksClient(clientID, clientSecret, refresh)
			if err != nil {
				return err
			}

			now := time.Now()
			var start, end time.Time

			if dateStr != "" {
				t, err := time.ParseInLocation("2006-01-02", dateStr, now.Location())
				if err != nil {
					return fmt.Errorf("invalid --date %q (use YYYY-MM-DD): %w", dateStr, err)
				}
				start = t
				end = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, now.Location())
			} else {
				start = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
				end = now
			}

			b, err := buildReview(context.Background(), srv, start, end, dateStr, all)
			if err != nil {
				return err
			}
			fmt.Print(b)
			return nil
		},
	}
	cmd.Flags().StringVar(&dateStr, "date", "", "target specific date as YYYY-MM-DD")
	cmd.Flags().BoolVar(&all, "all", false, "list every open task regardless of modification date")
	cmd.Flags().BoolVarP(&modified, "modified", "m", false, "list modified tasks (default behavior)")
	return cmd
}

func newOpenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "open",
		Aliases: []string{"all", "tasks"},
		Short:   "Print all open (uncompleted) Google Tasks in Markdown format",
		Long:    `open prints a clean Markdown list of all currently open (not-yet-completed) Google Tasks across your task lists.`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientID, clientSecret, refresh, err := resolveCredentials()
			if err != nil {
				return err
			}

			srv, err := tasksClient(clientID, clientSecret, refresh)
			if err != nil {
				return err
			}

			b, err := buildOpenReview(context.Background(), srv, time.Now())
			if err != nil {
				return err
			}
			fmt.Print(b)
			return nil
		},
	}
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

// buildReview queries tasks and renders clean GitHub Flavored Markdown review text.
func buildReview(ctx context.Context, srv *tasks.Service, start, end time.Time, dateStr string, all bool) (string, error) {
	if all {
		return buildOpenReview(ctx, srv, time.Now())
	}

	type taskItem struct {
		task     *tasks.Task
		listName string
	}
	var open, completed []taskItem
	var allOpen []taskItem

	lists, err := srv.Tasklists.List().Do()
	if err != nil {
		return "", fmt.Errorf("list task lists: %w", err)
	}
	for _, tl := range lists.Items {
		req := srv.Tasks.List(tl.Id).MaxResults(100).ShowCompleted(true).ShowHidden(true)
		for {
			tasksResp, err := req.Do()
			if err != nil {
				return "", fmt.Errorf("list tasks for %q: %w", tl.Title, err)
			}
			for _, t := range tasksResp.Items {
				if t == nil || t.Deleted {
					continue
				}
				item := taskItem{task: t, listName: tl.Title}
				if all {
					if t.Status != "completed" {
						allOpen = append(allOpen, item)
					}
				} else {
					if isUpdatedBetween(t.Updated, start, end) {
						if t.Status == "completed" || t.Completed != nil {
							completed = append(completed, item)
						} else {
							open = append(open, item)
						}
					}
				}
			}
			if tasksResp.NextPageToken == "" {
				break
			}
			req.PageToken(tasksResp.NextPageToken)
		}
	}

	var sb strings.Builder
	sb.WriteString("# Daily Google Tasks Review\n\n")

	if all {
		fmt.Fprintf(&sb, "> **Scope:** All open tasks (as of %s)\n\n", time.Now().Format("2006-01-02 15:04"))
	} else if dateStr != "" {
		fmt.Fprintf(&sb, "> **Target Date:** %s\n\n", dateStr)
	} else {
		fmt.Fprintf(&sb, "> **Period:** %s 00:00 ~ %s\n\n", start.Format("2006-01-02"), end.Format("2006-01-02 15:04"))
	}

	renderSection := func(header string, items []taskItem) {
		fmt.Fprintf(&sb, "### %s (%d)\n\n", header, len(items))
		if len(items) == 0 {
			sb.WriteString("*None*\n\n")
			return
		}
		for _, it := range items {
			t := it.task
			name := strings.TrimSpace(t.Title)
			if name == "" {
				name = "(untitled)"
			}

			var metaParts []string
			if t.Due != "" {
				if d, err := time.Parse("2006-01-02T15:04:05.000Z", t.Due); err == nil {
					metaParts = append(metaParts, "due "+d.Format("2006-01-02"))
				} else if len(t.Due) >= 10 {
					metaParts = append(metaParts, "due "+t.Due[:10])
				}
			}

			if t.Completed != nil {
				if d, err := time.Parse(time.RFC3339, *t.Completed); err == nil {
					metaParts = append(metaParts, "completed "+d.Format("2006-01-02"))
				} else {
					metaParts = append(metaParts, "completed")
				}
			}

			if t.Updated != "" {
				if d, err := time.Parse(time.RFC3339Nano, t.Updated); err == nil {
					dLocal := d.Local()
					if dLocal.Format("2006-01-02") == time.Now().Format("2006-01-02") {
						metaParts = append(metaParts, "updated "+dLocal.Format("15:04"))
					} else {
						metaParts = append(metaParts, "updated "+dLocal.Format("2006-01-02 15:04"))
					}
				} else if d, err := time.Parse(time.RFC3339, t.Updated); err == nil {
					dLocal := d.Local()
					if dLocal.Format("2006-01-02") == time.Now().Format("2006-01-02") {
						metaParts = append(metaParts, "updated "+dLocal.Format("15:04"))
					} else {
						metaParts = append(metaParts, "updated "+dLocal.Format("2006-01-02 15:04"))
					}
				} else if len(t.Updated) >= 10 {
					metaParts = append(metaParts, "updated "+t.Updated[:10])
				}
			}

			metaStr := ""
			if len(metaParts) > 0 {
				metaStr = " *(" + strings.Join(metaParts, " | ") + ")*"
			}

			fmt.Fprintf(&sb, "- **%s**%s\n", name, metaStr)

			if t.Notes != "" {
				notesLines := strings.Split(strings.TrimSpace(t.Notes), "\n")
				for _, nl := range notesLines {
					fmt.Fprintf(&sb, "  > %s\n", nl)
				}
			}
		}
		sb.WriteString("\n")
	}

	if all {
		renderSection("Open Tasks", allOpen)
	} else {
		renderSection("Completed", completed)
		renderSection("Open", open)
		if len(completed) == 0 && len(open) == 0 {
			sb.WriteString("*No task updates recorded for this period.*\n")
		}
	}
	return sb.String(), nil
}

// buildOpenReview queries and renders all open tasks formatted in Markdown.
func buildOpenReview(ctx context.Context, srv *tasks.Service, now time.Time) (string, error) {
	type taskItem struct {
		task     *tasks.Task
		listName string
	}
	type listGroup struct {
		title string
		items []taskItem
	}
	var groups []listGroup
	totalOpen := 0

	lists, err := srv.Tasklists.List().Do()
	if err != nil {
		return "", fmt.Errorf("list task lists: %w", err)
	}

	for _, tl := range lists.Items {
		req := srv.Tasks.List(tl.Id).MaxResults(100).ShowCompleted(false).ShowHidden(false)
		var groupItems []taskItem
		for {
			tasksResp, err := req.Do()
			if err != nil {
				return "", fmt.Errorf("list tasks for %q: %w", tl.Title, err)
			}
			for _, t := range tasksResp.Items {
				if t == nil || t.Deleted || t.Status == "completed" {
					continue
				}
				groupItems = append(groupItems, taskItem{task: t, listName: tl.Title})
			}
			if tasksResp.NextPageToken == "" {
				break
			}
			req.PageToken(tasksResp.NextPageToken)
		}
		if len(groupItems) > 0 {
			groups = append(groups, listGroup{title: tl.Title, items: groupItems})
			totalOpen += len(groupItems)
		}
	}

	var sb strings.Builder
	sb.WriteString("# Open Google Tasks\n\n")
	fmt.Fprintf(&sb, "> **Scope:** All open tasks (as of %s)\n\n", now.Format("2006-01-02 15:04"))

	if totalOpen == 0 {
		sb.WriteString("*No open tasks found.*\n")
		return sb.String(), nil
	}

	multipleLists := len(groups) > 1
	for _, g := range groups {
		if multipleLists {
			fmt.Fprintf(&sb, "### %s (%d)\n\n", g.title, len(g.items))
		} else {
			fmt.Fprintf(&sb, "### Open Tasks (%d)\n\n", len(g.items))
		}
		for _, it := range g.items {
			t := it.task
			name := strings.TrimSpace(t.Title)
			if name == "" {
				name = "(untitled)"
			}

			var metaParts []string
			if t.Due != "" {
				if d, err := time.Parse("2006-01-02T15:04:05.000Z", t.Due); err == nil {
					metaParts = append(metaParts, "due "+d.Format("2006-01-02"))
				} else if len(t.Due) >= 10 {
					metaParts = append(metaParts, "due "+t.Due[:10])
				}
			}

			if t.Updated != "" {
				if d, err := time.Parse(time.RFC3339Nano, t.Updated); err == nil {
					dLocal := d.Local()
					if dLocal.Format("2006-01-02") == now.Format("2006-01-02") {
						metaParts = append(metaParts, "updated "+dLocal.Format("15:04"))
					} else {
						metaParts = append(metaParts, "updated "+dLocal.Format("2006-01-02 15:04"))
					}
				} else if d, err := time.Parse(time.RFC3339, t.Updated); err == nil {
					dLocal := d.Local()
					if dLocal.Format("2006-01-02") == now.Format("2006-01-02") {
						metaParts = append(metaParts, "updated "+dLocal.Format("15:04"))
					} else {
						metaParts = append(metaParts, "updated "+dLocal.Format("2006-01-02 15:04"))
					}
				} else if len(t.Updated) >= 10 {
					metaParts = append(metaParts, "updated "+t.Updated[:10])
				}
			}

			metaStr := ""
			if len(metaParts) > 0 {
				metaStr = " *(" + strings.Join(metaParts, " | ") + ")*"
			}

			fmt.Fprintf(&sb, "- **%s**%s\n", name, metaStr)

			if t.Notes != "" {
				notesLines := strings.Split(strings.TrimSpace(t.Notes), "\n")
				for _, nl := range notesLines {
					fmt.Fprintf(&sb, "  > %s\n", nl)
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// isUpdatedBetween reports whether an updated timestamp falls within [start, end].
func isUpdatedBetween(updatedStr string, start, end time.Time) bool {
	if updatedStr == "" {
		return false
	}
	var t time.Time
	var err error
	if t, err = time.Parse(time.RFC3339Nano, updatedStr); err != nil {
		if t, err = time.Parse(time.RFC3339, updatedStr); err != nil {
			if len(updatedStr) >= 10 {
				if d, err2 := time.Parse("2006-01-02", updatedStr[:10]); err2 == nil {
					t = d
				} else {
					return false
				}
			} else {
				return false
			}
		}
	}
	tLocal := t.In(start.Location())
	return !tLocal.Before(start) && !tLocal.After(end)
}


