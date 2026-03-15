# claude-code-settings

Claude Code 플러그인 마켓플레이스 — 4개 독립 플러그인으로 필요한 기능만 선택 설치.

## 플러그인

| Plugin | 역할 | Hook Events |
|--------|------|-------------|
| **vibe-guard** | 파일/명령어 접근 제어 | PreToolUse |
| **vibe-format** | Go/Python 자동 포매팅 | PostToolUse |
| **vibe-verify** | 검증 파이프라인 + 환경 체크 | SessionStart, Stop, SubagentStop |
| **vibe-loops** | team-loops 세션 + 알림 + bot | Stop, SubagentStop, UserPromptSubmit |

## 설치

```shell
# 마켓플레이스 추가
/plugin marketplace add seonjunk

# 원하는 플러그인만 설치
/plugin install vibe-guard@seonjunk
/plugin install vibe-format@seonjunk
/plugin install vibe-verify@seonjunk
/plugin install vibe-loops@seonjunk
```

## vibe-guard

파일 읽기/쓰기와 Bash 명령어에 대한 접근 제어. `vibe.json`으로 차단 패턴 설정.

```json
{
  "guard": {
    "read":  { "blockedPatterns": [".env*", "*.pem", "*.key"] },
    "write": { "blockedPatterns": [".env*", "*.pem", "*.key"] },
    "bash": {
      "blockedCommands": ["rm -rf /"],
      "blockedPatterns": ["git push --force", "drop table"]
    }
  }
}
```

## vibe-format

Write/Edit 후 자동 포매팅:
- `.go` → gofmt
- `.py` → ruff format (또는 black)

## vibe-verify

- **SessionStart**: 환경 검증 (jq, uv, vibe.json)
- **Stop/SubagentStop**: 검증 파이프라인 (shell syntax, JSON validation, Python lint/test)

## vibe-loops

team-loops 세션 관리, 알림, Telegram bot.

### CLI 명령어

```
vibe-loops start <goal>         새 team-loops 세션 시작
vibe-loops status               현재 세션 상태 표시
vibe-loops stop-session         현재 세션 중단
vibe-loops notify <event>       알림 수동 전송
vibe-loops telegram-bot         Telegram bot 시작
```

### Telegram Bot

```shell
vibe-loops telegram-bot --token $BOT_TOKEN --chat-id $CHAT_ID
```

| 명령 | 동작 |
|------|------|
| `/status` | 현재 세션 상태 조회 |
| `/tasks` | task 목록 요약 |
| `/stop` | 활성 세션 중단 |
| `/help` | 사용 가능한 명령어 |

### 알림 설정 (vibe.json)

```json
{
  "notifications": {
    "telegram": {
      "enabled": true,
      "botToken": "${TELEGRAM_BOT_TOKEN}",
      "chatId": "${TELEGRAM_CHAT_ID}"
    },
    "slack": {
      "enabled": false,
      "webhookUrl": "${SLACK_WEBHOOK_URL}",
      "channel": "#claude-notifications"
    }
  }
}
```

## 빌드

각 플러그인은 첫 실행 시 자동 빌드됩니다 (`ensure-binary.sh`).

수동 빌드:
```shell
cd tools
go build ./cmd/vibe-guard
go build ./cmd/vibe-format
go build ./cmd/vibe-verify
go build ./cmd/vibe-loops
```

테스트:
```shell
cd tools && go test ./...
```

## 마켓플레이스 등록

```json
"extraKnownMarketplaces": {
  "seonjunk": {
    "source": {
      "source": "github",
      "repo": "SeonjunK/claude-code-settings"
    }
  }
}
```

## 구조

```
claude-code-settings/
├── .claude-plugin/
│   └── marketplace.json          # 4개 플러그인 목록
├── vibe-guard/
│   ├── .claude-plugin/plugin.json
│   ├── hooks/hooks.json          # PreToolUse
│   ├── vibe.json                 # guard 설정
│   ├── scripts/ensure-binary.sh
│   ├── tools -> ../tools
│   └── bin/
├── vibe-format/
│   ├── .claude-plugin/plugin.json
│   ├── hooks/hooks.json          # PostToolUse
│   ├── scripts/ensure-binary.sh
│   ├── tools -> ../tools
│   └── bin/
├── vibe-verify/
│   ├── .claude-plugin/plugin.json
│   ├── hooks/hooks.json          # SessionStart, Stop, SubagentStop
│   ├── scripts/ensure-binary.sh
│   ├── tools -> ../tools
│   └── bin/
├── vibe-loops/
│   ├── .claude-plugin/plugin.json
│   ├── hooks/hooks.json          # Stop, SubagentStop, UserPromptSubmit
│   ├── commands/team-loops.md
│   ├── agents/ (8 files)
│   ├── skills/interview/
│   ├── vibe.json                 # notifications 설정
│   ├── scripts/ensure-binary.sh
│   ├── tools -> ../tools
│   └── bin/
├── tools/                        # 공유 Go 모듈
│   ├── go.mod
│   ├── cmd/
│   │   ├── vibe-guard/main.go
│   │   ├── vibe-format/main.go
│   │   ├── vibe-verify/main.go
│   │   └── vibe-loops/main.go
│   └── internal/
├── README.md
└── .gitignore
```
