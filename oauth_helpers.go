package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// netListen opens a TCP listener on the given address.
func netListen(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}

// randomState returns a cryptographically random hex string for OAuth CSRF.
func randomState() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "state"
	}
	return hex.EncodeToString(b)
}

// receiveCode starts a tiny HTTP server on ln, serves the OAuth redirect,
// validates the state param, and returns the authorization code.
func receiveCode(ln net.Listener, state string) (string, error) {
	done := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errCh <- errors.New("OAuth state mismatch (possible CSRF)")
			return
		}
		if e := q.Get("error"); e != "" {
			http.Error(w, e, http.StatusBadRequest)
			errCh <- fmt.Errorf("OAuth error: %s (%s)", e, q.Get("error_description"))
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			errCh <- errors.New("missing authorization code")
			return
		}
		fmt.Fprint(w, "✓ Authorized. You can close this tab and return to the terminal.")
		done <- code
	})

	srv := &http.Server{Handler: mux}
	go srv.Serve(ln) //nolint:errcheck

	select {
	case code := <-done:
		srv.Close() //nolint:errcheck
		return code, nil
	case err := <-errCh:
		srv.Close() //nolint:errcheck
		return "", err
	}
}

// splitLines splits a string on '\n', trimming a trailing empty element.
func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// joinLines joins lines with '\n'.
func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
