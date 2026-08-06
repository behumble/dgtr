# dgtr — Daily Google Tasks Review

어제 00시부터 현재까지 구글 태스크(Google Tasks)의 변경/수정 내역을 리뷰해 주는 가벼운 단일 바이너리 CLI입니다.
LLM을 쓰지 않고, 요청당 토큰 비용도 없습니다 — **구글 Tasks API에 직접 접속**해
규칙 기반으로 요약을 만듭니다. 자격증명은 각자 자신의 `.env`에 두므로, 누구나
자기 구글 계정으로 사용할 수 있습니다.

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
`dgtr`은 그렇지 않습니다 — 어제 00시 이후 변경된 태스크를 직접 가져와 **결정적으로(format)** 출력합니다.
한계 비용 0, 즉시 실행, 오프라인에서도 잘 동작합니다.

## 기능

- **`dgtr login`** — 1회 OAuth 2.0 인증 → 장기 유효 refresh token을 `.env`에 저장
  (이후 자동 갱신, 재로그인 불필요)
- **`dgtr review`** (또는 **`dgtr brief`**) — 어제 00시부터 지금까지 변경된 모든 태스크 리뷰
- **`dgtr open`** (또는 **`dgtr all`**, **`dgtr tasks`**) — 수정일과 무관하게 열려 있는 모든 태스크 출력
- **`dgtr review --date YYYY-MM-DD`** — 특정 날짜에 변경된 태스크 지정
- 단일 정적 바이너리. macOS / Linux / Windows 지원

## 설치

### 바이너리(릴리즈) / Go 설치
[Releases](../../releases)에서 OS에 맞는 최신 바이너리를 받거나:

```bash
go install github.com/behumble/dgtr@latest
```

### 소스에서 빌드
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

브라우저가 열리면 로그인 → 완료. refresh token이 자동으로 `.env`에 저장됩니다. 이후에는 재로그인이 필요 없습니다.

(선택사항: 본인 소유의 커스텀 OAuth Client ID를 사용하고 싶다면 `.env`에 `GOOGLE_TASKS_CLIENT_ID`를 설정할 수 있습니다.)

## 사용법

```bash
dgtr review                 # 어제 00시부터 지금까지 변경된 태스크 리뷰
dgtr open                   # 모든 열린 태스크 (또는 dgtr all, dgtr tasks)
dgtr review --date 2026-08-03
dgtr login --force          # 재인증
dgtr version
```

기본이 아닌 env 파일을 가리키려면 `--env /path/to/.env`(또는 `DGTR_ENV`)를 쓰세요.
cron으로 자동화할 때 유용합니다.

## cron으로 자동화

`dgtr`은 실행마다 비용이 0이므로 자유롭게 예약하세요:

```cron
# 평일 매일 08:00
0 8 * * 1-5  cd /path/to/project && ./dgtr review
```

## 보안 참고

- 당신의 `.env`(client secret + refresh token)는 **비공개**입니다. `.gitignore`가
  기본적으로 제외하므로 절대 커밋하지 마세요.
- refresh token은 **tasks 전용 스코프**(`https://www.googleapis.com/auth/tasks`)로
  제한되어 있어, 유출돼도 이메일·드라이브·캘린더 등에는 닿을 수 없습니다.
- `dgtr login`은 OAuth 콜백을 `127.0.0.1`에 바인딩하고 `state` 매개변수를 검증해
  CSRF를 방지합니다.
- 접근 권한은 언제든 https://myaccount.google.com/permissions 에서 해제할 수 있습니다.

## 라이선스

[MIT](LICENSE)

---

## English

English version: [README.md](README.md)

