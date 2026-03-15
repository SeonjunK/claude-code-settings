# 공통 Teammate 규칙

## 공통 참조

공통 규칙은 `CLAUDE.md`, `docs/guides/README.md`, 그리고 전역 command hook(`verify.sh`)을 따르세요.

## Team Queue 규칙

`/team-loops` 세션에서는 team task list가 source of truth입니다.

- claim된 task의 `definition_of_done` 범위만 처리하세요
- 중앙 합류를 기다리지 말고, 끝나면 `completed` 가능한 결과를 남기세요
- 추가 가치가 보이면 새 follow-up task를 제안하세요
- 막히면 `blockedBy` 사유와 필요한 unblock 작업을 짧게 남기세요
- 실패를 발견하면 종료가 아니라 새 fix task를 만드세요
