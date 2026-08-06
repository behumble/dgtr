# dgtr — Daily Google Tasks Review

[![한국어](https://img.shields.io/badge/Language-%ED%95%9C%EA%B5%AD%EC%96%B4-blue.svg)](README.ko.md)

A tiny, single-binary CLI that prints a concise review of your Google Tasks modifications since yesterday 00:00.
No LLM, no per-request token cost — it talks **directly to the Google Tasks
API** and renders a rule-based summary. OAuth tokens stay safely stored in
`~/.dgtr/config.json`, so anyone can use it with their own Google account.

```markdown
$ dgtr review
# Daily Google Tasks Review

> **Period:** 2026-08-05 00:00 ~ 2026-08-06 20:30

### Completed (2)

- **QA review of dashboard** *(completed 2026-08-05 | updated 14:32)*
- **Reply to vendor contract** *(completed 2026-08-05 | updated 16:20)*

### Open (1)

- **Ship v1.0.0** *(due 2026-08-10 | updated 11:15)*
```

## Why

Most daily-review tools wrap an LLM and charge per call. `dgtr` doesn't:
it fetches your tasks and formats them deterministically. Zero marginal cost,
instant, and works offline-friendly.

## Features

- **`dgtr login`** — one-time OAuth 2.0 PKCE authorization → stores a long-lived
  refresh token in `~/.dgtr/config.json` (auto-refreshes afterwards, no re-login).
- **`dgtr review`** (or **`dgtr brief`**) — reviews all tasks modified from yesterday 00:00 until now.
- **`dgtr open`** (or **`dgtr all`**, **`dgtr tasks`**) — list all open tasks regardless of modification date.
- **`dgtr review --date YYYY-MM-DD`** — target a specific day's modifications.
- Single static binary. Executable from any working directory (`cwd`).

## Install

### Homebrew (once published) / download a release
Grab the latest binary from [Releases](../../releases) for your OS, or:

```bash
go install github.com/behumble/dgtr@latest
```

### Build from source
```bash
git clone https://github.com/behumble/dgtr
cd dgtr
go build -o dgtr .
```

## Setup (one time)

Simply run the one-time OAuth authorization:

```bash
dgtr login
```

Your browser opens → sign in with your Google account → done. The refresh token is saved to `~/.dgtr/config.json` automatically. Subsequent runs require no re-login and work from any working directory (`cwd`).

(Optional: If you wish to use a custom config file path, pass `--config /path/to/config.json` or set `DGTR_CONFIG`.)

## Usage

```bash
dgtr review                 # modified tasks from yesterday 00:00 until now
dgtr open                   # all open tasks (or: dgtr all, dgtr tasks)
dgtr review --date 2026-08-03
dgtr login --force          # re-authorize
dgtr version
```

Use `--config /path/to/config.json` (or `DGTR_CONFIG`) to point at a non-default config file,
e.g. when automating with cron.

## Automating with cron

Because `dgtr` costs nothing per run, schedule it freely:

```cron
# every weekday at 08:00
0 8 * * 1-5  /usr/local/bin/dgtr review
```

## Security notes

- Your `~/.dgtr/config.json` token file is **private**. It is created with restrictive `0600` permissions (readable/writable only by your user account).
- The refresh token is scoped to **tasks only** (`https://www.googleapis.com/auth/tasks`),
  so a leaked token can't touch your email, drive, calendar, etc.
- `dgtr login` binds the OAuth callback to `127.0.0.1` and validates the
  `state` parameter to prevent CSRF.
- Revoke access any time at https://myaccount.google.com/permissions

## License

[MIT](LICENSE)

---

## 한국어 (Korean)

[![한국어](https://img.shields.io/badge/Language-%ED%95%9C%EA%B5%AD%EC%96%B4-blue.svg)](README.ko.md)
한국어 버전: [README.ko.md](README.ko.md)

