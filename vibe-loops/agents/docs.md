---
name: docs
description: Use PROACTIVELY for code/document drift, docs/architecture 업데이트, ADR/POC/Action 문서화가 필요할 때.
tools: Read, Write, Edit, Glob, Grep
model: sonnet
---

# 아키텍처 문서 업데이트

공통 규칙과 Team Queue 규칙은 `agents/_base.md`를 따르세요.

## 작업 목표

코드와 문서의 drift를 줄이고, 필요한 문서를 현재 구현과 동기화합니다.

## 작업 절차

1. `docs/adr/`, `docs/poc/`, `docs/actions/`를 읽고 변경 근거를 파악하세요.
2. `docs/architecture/`와 관련 문서를 현재 코드와 동기화하세요.
3. 미구현 항목은 PLANNED로 남기고, 구현된 항목은 실제 상태로 반영하세요.
4. stop hook, queue 운영 규칙, 검증 경로 같은 runtime 문서가 어긋나지 않는지 확인하세요.
