# dgtr — Daily Google Tasks Review

[![English](https://img.shields.io/badge/Language-English-blue.svg)](README.md)

어제 00시부터 현재까지 구글 태스크(Google Tasks)의 변경/수정 내역을 깔끔한 **마크다운(Markdown)** 형식으로 출력해 주는 가벼운 단일 바이너리 CLI입니다.
LLM을 쓰지 않고, 요청당 토큰 비용도 없습니다 — **구글 Tasks API에 직접 접속**해
규칙 기반으로 요약을 만듭니다. OAuth 인증 정보는 **`~/.dgtr/config.json`**에 안전하게 보관되므로,
누구나 자기 구글 계정으로 터미널 위치(cwd)와 무관하게 사용할 수 있습니다.

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

## 왜 만들었나

대부분의 일일 리뷰/브리핑 도구는 LLM을 붙여서 호출할 때마다 비용이 청구됩니다.
`dgtr`은 그렇지 않습니다 — 어제 00시 이후 변경된 태스크를 직접 가져와 마크다운으로 **결정적으로(format)** 출력합니다.
한계 비용 0, 즉시 실행, 오프라인에서도 잘 동작합니다.

## 기능

- **`dgtr login`** — 1회 OAuth 2.0 PKCE 인증 → 장기 유효 refresh token을 `~/.dgtr/config.json`에 저장
  (이후 자동 갱신, 재로그인 불필요)
- **`dgtr review`** (또는 **`dgtr brief`**) — 어제 00시부터 지금까지 변경된 태스크를 마크다운 리포트로 출력
- **`dgtr open`** (또는 **`dgtr all`**, **`dgtr tasks`**) — 수정일과 무관하게 열려 있는 모든 태스크를 마크다운 목록으로 출력
- **`dgtr review --date YYYY-MM-DD`** — 특정 날짜에 변경된 태스크 지정
- 단일 정적 바이너리. 터미널의 어떤 작업 경로(`cwd`)에서든 실행 가능.

## 설치

### Go 사용자의 경우 (`go install`)
Go가 설치되어 있다면 다음 명령어로 설치할 수 있습니다:

```bash
go install github.com/behumble/dgtr@latest
```

*(참고: 설치 후 `dgtr` 명령어를 찾을 수 없다면 `~/go/bin`이 `PATH`에 등록되어 있는지 확인하세요: `export PATH="$HOME/go/bin:$PATH"`)*

### 사전 빌드된 바이너리 다운로드 (Pre-built Binary)
[Releases](../../releases) 페이지에서 본인의 OS(macOS, Linux, Windows)에 맞는 실행 바이너리를 직접 다운로드하세요.

### 소스에서 직접 빌드
```bash
git clone https://github.com/behumble/dgtr
cd dgtr
go build -o dgtr .
```

## 설정 (1회만 수행)

원타임 OAuth 인증만 실행하면 됩니다:

```bash
dgtr login
```

`dgtr` 실행 시 GCP OAuth Client ID 및 Client Secret을 묻는 프롬프트가 실행됩니다 (또는 `--client-id`, `--client-secret` 플래그 전달 가능).
사용자가 작성한 정적 설정은 `~/.dgtr/config.json`에, 프로그램이 자동 생성한 토큰은 `~/.dgtr/credentials.json`에 각각 분리되어 안전하게 저장됩니다 (`0600` 권한).
이후에는 재로그인 없이 어디서나 자유롭게 실행할 수 있습니다.

## 사용법

```bash
dgtr review                 # 어제 00시부터 지금까지 변경된 태스크 리뷰
dgtr open                   # 모든 열린 태스크 (또는 dgtr all, dgtr tasks)
dgtr review --date 2026-08-03
dgtr login --force          # 재인증
dgtr version
```

기본 위치가 아닌 커스텀 설정 파일을 사용하려면 `--config /path/to/config.json`(또는 `DGTR_CONFIG`)을 지정하세요.

## cron으로 자동화

`dgtr`은 실행마다 비용이 0이므로 자유롭게 예약하세요:

```cron
# 평일 매일 08:00
0 8 * * 1-5  /usr/local/bin/dgtr review
```

## 보안 참고

- 인증 정보가 담긴 `~/.dgtr/config.json` 파일은 **비공개**입니다. 사용자 전용 권한(`0600`)으로 안전하게 저장됩니다.
- refresh token은 **tasks 전용 스코프**(`https://www.googleapis.com/auth/tasks`)로
  제한되어 있어, 유출돼도 이메일·드라이브·캘린더 등에는 닿을 수 없습니다.
- `dgtr login`은 OAuth 콜백을 `127.0.0.1`에 바인딩하고 `state` 매개변수를 검증해
  CSRF를 방지합니다.
- 접근 권한은 언제든 https://myaccount.google.com/permissions 에서 해제할 수 있습니다.

## 라이선스

[MIT](LICENSE)

