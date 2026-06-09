# .backlog AI Agent Guide

## 1. Directory Layout

```
<project-root>/
├── .backlog/
│   ├── project.md        # Project config: project_id, external_projects map
│   ├── AGENTS.md         # This file — backlog system instructions for AI agents
│   ├── backlog.md        # Dashboard — all tasks grouped by status
│   ├── tasks/            # Work items: TASK, BUG, STORY, SPIKE, CHORE
│   ├── epics/            # Epics only
│   └── sprints/          # Sprint definitions
├── ...
```

## 2. Project Configuration

Each project has a `.backlog/project.md` file defining its identity and external dependencies.

```yaml
---
project_id: "FE"              # Short prefix used in task IDs
name: "Frontend Application"
external_projects:
  BE: "../backend/.backlog"   # Maps project prefix → relative path to that project's .backlog/
  API: "../api/.backlog"
---
```

### Resolution rules
- The task ID prefix (e.g. `BE` in `BE-API-003`) identifies the source project.
- AI extracts the prefix, looks up `external_projects` in `project.md`, resolves the relative path.
- If prefix matches this project's `project_id`, search `tasks/` locally.
- If prefix is not in `external_projects`, treat as error — task references unknown project.

### Access boundary for AI
- **Current project**: Full read/write access to ALL files — source code, config, docs, `.backlog/`, tests. This is where implementation happens.
- **External projects** (listed in `external_projects`): Read-only access, restricted to `.backlog/` only. Never read external source code, config, or any non-backlog file. Never write to external projects.

## 3. File Naming Convention

```
<PROJECT>-<TYPE>-<NUM>.md
```

| Part | Example | Meaning |
|------|---------|---------|
| PROJECT | `FE` | Frontend, Backend `BE`, API `API`, Mobile `MB` |
| TYPE | `TASK`, `BUG`, `STORY`, `SPIKE`, `CHORE`, `EPIC` | Agile work type |
| NUM | `001` | Sequential, minimum 3 digits |

Examples:
- `FE-TASK-001.md` — frontend development task
- `BE-BUG-002.md` — backend bug fix
- `FE-STORY-003.md` — frontend user story
- `BE-SPIKE-001.md` — backend research spike
- `MB-CHORE-001.md` — mobile maintenance

Files reside in `tasks/` for TASK/BUG/STORY/SPIKE/CHORE, and `epics/` for EPIC.

## 4. Frontmatter Fields

All fields use YAML frontmatter between `---` delimiters.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | yes | Unique ID. Format: `<PROJECT>-<TYPE>-<NUM>`. Must match filename. |
| `type` | enum | yes | `task` | `bug` | `story` | `spike` | `chore` |
| `status` | enum | yes | `todo` | `in-progress` | `review` | `on-hold` | `done` | `cancelled`. Default: `todo`. |
| `priority` | enum | yes | `critical` | `high` | `medium` | `low`. Default: `medium`. |
| `severity` | enum | null | `blocker` | `critical` | `major` | `minor` | `trivial`. Bug only. |
| `assignee` | string | null | GitHub username of person working this. |
| `reporter` | string | null | GitHub username of person who created this. |
| `labels` | string[] | yes | Tags. Examples: `tech-debt`, `retro-item`, `frontend`, `backend`. |
| `components` | string[] | yes | Affected area. Examples: `auth`, `api`, `ui`, `database`. |
| `story_points` | number | null | Estimation. Must be Fibonacci: 1, 2, 3, 5, 8, 13. |
| `epic` | string | null | Parent epic ID this belongs to. |
| `sprint` | string | null | Sprint name or ID this is assigned to. |
| `fix_version` | string | null | Release version this targets. |
| `resolution` | enum | null | `done` | `wontfix` | `duplicate` | `cannot-reproduce`. Only when `status` is `done` or `cancelled`. |
| `parent` | string | null | Parent item ID. epic→story, story→task, task→sub-task. Use `null` for top-level items. |
| `children` | string[] | yes | IDs of child items. Inverse of `parent`. |
| `depends_on` | string[] | yes | IDs this task blocks on. Cannot start until those are done. |
| `blocks` | string[] | yes | IDs that depend on this. Inverse of `depends_on`. |
| `related_to` | string[] | yes | IDs of related tasks. Purely informational — no status blocking or cascade. |
| `due_date` | string | null | Deadline. Format: `YYYY-MM-DD`. |
| `created` | string | yes | Creation date. Format: `YYYY-MM-DD`. |
| `updated` | string | yes | Last modification date. Format: `YYYY-MM-DD`. |

## 5. Task Creation Template

Copy this when creating a new task file. Replace placeholders in `<< >>`.

File: `.backlog/tasks/<<PROJECT>>-<<TYPE>>-<<NUM>>.md`

```markdown
---
id: "<<PROJECT>>-<<TYPE>>-<<NUM>>"
type: "<<TYPE>>"
status: "todo"
priority: "medium"
severity: null
assignee: null
reporter: null
labels: []
components: []
story_points: null
epic: null
sprint: null
fix_version: null
resolution: null
parent: null
children: []
depends_on: []
blocks: []
related_to: []
due_date: null
created: "<<YYYY-MM-DD>>"
updated: "<<YYYY-MM-DD>>"
---

## Summary

<<One-line title>>

## Description

<<What needs to be done and why>>

## Acceptance Criteria

- [ ] <<Condition 1>>
- [ ] <<Condition 2>>
- [ ] <<Condition 3>>

## Comments

<<Optional: context, decisions, links>>
```

**AI agent instructions when creating:**
- Increment NUM from the highest existing file of same PROJECT-TYPE.
- Fill all `<< >>` placeholders before saving.
- If `type=bug`, add `## Reproducer` and `## Environment` sections.
- If `type=story`, ensure at least 3 acceptance criteria.
- Update `backlog.md` dashboard immediately after creation.
- Commit with message: `feat(backlog): create <<ID>> - <<summary>>`.

## 6. Type-Specific Rules

### story
- Requires `## Acceptance Criteria` section with minimum 3 checkboxes.
- Format: `- [ ] <concrete, testable condition>`.
- AC must be verifiable without ambiguity.
- A story is not done unless all AC checkboxes are ticked.

### bug
- Requires `## Reproducer` section with numbered steps to reproduce.
- Requires `## Environment` section with OS, browser, version fields.
- Must have `severity` set (not null).
- Invalid until minimum viable reproducer is documented.
- Bugs always take implicit priority over equivalent new feature work.

### spike
- Purpose: research, exploration, proof-of-concept, decision.
- Not done until it produces at least one follow-up:
  - A concrete `task` or `story` for implementation.
  - A decision record or ADR link.
  - A documented "no action needed" conclusion.
- Spikes that produce no output are waste.

### epic
- Lives in `epics/` directory, not `tasks/`.
- Represents a large initiative spanning multiple sprints.
- Children are stories (not tasks directly).

### chore
- Maintenance, CI/CD, tooling, infrastructure, refactoring.
- No user-facing change expected.
- No acceptance criteria required unless complex.

## 7. Status Lifecycle

```
                    ┌──> cancelled
                    │
todo ──> in-progress ──> review ──> done
  │          │            │
  │          │            │
  └── on-hold <───────────┘
         │  │
         │  └──> cancelled
         │
         └──> todo | in-progress | review
```

- **todo**: Ready to pick up. Frontmatter should be complete.
- **in-progress**: Actively being worked. Only this person works it.
- **review**: Implementation done. Needs verification (review, QA, AC check).
- **done**: All AC met, reviewed, merged. No further action.
- **on-hold**: Paused. Reason must be documented in `## Comments`. Common reasons: blocker, deprioritized, waiting on external dependency, awaiting decision.
- **cancelled**: Abandoned. `resolution` must be set.

Transitions in:
- `todo → in-progress`: Only if all `depends_on` are `done`.
- `todo → on-hold`: Deprioritized, not ready, or blocked externally.
- `in-progress → review`: Only if `## Acceptance Criteria` met (for stories) or equivalent quality bar.
- `in-progress → on-hold`: Work paused — blocker encountered, waiting on dependency or external input.
- `review → done`: Only after successful review. Set `resolution: done`.
- `review → in-progress`: If review fails or changes requested.
- `review → on-hold`: Review paused — waiting on author or external input.
- `on-hold → todo`: Unblocked, needs fresh start.
- `on-hold → in-progress`: Unblocked, resume directly.
- `on-hold → review`: Issue resolved while held in review — resume verification.
- `on-hold → cancelled`: Abandoned while held.
- `todo | in-progress | review | on-hold → cancelled`: Set `resolution` to `wontfix`, `duplicate`, or `cannot-reproduce`.
- `done` and `cancelled` are terminal — no transitions out.

## 8. Dependency Semantics

### depends_on
- Task X depends on Y: X cannot start until Y is `done`.
- AI must not move X to `in-progress` if any dependency is not `done`.
- If Y is `cancelled` or permanently blocked, X needs re-evaluation (split, reprioritize, or cancel).

### blocks
- Task X blocks Y: Y waits on X.
- When X reaches `done`, check all items in X.blocks — they can now proceed.
- When X is `cancelled`, all items in X.blocks must be notified and re-evaluated.

### Bidirectional consistency
- Both fields must be kept in sync:
  - If `depends_on: [FE-TASK-003]` then FE-TASK-003 must have `blocks: [this-task-ID]`.
  - If `blocks: [FE-TASK-004]` then FE-TASK-004 must have `depends_on: [this-task-ID]`.
- AI validates this on every update.

### No circular dependencies
- `depends_on` / `blocks` must never form a cycle.
- AI validates acyclic graph on every write.

### Cross-project resolution
- Dependencies reference external tasks by full ID: `depends_on: ["BE-API-003"]`.
- AI extracts the project prefix (`BE`), looks up `external_projects` in `project.md`, resolves relative path to external `.backlog/tasks/BE-API-003.md`.
- Reads only that file's frontmatter to check `status` — never reads anything outside a `.backlog/` directory.
- If external project's `.backlog/project.md` is missing, treat as error.
- When creating an external dependency, AI must also update the external project's `blocks` field to maintain bidirectional consistency.

### Related tasks
- `related_to` lists task IDs that are conceptually linked — same feature, complementary work, bug + hotfix.
- No status implications: a task can move to `done` regardless of its `related_to` items' status.
- Bidirectional convention: if `FE-TASK-001` has `related_to: ["BE-API-003"]`, then `BE-API-003` should include `FE-TASK-001` in its `related_to`.
- Cross-project references in `related_to` resolve the same way as `depends_on` via `project.md`.

## 9. Parent/Child Hierarchy

```
epic (level 3)
  └── story (level 2)
        └── task (level 1)
              └── sub-task (level 0)
```

### Rules
- `parent` points to the immediate parent ID. `null` for top-level items.
- `children` lists immediate child IDs.
- **epic** → children are stories only.
- **story** → children are tasks.
- **task** → children are sub-tasks (smaller units, optional).
- A **story** must have at least one child `task` to be estimable.
- No cross-level parenting (epic → task not allowed).

### Status cascading (suggestions, not automations)
- If parent is `done`, all children should be `done`. Flag inconsistency.
- If all children are `done`, parent can move to `done` or `review`.
- If a child is `cancelled`, parent should be re-evaluated (maybe split or descoped).

## 10. Process Rules

### Definition of Done (universal)
Before any item moves to `done`, ALL must be true:
- Code merged to target branch.
- All acceptance criteria met.
- Peer review completed.
- Tests pass (unit + integration).
- No regression introduced.
- Documentation updated if applicable.

### WIP Limit
- Maximum 1-3 items `in-progress` per person at any time.
- Prevents context switching overhead.
- AI should warn if WIP limit is exceeded.

### No Orphan Tasks
- Every item must have a `parent` ID or explicitly be standalone (`parent: null`).
- Prevents floating work with no context.

### Sprint Scope Frozen
- Once a sprint starts, no new additions without removing equal effort.
- Protects team commitment and velocity.

## 11. Quality Rules

### Bug Before Feature
- Bugs always take implicit priority over equivalent-scope new feature work.
- When choosing between a bug and a new feature at same priority level, do the bug first.

### Minimal Reproducer Required
- Bug tasks are invalid until `## Reproducer` has concrete steps.
- Steps must be reproducible by someone other than the reporter.
- AI should reject bugs without reproducer.

### Task Size Limit
- Maximum 13 story points.
- If estimate > 13, split into sub-tasks or smaller stories.
- A task spanning more than 1 sprint is a code smell.

### No Circular Dependencies
- `depends_on` / `blocks` must form a DAG (directed acyclic graph).
- AI validates this on every update.
- Simple checker: if traversing `depends_on` leads to the starting node, cycle exists.

## 12. Estimation Rules

### Fibonacci Scale
- Story points must be from: 1, 2, 3, 5, 8, 13.
- No fractional, negative, or non-Fibonacci values.
- 0 = trivial (use 1 instead). >13 = must split.
- Points represent relative effort, complexity, and uncertainty combined.

### Spikes Produce Output
- A spike is done only when it produces concrete output:
  - At least one follow-up `task` or `story` with estimate, OR
  - A documented decision with rationale, OR
  - Explicit conclusion that no action is needed.
- Spikes without output are waste — flag for review.

## 13. Maintenance Rules

### 5% Tech Debt Per Sprint
- Each sprint must allocate minimum 5% of capacity to items with label `tech-debt`.
- Protects codebase health from gradual decay.
- Track via `labels: ["tech-debt"]` on CHORE or TASK items.

### Retro Items Become Tasks
- Every retrospective improvement item gets a task created within 48 hours.
- Label: `labels: ["retro-item"]`.
- Prevents retro insights from being forgotten.

### Stale Task Rule
- If a task stays `todo` for 3+ consecutive sprints, re-evaluate:
  - Split into smaller pieces.
  - Reprioritize (up or down).
  - Close as `cancelled` with `resolution: wontfix`.
- Stale tasks signal backlog bloat.

## 14. Backlog Health

### Top 10 Completion
- Top 10 items by priority must have complete frontmatter before sprint planning.
- No nulls allowed in: `assignee`, `priority`, `story_points`, `epic`.
- Incomplete frontmatter means the item is not ready for planning.

### Dependency Freshness
- When a task changes status, check its `blocks` list — downstream dependents may proceed or need re-evaluation.
- When a task in `depends_on` of another changes to `cancelled`, the dependent must be re-evaluated.
- AI should flag stale dependencies on every status change.

## 15. Hierarchy Validation

### Epic Completeness
- An epic is not `done` unless all descendant stories are `done`.
- AI should flag epic as ready for review when all children are done.

### Story Completeness
- A story must have at least one child task to be estimable.
- A story with no child tasks cannot have `story_points` set.

### Cross-Project Dependencies
- Dependencies reference external tasks by full ID: `depends_on: ["BE-API-003"]`.
- Project prefix (`BE`) maps via `project.md` → `external_projects["BE"]` → relative path.
- AI resolves the external task file and reads only its frontmatter.
- AI never reads code, config, or any file outside `.backlog/` directories of any project.
- If `project.md` does not list the prefix in `external_projects`, flag as unknown project.
- External dependency `status` changes should trigger re-evaluation of local items that list them in `depends_on`.

## 16. Git Integration

### Branch Naming
- Format: `feature/<task-id>-<kebab-case-summary>`.
- Example: `feature/FE-TASK-001-add-login-page`.
- Branches off `develop` (per AGENTS.md git flow rules).

### Commits
- Format: `type(scope): [TASK-ID] message`.
- Example: `feat(auth): [FE-TASK-001] implement login form`.
- Subject line ≤72 characters.

### Pull Requests
- Title references task ID.
- Description lists which dependencies/blocks are resolved.
- Body includes link to backlog item.

## 17. AI Agent Rules

- Read `backlog.md` first for a task overview before drilling into individual files.
- Parse frontmatter with a YAML parser — never regex-guess field values.
- Before editing any task file, read the full file.
- Validate consistency across related tasks on every update (dependencies, parent/children, blocks).
- When updating a task, also update the dashboard `backlog.md`.
- Never create a task without filling all `<< >>` placeholders.
- Never use emojis in frontmatter, summaries, commit messages, or any backlog content.
- When blocked by missing information (e.g., incomplete frontmatter), ask the user — do not guess.
- On task creation, increment NUM from the highest existing ID of same PROJECT-TYPE.
- All dates in `YYYY-MM-DD` format.
- **Current project access**: Full read/write access to ALL files — source code, config, `.backlog/`, tests, docs. This is where you implement changes and write code.
- **External project access**: Read-only, restricted to `.backlog/` only. When resolving cross-project dependencies, read only:
  1. External project's `.backlog/project.md` (to confirm `project_id`).
  2. The specific task file `.backlog/tasks/<ID>.md` (to check `status`).
  Never read source code, config, or any non-backlog file from external projects. Never write to external projects.
- **Cross-project resolution**: Use `project.md` + task ID prefix to locate external task files. Never search by path guesswork.
