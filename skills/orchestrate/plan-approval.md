---
name: orchestrate/plan-approval
description: Flow 0/A/B plan approval protocol, clarification checkpoints, developer-to-tester handoff enrichment, and convention injection for Small tasks. Load when orchestrator is about to invoke developer for a Medium+ task.
---

# Plan Approval

**Load when:** about to invoke developer for a Medium+ task (Flow 0/A/B decision, handoff chain prep).

---

## Clarification checkpoints (MANDATORY)

Before launching certain agents, the orchestrator MUST ask the user questions. DO NOT assume — ask first.

### Before Architect (if task touches DB or schema)

Ask: (1) "What existing tables are related?" (2) "Extend existing table or create new?" (3) "Constraints or relationships to consider?"

**Why:** prevents Architect from designing a new table when ALTER TABLE would suffice.

### Before Developer

**For Medium+ tasks**, check for an existing handoff note per `/handoff` skill (Read operation). If found, pass it inline to the developer — this is a continuation.

**If no handoff exists**, ask the user:
1. "Do you already have progress on this feature? What files already exist?"
2. "Is there partial code or a branch with prior work?"

**Why:** The handoff prevents the Developer from wasting tokens re-reading PRD, design, and code already processed. If the user confirms prior work (and there's no handoff), be specific: "Only X, Y, Z are missing — don't read the rest."

**Skip handoff check for Small tasks (1-5 pts).**

---

## Plan approval flows (CRITICAL)

**The USER approves plans, not the orchestrator.** Hard rule. Three flows, chosen by whether an architect already ran.

### Flow 0 — Architect ran, reuse §8 checklist (PREFERRED for Complex tasks)

When the architect just produced `design.md` with a `§8 Implementation checklist` (or equivalent "pasos secuenciales"), **do not re-synthesize** the plan. The checklist IS the plan at file-level granularity.

1. Orchestrator already has `design.md` in context from the architect call
2. Present the `§8 checklist` **verbatim** to the user + a 2-line summary of key decisions D1-Dn
3. User approves explicitly (`sí`, `dale`, `apruebo`)
4. Invoke developer with `plan_preapproved=true` and the design.md inlined (see Design inline rule below)

**Forbidden in Flow 0:** rewriting the checklist, adding files the architect didn't list, removing files without user approval, asking the developer to "come up with the file list". The DASH-FEAT-008 retrospective showed Flow 0 absence caused duplicated tokens and minor drift.

### Flow A — User-dictated plan, no architect (shortcut)

Valid ONLY when the orchestrator has: read context itself, designed a concrete plan with file list + patterns + decisions, presented it to the user in the main conversation, and received EXPLICIT approval (`sí`, `dale`, `apruebo`) — not a generic "continue".

Then invoke developer once with `plan_preapproved=true` + the full plan inline + instructions to create `.handoff/<TASK-ID>.md` as progress artifact, proceed directly, update handoff during work, fill `## Handoff for tester` before finishing.

**Flow A legitimacy test:** did the USER type the file list in chat, or did the ORCHESTRATOR synthesize it and the user only approved the strategy ("go with option B")? Strategic approval = Flow B, not Flow A. See `anti-patterns.md` #6.

### Flow B — Developer designs the plan (no architect, no user plan)

1. Invoke developer with: `"MANDATORY FIRST STEP: Create .handoff/<TASK-ID>.md with execution plan. Then STOP and return plan summary — do NOT write production code. Do NOT present directly to user."`
2. Developer returns plan summary
3. **Orchestrator surfaces the plan to the user and WAITS for explicit approval.** Forbidden phrases: "el plan coincide, apruebo", "sigo adelante" — user decides, not orchestrator
4. Loop:
   - `dale` / `ok` / `aprobado` → invoke developer with `plan_preapproved=true` + approved plan inline
   - User asks changes → new plan → surface again
   - User rejects → restart scope

**Never:** auto-approve, interpret silence as approval, interpret generic "sigue" as approval, skip the surface step.

### Design inline rule (Flow 0 + tester handoffs)

When invoking the developer after Flow 0, **inline the `design.md` content in the prompt** instead of only passing the path, when ALL: `design.md` ≤500 lines, orchestrator already has it in context, developer needs the full design.

This saves a full `Read` call (~3-5k tokens) and guarantees the developer sees the same version the user approved. Pass it as:

```
Design (pre-approved by user, from <docs>/03-tasks/<TASK-ID>/design.md):
<inline content>

DO NOT re-read the design file — you have the complete content above.
```

For files >500 lines: pass the path + tell the developer which sections are load-bearing.

**Same rule for tester handoffs:** inline the `## Handoff for tester` section if the orchestrator has it in context, instead of asking the tester to re-read the handoff file.

---

## Handoff path rule (CRITICAL)

Handoffs → `.handoff/` in project root. Docs → `<docs>/` vault. Never mix. See `vault-setup.md` path map.

---

## Developer → Tester handoff enrichment (MANDATORY)

Before the orchestrator invokes the tester, it MUST verify that the developer filled the `## Handoff for tester` section of `.handoff/<TASK-ID>.md`. This section exists precisely so the tester does not re-read production files.

**Verification checklist (before invoking tester):**

1. [ ] `.handoff/<TASK-ID>.md` has a non-empty `## Handoff for tester` section
2. [ ] "Public interfaces / contracts" has the exact signatures of new/modified functions, types, DTOs
3. [ ] "Edge cases descubiertos" is filled (not just "N/A" — if there truly are none, the developer should say "sin edge cases no triviales")
4. [ ] "Tests requeridos — por stack" has tests grouped by stack (`#### Tests Go`, `#### Tests React/TS`, etc.) — each group with file path, run command, and numbered list. **Single flat list is NOT accepted for cross-stack tasks.**
5. [ ] "Validación ya ejecutada" lists the commands the developer ran (go build, go vet, npm run build)
6. [ ] `## Output entregado` table is filled with build/lint/test results
7. [ ] `## Puente de contratos` is filled (cross-stack tasks only) — both "Backend expone" and "Frontend consume" have exact types
8. [ ] `## Dependencias cross-service` is filled (cross-service tasks only)

**If any check fails:** re-invoke the developer with the specific gap: "Fill [missing section] in `.handoff/<TASK-ID>.md`. Do NOT touch production code." This is cheaper than letting the tester re-read the codebase.

**After QA passes (before archive):** verify the developer filled `## Retro` → "Qué funcionó" and "Qué no funcionó". The orchestrator fills "Métricas" with actual invocation counts and QA bounces.

**Tester prompt template (after verification passes):**

```
Stack(s): <go|react|flutter|...>. Skill: <convention-skill>.

PRIMARY INPUT: Read `.handoff/<TASK-ID>.md` — specifically the `## Handoff for tester` section. That section contains:
- files the developer touched (with their role)
- exact signatures of new interfaces/DTOs
- patterns applied
- edge cases discovered
- build tags / constraints
- **tests requeridos — por stack** — tests grouped by stack (#### Tests Go, #### Tests React/TS, etc.), each with file path, run command, and numbered list. Work one stack at a time.
- validation already run (do NOT repeat build checks)

For cross-stack tasks, also check `## Puente de contratos` — it shows the exact contract between backend and frontend. If your tests touch the boundary, verify both sides match.

Do NOT re-read the production files unless the handoff is missing a specific detail you need. If the handoff is incomplete, STOP and report to the orchestrator.

Your job: implement ONLY the tests listed in each stack group of "Tests requeridos — por stack". Do NOT add extra tests beyond these lists.
```

Developer boundary: never writes test files. If tester finds dev-authored tests, report violation (see tester.md).

---

## Convention injection for Small tasks

For Small tasks (1-5 pts), do NOT tell developer to load the full convention skill. Instead, read its essential rules and inject inline in the prompt:

- **Go:** `go-conventions/rules/coding.md` + `rules/architecture.md`
- **React / Flutter / Astro:** read `<stack>-conventions` essential rules, include inline

**Context injection rule:** if user provided context in conversation (screenshots, files, decisions), pass it inline — do NOT tell the agent "read file X". See global instructions for full protocol.
