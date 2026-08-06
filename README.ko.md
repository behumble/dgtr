# dgtb — Daily Google Tasks Brief

구글 태스크(Google Tasks)를 간결한 브리핑으로 출력하는 가벼운 단일 바이너리 CLI입니다.
LLM을 쓰지 않고, 요청당 토큰 비용도 없습니다 — **구글 Tasks API에 직접 접속**해
규칙 기반으로 요약을 만듭니다. 자격증명은 각자 자신의 `.env`에 두므로, 누구나
자기 구글 계정으로 사용할 수 있습니다.

```
$ dgtb brief
## Task briefing for 2026-08-03

**Completed:**
- [x] 2026-08-03 QA review of dashboard ✓ completed 2026-08-03
- [x] Reply to vendor contract ✓ completed 2026-08-02

**Open (due on/before today):**
- [ ] Ship v1.0.0 (due 2026-08-05)
```

## 왜 만들었나

대부분의 일일 브리핑 도구는 LLM을 붙여서 호출할 때마다 비용이 청구됩니다.
`dgtb`는 그렇지 않습니다 — 태스크를 직접 가져와 **결정적으로(format)** 출력합니다.
한계 비용 0, 즉시 실행, 오프라인에서도 잘 동작합니다.

## 기능

- **`dgtb login`** — 1회 OAuth 2.0 인증 → 장기 유효 refresh token을 `.env`에 저장
  (이후 자동 갱신, 재로그인 불필요)
- **`dgtb brief`** — 어제 완료된 태스크 + 오늘(또는 이전) 마감인 열린 태스크
- **`dgtb brief --date YYYY-MM-DD`** — 특정 날짜 지정
- **`dgtb brief --all`** — 마감일과 무관하게 모든 열린 태스크 출력
- 단일 정적 바이너리. macOS / Linux / Windows 지원

## 설치

### 바이너리(릴리즈) / Go 설치
[Releases](../../releases)에서 OS에 맞는 최신 바이너리를 받거나:

```bash
go install github.com/behumble/dgtb@latest
```

### 소스에서 빌드
```bash
git clone https://github.com/behumble/dgtb
cd dgtb
go build -o dgtb .
```

## 설정 (사용자당 1회)

각 사용자가 **자신만의** OAuth 클라이언트를 만듭니다 — 공유되는 것은 없습니다.

1. **Google Cloud 프로젝트 생성**(또는 기존 것 사용):
   https://console.cloud.google.com/projectselector2

2. 해당 프로젝트에서 **Google Tasks API 사용 설정**:
   https://console.cloud.google.com/apis/library/tasks.googleapis.com

3. **데스크톱(Desktop) OAuth 클라이언트 생성**:
   https://console.cloud.google.com/apis/credentials \
   `사용자 인증 정보 만들기 → OAuth 클라이언트 ID → 애플리케이션 유형: 데스크톱 앱`
   (앱이 아직 *Testing* 상태라면, `API 및 서비스 → OAuth 동의 화면 → 테스트 사용자`에
   본인 구글 계정을 추가하세요.)

4. **`.env.example`을 `.env`로 복사**하고 클라이언트 id/secret 입력:
   ```bash
   cp .env.example .env
   # .env 편집 -> GOOGLE_TASKS_CLIENT_ID / GOOGLE_TASKS_CLIENT_SECRET
   ```

5. **인증:**
   ```bash
   dgtb login
   ```
   브라우저가 열리면 로그인 → 완료. refresh token이 자동으로 `.env`에 저장됩니다.
   이후에는 재로그인이 필요 없습니다.

## 사용법

```bash
dgtb brief                  # 어제 + 열린 태스크
dgtb brief --date 2026-08-03
dgtb brief --all
dgtb login --force          # 재인증
dgtb version
```

기본이 아닌 env 파일을 가리키려면 `--env /path/to/.env`(또는 `DGTB_ENV`)를 쓰세요.
cron으로 자동화할 때 유용합니다.

## cron으로 자동화

`dgtb`는 실행마다 비용이 0이므로 자유롭게 예약하세요:

```cron
# 평일 매일 08:00
0 8 * * 1-5  cd /path/to/project && ./dgtb brief
```

## 보안 참고

- 당신의 `.env`(client secret + refresh token)는 **비공개**입니다. `.gitignore`가
  기본적으로 제외하므로 절대 커밋하지 마세요.
- refresh token은 **tasks 전용 스코프**(`https://www.googleapis.com/auth/tasks`)로
  제한되어 있어, 유출돼도 이메일·드라이브·캘린더 등에는 닿을 수 없습니다.
- `dgtb login`은 OAuth 콜백을 `127.0.0.1`에 바인딩하고 `state` 매개변수를 검증해
  CSRF를 방지합니다.
- 접근 권한은 언제든 https://myaccount.google.com/permissions 에서 해제할 수 있습니다.

## 라이선스

[MIT](LICENSE)

---

## English

English version: [README.md](README.md)
