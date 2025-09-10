# Implementation Plan: Intuitive Kubernetes TUI Manager

**Branch**: `001-we-will-be` | **Date**: 2025-09-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-we-will-be/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
4. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
5. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, or `GEMINI.md` for Gemini CLI).
6. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
7. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
8. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
Build an intuitive terminal-based Kubernetes cluster manager with features similar to SUSE Rancher or k9s, but with improved visual design inspired by lazydocker. The system must provide resource browsing, configuration editing, pod shell access, log viewing, and performance metrics monitoring through a unified TUI interface.

## Technical Context
**Language/Version**: Go 1.21+ (native Kubernetes ecosystem language)  
**Primary Dependencies**: Bubble Tea v1.0 (modern TUI framework), client-go (official Kubernetes client), lipgloss (styling)  
**Storage**: Local config files for settings and cluster connections, in-memory caching for resource data  
**Testing**: go test with teatest for TUI components, standard HTTP mocking for Kubernetes API  
**Target Platform**: Linux terminals (with potential future support for macOS/Windows)
**Project Type**: single (terminal application)  
**Performance Goals**: <100ms response time for navigation, real-time log streaming, smooth scrolling for large resource lists  
**Constraints**: Low memory footprint (<50MB idle), work with standard terminal capabilities, respect Kubernetes RBAC  
**Scale/Scope**: Support multiple clusters, thousands of resources per cluster, concurrent log streams

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Simplicity**:
- Projects: 1 (main TUI application)
- Using framework directly? Yes (will use TUI framework without custom wrappers)
- Single data model? Yes (unified Kubernetes resource representation)
- Avoiding patterns? Yes (direct API calls, no unnecessary abstractions)

**Architecture**:
- EVERY feature as library? Yes (kubernetes-client, tui-components, resource-managers, metrics-collector)
- Libraries listed: 
  - kubernetes-client: K8s API communication and authentication
  - tui-components: Reusable terminal UI widgets and layouts
  - resource-manager: Resource discovery, caching, and real-time updates
  - metrics-collector: Performance data collection and aggregation
- CLI per library: Each library will expose debug/test commands with --help/--version/--format
- Library docs: llms.txt format planned for each library

**Testing (NON-NEGOTIABLE)**:
- RED-GREEN-Refactor cycle enforced? Yes (tests written first, must fail before implementation)
- Git commits show tests before implementation? Yes (strict TDD workflow)
- Order: Contract→Integration→E2E→Unit strictly followed? Yes
- Real dependencies used? Yes (actual Kubernetes clusters for integration tests)
- Integration tests for: new libraries, Kubernetes API changes, cross-library communication
- FORBIDDEN: Implementation before test, skipping RED phase

**Observability**:
- Structured logging included? Yes (JSON logs with context)
- Frontend logs → backend? N/A (single process, unified logging)
- Error context sufficient? Yes (stack traces, user actions, cluster state)

**Versioning**:
- Version number assigned? 0.1.0 (MAJOR.MINOR.BUILD)
- BUILD increments on every change? Yes
- Breaking changes handled? Yes (migration scripts, backward compatibility testing)

## Project Structure

### Documentation (this feature)
```
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
# Option 1: Single project (DEFAULT)
src/
├── models/
├── services/
├── cli/
└── lib/

tests/
├── contract/
├── integration/
└── unit/

# Option 2: Web application (when "frontend" + "backend" detected)
backend/
├── src/
│   ├── models/
│   ├── services/
│   └── api/
└── tests/

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure]
```

**Structure Decision**: Option 1 (Single project) - Terminal application with library-based architecture

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:
   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `/scripts/update-agent-context.sh [claude|gemini|copilot]` for your AI assistant
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- Each library contract → contract test task [P] (4 libraries = 4 tasks)
- Each entity → model creation task [P] (6 entities = 6 tasks)
- Each user story from quickstart → integration test task (5 scenarios = 5 tasks)
- TUI component tasks based on components contract (5 components = 5 tasks)
- Implementation tasks to make tests pass (libraries + main app = 8 tasks)
- CLI commands for each library (4 libraries × 3 commands = 12 tasks)

**Ordering Strategy**:
- **Phase 1** (Parallel): Contract tests for all 4 libraries
- **Phase 2** (Parallel): Data model implementations (6 entities)
- **Phase 3** (Sequential): Core library implementations (dependency order)
  1. kubernetes-client (foundation)
  2. resource-manager (depends on kubernetes-client)
  3. metrics-collector (depends on kubernetes-client)
  4. tui-components (independent)
- **Phase 4** (Sequential): Integration tests (requires libraries)
- **Phase 5** (Sequential): Main application assembly
- **Phase 6** (Parallel): CLI commands and documentation

**Dependency Mapping**:
- kubernetes-client: No dependencies (foundation)
- resource-manager: Requires kubernetes-client
- metrics-collector: Requires kubernetes-client
- tui-components: Independent (pure UI logic)
- Main app: Requires all libraries

**Estimated Output**: 40-45 numbered, ordered tasks in tasks.md with clear [P] markings for parallel execution

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [ ] Complexity deviations documented

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*