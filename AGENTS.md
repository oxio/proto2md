# Agent Instructions

## Git Workflow
- Always stage (`git add`) any files that you create, modify, or delete.

## Versions, tags, releases

- For requests about versions, tags, or releases, use the skill at `.agents/skills/versioning/SKILL.md`.

## Build & Task Runner
- Use `Taskfile` (`Taskfile.yml` / `task`) rather than `Makefile` (`make`).

## Verification
- Never report any task as finished before running all checks and tests (`task check` / `task test`) and verifying that they pass.
