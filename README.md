# dgtb — Daily Google Tasks Brief

A tiny, single-binary CLI that prints a concise briefing of your Google Tasks.
No LLM, no per-request token cost — it talks **directly to the Google Tasks
API** and renders a rule-based summary. Your credentials stay in your own
`.env`, so anyone can use it with their own Google account.

```
$ dgtb brief
## Task briefing for 2026-08-03

**Completed:**
- [x] 2026-08-03 QA review of dashboard ✓ completed 2026-08-03
- [x] Reply to vendor contract ✓ completed 2026-08-02

**Open (due on/before today):**
- [ ] Ship v1.0.0 (due 2026-08-05)
```

## Why

Most daily-briefing tools wrap an LLM and charge per call. `dgtb` doesn't:
it fetches your tasks and formats them deterministically. Zero marginal cost,
instant, and works offline-friendly.

## Features

- **`dgtb login`** — one-time OAuth 2.0 authorization → stores a long-lived
  refresh token in your `.env` (auto-refreshes afterwards, no re-login).
- **`dgtb brief`** — yesterday's completed tasks + open tasks due on/before
  today.
- **`dgtb brief --date YYYY-MM-DD`** — target a specific day.
- **`dgtb brief --all`** — list every open task regardless of due date.
- Single static binary. macOS / Linux / Windows.

## Install

### Homebrew (once published) / download a release
Grab the latest binary from [Releases](../../releases) for your OS, or:

```bash
go install github.com/alangoo/dgtb@latest
```

### Build from source
```bash
git clone https://github.com/alangoo/dgtb
cd dgtb
go build -o dgtb .
```

## Setup (one time, per user)

Every user creates their **own** OAuth client — nothing is shared.

1. **Create a Google Cloud project** (or reuse one):
   https://console.cloud.google.com/projectselector2

2. **Enable the Google Tasks API** for that project:
   https://console.cloud.google.com/apis/library/tasks.googleapis.com

3. **Create a Desktop OAuth client**:
   https://console.cloud.google.com/apis/credentials \
   `Create credentials → OAuth client ID → Application type: Desktop app`
   (If the app is still in *Testing*, add your Google account under
   `Audience → Test users`.)

4. **Copy `.env.example` to `.env`** and fill in your client id/secret:
   ```bash
   cp .env.example .env
   # edit .env -> GOOGLE_TASKS_CLIENT_ID / GOOGLE_TASKS_CLIENT_SECRET
   ```

5. **Authorize:**
   ```bash
   dgtb login
   ```
   Your browser opens → sign in → done. A refresh token is saved to `.env`
   automatically. Subsequent runs need no re-login.

## Usage

```bash
dgtb brief                  # yesterday + open tasks
dgtb brief --date 2026-08-03
dgtb brief --all
dgtb login --force          # re-authorize
dgtb version
```

Use `--env /path/to/.env` (or `DGTB_ENV`) to point at a non-default env file,
e.g. when automating with cron.

## Automating with cron

Because `dgtb` costs nothing per run, schedule it freely:

```cron
# every weekday at 08:00
0 8 * * 1-5  cd /path/to/project && ./dgtb brief
```

## Security notes

- Your `.env` (client secret + refresh token) is **private**. `.gitignore`
  excludes it by default — never commit it.
- The refresh token is scoped to **tasks only** (`https://www.googleapis.com/auth/tasks`),
  so a leaked token can't touch your email, drive, calendar, etc.
- `dgtb login` binds the OAuth callback to `127.0.0.1` and validates the
  `state` parameter to prevent CSRF.
- Revoke access any time at https://myaccount.google.com/permissions

## License

[MIT](LICENSE)

---

## 한국어 (Korean)

한국어 버전: [README.ko.md](README.ko.md)
