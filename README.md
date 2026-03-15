# blueprint-vibe-hooks

Claude Code 플러그인 — Go 기반 hooks: guard 시스템, 세션 관리, Python 검증.

## 기능

| Hook Event | 동작 |
|------------|------|
| `SessionStart` | 환경 검증 (jq, uv, guard.json 존재 여부) |
| `UserPromptSubmit` | `/team-loops` 감지 → 세션 마커 저장 |
| `PreToolUse (Read)` | guard.json `read.blockedPatterns` 검사 |
| `PreToolUse (Write\|Edit)` | guard.json `write.blockedPatterns` 검사 |
| `PreToolUse (Bash)` | guard.json `bash.blockedCommands` + `blockedPatterns` 검사 |
| `PostToolUse (Write\|Edit)` | `.py` 파일에 ruff format 자동 적용 |
| `Stop / SubagentStop` | team-loop 계속/중단 확인 → verify (ruff, mypy, pytest) |

## 빌드

```shell
./scripts/setup.sh
```

바이너리 생성 위치: `bin/claude-code-hooks`

## 로컬 테스트

```shell
# 로컬 마켓플레이스로 추가
/plugin marketplace add ./packages/claude-plugin

# 플러그인 설치
/plugin install blueprint-vibe-hooks@blueprint-vibe-hooks

# 설치 확인
/plugin list
```

## 마켓플레이스 등록

GitHub 레포지토리에 push 후 `extraKnownMarketplaces`에 등록:

```json
"extraKnownMarketplaces": {
  "blueprint-vibe": {
    "source": {
      "source": "github",
      "repo": "neurumaru/blueprint-vibe"
    }
  }
}
```

## 구조

```
packages/claude-plugin/
├── .claude-plugin/
│   ├── plugin.json          # 플러그인 메타데이터
│   └── marketplace.json     # 마켓플레이스 등록 정보
├── agents/                  # Claude Code 에이전트 정의
├── commands/                # Claude Code 커맨드
├── skills/                  # Claude Code 스킬
├── hooks/
│   └── hooks.json           # hook 설정 (${CLAUDE_PLUGIN_ROOT} 경로 사용)
├── guard.json               # 접근 제어 규칙
├── bin/
│   └── claude-code-hooks    # 빌드된 바이너리 (git 제외)
├── scripts/
│   └── setup.sh             # 빌드 스크립트
└── tools/                   # Go 소스
    ├── cmd/claude-code-hooks/
    └── internal/
        ├── application/hook/      # guard 로직
        ├── domain/session/
        ├── infrastructure/
        └── presentation/cli/      # 각 hook event 핸들러
```
