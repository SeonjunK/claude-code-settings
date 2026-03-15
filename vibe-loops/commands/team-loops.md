---
description: 현재 작업을 분석해 적절한 team을 만들고, Task 기반 continuous queue를 유지합니다.
argument-hint: "[PROMPT] [--max-iterations N] [--max-parallel N]"
allowed-tools:
  - "Bash(team-loops:*)"
  - "TeamCreate"
  - "TeamDelete"
  - "SendMessage"
  - "TaskCreate"
  - "TaskUpdate"
  - "TaskList"
  - "TaskGet"
  - "Agent"
---

/team-loops $ARGUMENTS

# Continuous Queue Orchestration (Go CLI)

이 command는 **Teams-only, task-centric continuous queue**를 전제로 동작합니다.
핵심은 `계획 → 병렬 실행 → 합류 → 검증 → 다음 라운드`가 아니라, **team task list를 source of truth로 유지하면서 병렬 작업을 계속 이어가는 구조**입니다.

## Step 0 — Setup

```!
"${CLAUDE_PLUGIN_ROOT}/bin/claude-code-hooks" start "$ARGUMENTS"
```

이 명령은 세션 파일을 생성하고 초기 설정을 반환합니다.

`vibe.json`에 notifications가 설정되어 있으면 세션 시작 알림이 자동 전송됩니다.

## Step 1 — TeamCreate

세션 파일에서 team_name을 읽어 TeamCreate 호출

```javascript
TeamCreate({
  team_name: "<from session>",
  description: "<goal description>"
})
```

## Step 2 — Spawn Teammates

각 teammate에 대해 Agent tool 호출:
```javascript
Agent({
  subagent_type: "<role>",  // impl, quality, docs, security, refactor, ui-ux, idea
  prompt: "...",
  team_name: "<team_name>",
  name: "<teammate_name>",
  run_in_background: true
})
```

## Step 3 — Monitor

```!
"${CLAUDE_PLUGIN_ROOT}/bin/claude-code-hooks" status
```

Telegram bot이 실행 중이면 `/status`, `/tasks` 명령으로도 조회 가능합니다.

## Step 4 — Notify (선택)

수동 알림이 필요한 경우:
```shell
"${CLAUDE_PLUGIN_ROOT}/bin/claude-code-hooks" notify task_complete -m "API 응답 포맷 수정 완료"
```

## Step 5 — Shutdown

완료 시:
```!
"${CLAUDE_PLUGIN_ROOT}/bin/claude-code-hooks" stop-session --reason "completed"
```

---

## Runtime Contract

실제 contract는 세션의 `init.tools`를 기준으로 판단하세요.
현재 기준으로 전제하는 도구는 아래입니다.

- `TeamCreate`
- `TeamDelete`
- `SendMessage`
- `Task` — shared task list 생성/조회/갱신
- `TaskOutput` — 백그라운드 task 상태 확인
- `TaskStop` — 장시간 멈춘 task 정리
- `Agent` — teammate spawn

현재 contract에 **없는 것**:

- standalone `Teammate` tool

공식 teammate lifecycle은 team/task/message 도구와 hook event로 해석하세요.
headless에서 hidden `Agent`가 callable이어도 이 workflow는 spawn 경로에 의존하지 않습니다.
이미 세션에 보이는 teammate 이름만 recipient/owner로 사용하세요.

권장 owner:

- `impl`
- `quality`
- `docs`
- `security`
- `refactor`
- `ui-ux`
- `idea`

보이지 않는 이름은 owner나 recipient로 쓰지 마세요.

## Core Model

- team task list가 **유일한 runtime source of truth**입니다.
- hard round, 중앙 합류, 라운드 종료 배리어를 두지 않습니다.
- 각 task는 `pending`, `in_progress`, `completed` 상태를 기준으로 해석합니다.
- blocked work는 별도 status가 아니라 `status=pending` + non-empty `blockedBy`로 표현합니다.
- `owner` 필드는 현재 lease holder를 의미합니다.
- follow-up은 텍스트 보고가 아니라 **새 task 생성**으로 남깁니다.

Task 도구 사용:

- `Task` — 새 task 생성, 상태/owner 변경, 목록 조회, 상세 조회를 모두 처리합니다.
- 세션별 실제 input shape는 live contract를 따르되, 운영 규칙은 항상 shared task list 기준으로 해석합니다.

백그라운드 task 관리:

- `TaskOutput` — 백그라운드 실행 중인 task 결과 확인
- `TaskStop` — 멈춘 task 정리

## Step 1 — TeamCreate

작업이 시작되면 가장 먼저 team을 만드세요.

- 문서를 읽기 전
- 목표를 확정하기 전
- seed task를 만들기 전

이 단계에서는 아래만 정합니다.

- 임시 목표 1문장
- prompt 1줄
- 제약 3개 이내

예시:

```text
임시 목표: 문서 기반 다음 목표를 찾아 continuous queue 시작
PROMPT: <$ARGUMENTS 또는 없음>
제약: 문서 근거 우선 / 검증 유지 / 사용자 명시 종료 전 팀 유지
```

## Step 2 — Goal Lock

team이 만들어진 뒤 아래를 확정하세요.

- 목표 1문장
- 완료 조건
- source 문서
- 사용할 owner 집합

사용자 목표가 없으면 문서에서 다음 목표를 고릅니다.

우선순위:

1. `docs/actions/README.md`, `docs/actions/*.md`
2. `docs/adr/README.md`, `docs/adr/ADR-*.md`
3. `docs/architecture/**/*.md`
4. 최근 검증 실패, 코드-문서 불일치, 명백한 품질/보안 갭

회피 규칙:

- 대규모 리디자인
- 제품 결정이 필요한 신규 기능
- 외부 자격증명 없이는 검증 불가한 작업
- 문서 근거 없는 임의 아이디어

## Step 3 — Seed Queue

Task 도구로 초기 task를 만드세요.
각 task는 한 가지 목적만 가져야 하고, 20~40분 이상이면 더 쪼개세요.

표준 필드:

- `subject`: 짧은 작업 제목
- `description`: owner가 바로 실행할 수 있는 작업 설명
- `owner`: 현재 담당 owner 이름
- `status`: `pending | in_progress | completed`
- `blockedBy`: dependency나 외부 입력만 표현
- `metadata.goal_id`
- `metadata.definition_of_done`
- `metadata.verification_scope`
- `metadata.source_doc`
- `metadata.parent_task_id`

`metadata.verification_scope` 값:

- `none`
- `local`
- `global`
- `approval`

기본 seed 규칙:

- 구현/분석 task 외에 `quality` owner의 verify task를 최소 1개 생성
- 외부 입력, 권한, 비밀정보, shell 실행이 있으면 `security` task 생성
- 문서 갱신 가능성이 있으면 `docs` task 생성
- 구조 변경이 핵심이면 `refactor` task 생성
- UI 영향이 있으면 `ui-ux` task 생성
- 탐색/ADR 초안이 핵심이면 `idea` task 생성

예시:

```text
T-101 impl      API 응답 포맷 수정
T-102 quality   변경 영역 local verify
T-103 docs      문서 동기화 필요 여부 확인
T-104 security  입력 검증/권한 점검
```

## Step 4 — Broadcast and Direct Messages

초기 queue가 만들어지면 `SendMessage`를 사용해 현재 규칙을 공지하세요.

권장 broadcast 본문:

```text
목표: <한 줄 목표>
source: <문서 경로>
queue source of truth: team task list
규칙:
- claim한 task만 수행
- 막히면 `blockedBy`를 채우고 unblock task 생성
- 끝나면 completed로 바꾸고 follow-up/verify task 생성
- 중앙 합류를 기다리지 말 것
```

개별 task context가 더 필요하면 owner에게 direct `message`를 보냅니다.

## Step 5 — Claim / Execute / Release

각 owner는 아래 규칙을 따릅니다.

### Claim

- `blockedBy`가 비어 있는 `pending` task만 가져갑니다.
- 가져갈 때 `status=in_progress`, `owner=<teammate_name>`으로 갱신합니다.
- 이미 다른 owner가 잡은 task는 가로채지 않습니다.

### Execute

- task의 `definition_of_done` 범위만 처리합니다.
- scope를 넓히지 않습니다.
- 추가 가치가 보이면 follow-up을 새 task로 만듭니다.

### Block

- 외부 입력이나 dependency 때문에 막히면 `status=pending`으로 되돌리고 `blockedBy`를 채웁니다.
- `blockedBy`를 구체적으로 채웁니다.
- 필요한 unblock/fix task를 새로 만듭니다.

### Release

- 끝나면 `status=completed`로 바꿉니다.
- 필요한 verify task를 즉시 생성합니다.
- 필요한 경우 `TaskOutput`으로 결과를 남기고, direct `SendMessage`로 관련 owner를 깨웁니다.

## Verification Rules

검증은 라운드 종료 단계가 아니라 queue 내부 규칙입니다.

### Local Verify

task 완료 직후 필요한 범위의 검증 task를 만듭니다.

예시:

- 구현 완료 → `quality` local verify
- 민감 입력 변경 → `security` review
- 구조 변경 → `refactor` or `docs` sync review

### Global Verify

idle 직전에는 전역 검증이 필요합니다.
최종 gate는 `Stop`/`SubagentStop` hook의 `claude-code-hooks verify`가 맡습니다.

verify 실패 규칙:

- 종료하지 않습니다.
- 실패 원인을 새 fix task로 queue에 재주입합니다.
- global verify가 통과할 때까지 idle로 가지 않습니다.

## Approval-Gated Work

승인이 필요한 산출물은 새로운 lifecycle을 만들지 말고 task 상태로 표현하세요.

표준 표현:

- `status=pending`
- `blockedBy=approval_needed`
- `metadata.verification_scope=approval`

## User Control and Idle

queue 내부 상태는 task list가 기준입니다.

### Idle 진입 조건

아래를 모두 만족할 때만 idle로 전환합니다.

- 진행 가능한 `pending` task가 없음
- 의미 있는 unblock/follow-up이 없음
- 최신 global verify 통과
- 사용자 새 지시 없음

### Idle 해제 조건

아래 중 하나가 생기면 queue를 재개합니다.

- 새 사용자 지시
- unblock
- 새 문서 기반 목표 선택
- verify 실패로 생성된 fix task

### TeamDelete

`TeamDelete`는 아래 경우에만 사용합니다.

- 사용자가 명시적으로 중단 요청
- queue 상태 요약 완료
- 더 이상 stop 차단 세션을 유지할 이유 없음

idle 상태에서는 team을 삭제하지 마세요.

## Operational Rules

- task list가 항상 source of truth입니다.
- hard round와 중앙 합류를 만들지 마세요.
- 새 문제를 발견하면 텍스트 보고보다 task 생성이 우선입니다.
- `TaskOutput`은 추적이 필요할 때만 사용하세요.
- `TaskStop`은 장시간 멈춘 task를 정리할 때만 사용하세요.
- 검증 실패는 종료 신호가 아니라 새 fix task 생성 신호입니다.
- 사용자가 명시적으로 멈추기 전까지는 가치 있는 task가 있으면 계속 이어가세요.
- 가치 있는 task가 없으면 자동 종료하지 말고 idle로 전환하세요.
