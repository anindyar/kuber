# Tasks: Intuitive Kubernetes TUI Manager

**Input**: Design documents from `/specs/001-we-will-be/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → If not found: ERROR "No implementation plan found"
   → Extract: tech stack, libraries, structure
2. Load optional design documents:
   → data-model.md: Extract entities → model tasks
   → contracts/: Each file → contract test task
   → research.md: Extract decisions → setup tasks
3. Generate tasks by category:
   → Setup: project init, dependencies, linting
   → Tests: contract tests, integration tests
   → Core: models, services, CLI commands
   → Integration: DB, middleware, logging
   → Polish: unit tests, performance, docs
4. Apply task rules:
   → Different files = mark [P] for parallel
   → Same file = sequential (no [P])
   → Tests before implementation (TDD)
5. Number tasks sequentially (T001, T002...)
6. Generate dependency graph
7. Create parallel execution examples
8. Validate task completeness:
   → All contracts have tests?
   → All entities have models?
   → All endpoints implemented?
9. Return: SUCCESS (tasks ready for execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Single project**: `src/`, `tests/` at repository root
- Paths assume single project structure per implementation plan

## Phase 3.1: Setup
- [ ] T001 Create project structure with src/ and tests/ directories per implementation plan
- [ ] T002 Initialize Go module with go.mod and required dependencies (Bubble Tea v1.0, client-go, lipgloss)
- [ ] T003 [P] Configure golangci-lint for code quality and gofmt for formatting
- [ ] T004 [P] Create Makefile with build, test, lint, and install targets
- [ ] T005 [P] Setup .gitignore for Go projects with binary and IDE exclusions

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**

### Contract Tests (4 libraries)
- [ ] T006 [P] Contract test for kubernetes-client library in tests/contract/test_kubernetes_client.go
- [ ] T007 [P] Contract test for tui-components library in tests/contract/test_tui_components.go  
- [ ] T008 [P] Contract test for resource-manager library in tests/contract/test_resource_manager.go
- [ ] T009 [P] Contract test for metrics-collector library in tests/contract/test_metrics_collector.go

### Integration Tests (5 quickstart scenarios)
- [ ] T010 [P] Integration test basic connectivity scenario in tests/integration/test_connectivity.go
- [ ] T011 [P] Integration test resource browsing scenario in tests/integration/test_browsing.go
- [ ] T012 [P] Integration test log viewing scenario in tests/integration/test_logs.go
- [ ] T013 [P] Integration test resource editing scenario in tests/integration/test_editing.go
- [ ] T014 [P] Integration test metrics display scenario in tests/integration/test_metrics.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Data Models (7 entities)
- [ ] T015 [P] Cluster model in src/models/cluster.go
- [ ] T016 [P] Resource model in src/models/resource.go
- [ ] T017 [P] Namespace model in src/models/namespace.go
- [ ] T018 [P] LogEntry model in src/models/log_entry.go
- [ ] T019 [P] MetricDataPoint model in src/models/metric.go
- [ ] T020 [P] UserSession model in src/models/session.go
- [ ] T021 [P] NavigationContext model in src/models/navigation.go

### Library Implementations (Sequential by dependency)
- [ ] T022 kubernetes-client library core implementation in src/lib/kubernetes/client.go
- [ ] T023 kubernetes-client authentication handling in src/lib/kubernetes/auth.go
- [ ] T024 kubernetes-client resource operations in src/lib/kubernetes/resources.go
- [ ] T025 tui-components table component in src/lib/tui/table.go
- [ ] T026 tui-components log viewer component in src/lib/tui/logs.go
- [ ] T027 tui-components metrics chart component in src/lib/tui/charts.go
- [ ] T028 tui-components navigation component in src/lib/tui/navigation.go
- [ ] T029 tui-components resource editor component in src/lib/tui/editor.go
- [ ] T030 resource-manager cache implementation in src/lib/resource/cache.go
- [ ] T031 resource-manager discovery service in src/lib/resource/discovery.go
- [ ] T032 resource-manager watch/update service in src/lib/resource/watcher.go
- [ ] T033 metrics-collector data collection in src/lib/metrics/collector.go
- [ ] T034 metrics-collector aggregation service in src/lib/metrics/aggregator.go
- [ ] T035 metrics-collector alert system in src/lib/metrics/alerts.go

### CLI Commands (4 libraries × 3 commands each)
- [ ] T036 [P] kubernetes-client CLI with --help, --version, --format in src/cli/k8s_client.go
- [ ] T037 [P] tui-components CLI with --help, --version, --format in src/cli/tui_components.go
- [ ] T038 [P] resource-manager CLI with --help, --version, --format in src/cli/resource_manager.go
- [ ] T039 [P] metrics-collector CLI with --help, --version, --format in src/cli/metrics_collector.go

### Main Application
- [ ] T040 Main TUI application structure in cmd/kuber/main.go
- [ ] T041 Application state management in src/app/state.go
- [ ] T042 Event handling and routing in src/app/events.go
- [ ] T043 View management and navigation in src/app/views.go

## Phase 3.4: Integration
- [ ] T044 Connect resource-manager to kubernetes-client in src/services/integration.go
- [ ] T045 Connect metrics-collector to kubernetes-client for data source
- [ ] T046 Integrate all TUI components into main application views
- [ ] T047 Configuration file handling for cluster connections in src/config/config.go
- [ ] T048 Structured logging setup with JSON format in src/logging/logger.go
- [ ] T049 Error handling and recovery mechanisms in src/errors/handler.go

## Phase 3.5: Polish
- [ ] T050 [P] Unit tests for all models in tests/unit/models_test.go
- [ ] T051 [P] Unit tests for utilities and helpers in tests/unit/utils_test.go
- [ ] T052 [P] Performance tests ensuring <100ms navigation response in tests/performance/response_test.go
- [ ] T053 [P] Memory usage tests ensuring <50MB idle consumption in tests/performance/memory_test.go
- [ ] T054 [P] Create library documentation in llms.txt format for each library
- [ ] T055 [P] Update README.md with installation and usage instructions
- [ ] T056 [P] Create example configurations and cluster profiles
- [ ] T057 Code cleanup and refactoring for maintainability
- [ ] T058 Manual testing using quickstart.md scenarios
- [ ] T059 Build optimization and release preparation

## Dependencies
### Critical Path
- Setup (T001-T005) before everything
- Contract tests (T006-T009) before any library implementation
- Integration tests (T010-T014) before any main application code
- Models (T015-T021) before services that use them
- kubernetes-client (T022-T024) before resource-manager and metrics-collector
- Libraries (T022-T035) before main application (T040-T043)
- Core implementation before integration (T044-T049)
- Everything before polish (T050-T059)

### Specific Dependencies
- T030-T032 (resource-manager) require T022-T024 (kubernetes-client)
- T033-T035 (metrics-collector) require T022-T024 (kubernetes-client)
- T040-T043 (main app) require T025-T029 (tui-components)
- T044-T046 (integration) require all libraries complete
- T058 (manual testing) requires complete application

## Parallel Execution Examples

### Phase 3.2: All Contract and Integration Tests
```bash
# Launch T006-T014 together (all test files are independent):
Task: "Contract test for kubernetes-client library in tests/contract/test_kubernetes_client.go"
Task: "Contract test for tui-components library in tests/contract/test_tui_components.go"
Task: "Contract test for resource-manager library in tests/contract/test_resource_manager.go"
Task: "Contract test for metrics-collector library in tests/contract/test_metrics_collector.go"
Task: "Integration test basic connectivity in tests/integration/test_connectivity.go"
Task: "Integration test resource browsing in tests/integration/test_browsing.go"
Task: "Integration test log viewing in tests/integration/test_logs.go"
Task: "Integration test resource editing in tests/integration/test_editing.go"
Task: "Integration test metrics display in tests/integration/test_metrics.go"
```

### Phase 3.3: Data Models
```bash
# Launch T015-T021 together (all model files are independent):
Task: "Cluster model in src/models/cluster.go"
Task: "Resource model in src/models/resource.go" 
Task: "Namespace model in src/models/namespace.go"
Task: "LogEntry model in src/models/log_entry.go"
Task: "MetricDataPoint model in src/models/metric.go"
Task: "UserSession model in src/models/session.go"
Task: "NavigationContext model in src/models/navigation.go"
```

### Phase 3.3: CLI Commands
```bash
# Launch T036-T039 together (all CLI files are independent):
Task: "kubernetes-client CLI with --help, --version, --format in src/cli/k8s_client.go"
Task: "tui-components CLI with --help, --version, --format in src/cli/tui_components.go"
Task: "resource-manager CLI with --help, --version, --format in src/cli/resource_manager.go"
Task: "metrics-collector CLI with --help, --version, --format in src/cli/metrics_collector.go"
```

### Phase 3.5: Documentation and Testing
```bash
# Launch T050-T056 together (independent documentation and test files):
Task: "Unit tests for all models in tests/unit/models_test.go"
Task: "Unit tests for utilities in tests/unit/utils_test.go"
Task: "Performance tests <100ms navigation in tests/performance/response_test.go"
Task: "Memory usage tests <50MB idle in tests/performance/memory_test.go"
Task: "Create library documentation in llms.txt format for each library"
Task: "Update README.md with installation and usage instructions"
Task: "Create example configurations and cluster profiles"
```

## Notes
- [P] tasks = different files, no shared dependencies
- Verify all tests fail before implementing (TDD requirement)
- Commit after each task completion
- Use teatest framework for TUI component testing
- Real Kubernetes clusters required for integration tests
- Follow constitutional principles: simplicity, observability, versioning

## Task Generation Rules
*Applied during main() execution*

1. **From Contracts**:
   - 4 contract files → 4 contract test tasks [P] (T006-T009)
   - Each contract defines library interface → implementation tasks (T022-T035)
   
2. **From Data Model**:
   - 7 entities → 7 model creation tasks [P] (T015-T021)
   - Entity relationships → service layer integration (T044-T046)
   
3. **From Quickstart Scenarios**:
   - 5 user scenarios → 5 integration tests [P] (T010-T014)
   - Test scenarios → manual validation task (T058)

4. **Library Architecture**:
   - 4 libraries → 4 CLI command tasks [P] (T036-T039)
   - Library dependencies → sequential implementation order

## Validation Checklist
*GATE: Checked by main() before returning*

- [x] All 4 contracts have corresponding test tasks (T006-T009)
- [x] All 7 entities have model tasks (T015-T021)
- [x] All 9 tests come before implementation (T006-T014 before T015+)
- [x] Parallel tasks are truly independent (different files, no dependencies)
- [x] Each task specifies exact file path
- [x] No [P] task modifies same file as another [P] task
- [x] TDD workflow enforced (tests must fail before implementation)
- [x] Constitutional principles followed (library-first, CLI per library, structured logging)