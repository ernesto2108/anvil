---
name: code-review-rubric
description: Scoring rubric and report format for code reviews. Defines evaluation criteria, scoring scale, and output structure. Used by the QA agent and anyone reviewing code quality.
---

# Code Review Rubric

## Evaluation Criteria

### Correctness
- Logic bugs, edge cases, nil/null safety
- Error handling (wrapped errors, no bare returns)
- Contract compliance (matches PRD/design)

### Performance
- Unnecessary allocations, blocking I/O
- N+1 queries, inefficient algorithms
- Missing pagination, unbounded queries

### Code Quality
- Naming clarity, readability
- Cyclomatic complexity, duplication
- Single responsibility, minimal abstractions

### Testing
- Unit tests present for business logic
- Critical paths covered
- Edge cases and error paths tested
- Deterministic (no flaky, no time-dependent)

### Concurrency Safety
- Race conditions, shared state
- Proper use of mutexes/channels/errgroups
- Context propagation and cancellation

### Security
- Input validation at boundaries
- SQL injection, XSS, command injection
- Auth/authorization checks present
- Secrets not hardcoded

## Scoring Scale

| Score | Meaning | Action |
|---|---|---|
| 9-10 | Excellent — production ready | Approve |
| 7-8 | Good — minor improvements only | Approve with suggestions |
| 5-6 | Needs work — significant issues | Block, create tasks |
| 3-4 | Major problems — rethink approach | Block, escalate to architect |
| 1-2 | Critical — security/data risk | Block immediately |

**Threshold: score < 7 → BLOCK and create backlog tasks.**

## Report Format

Write to: `<docs>/03-tasks/<TASK-ID>/qa-review.md`

```markdown
# QA Review — <TASK-ID>

## Score: X/10

## Summary
One paragraph: what was reviewed, overall assessment.

## Strengths
- What was done well (acknowledge good work)

## Issues
| # | Severity | Category | File | Description |
|---|---|---|---|---|
| 1 | critical | correctness | internal/user/service.go:45 | Missing nil check on... |
| 2 | high | testing | — | No tests for error path in... |
| 3 | medium | quality | internal/order/handler.go:12 | Function too complex (cyclomatic 15) |

## Improvements
- Actionable suggestions (not vague "improve this")

## Risk Level
low / medium / high — based on blast radius of issues found
```

## Backlog Task Format

When issues are found, append to: `<docs>/02-backlog/sprint-current.md`

Each task must include:
- Title (imperative: "Fix nil check in user service")
- Type (bug / tech-debt / test-gap)
- Description (what's wrong and why it matters)
- Severity (critical / high / medium / low)
- Suggested fix (concrete, not vague)
- Affected files (with line numbers)
