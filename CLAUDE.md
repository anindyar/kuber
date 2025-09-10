# Claude Code Context: Kubernetes TUI Manager

## Project Overview
Building an intuitive terminal-based Kubernetes cluster manager similar to k9s/Rancher but with improved UX inspired by lazydocker. Focus on resource browsing, editing, shell access, logs, and metrics.

## Technology Stack
- **Language**: Go 1.21+ (native Kubernetes ecosystem)
- **TUI Framework**: Bubble Tea v1.0 (modern functional architecture)
- **Kubernetes Client**: client-go (official, most mature)
- **Styling**: lipgloss (terminal styling)
- **Testing**: go test + teatest for TUI components

## Architecture
Library-first approach with 4 core libraries:
1. **kubernetes-client**: K8s API communication and auth
2. **tui-components**: Reusable terminal UI widgets
3. **resource-manager**: Resource discovery, caching, real-time updates
4. **metrics-collector**: Performance data collection/aggregation

## Key Implementation Points
- Functional architecture with Bubble Tea's Elm-inspired patterns
- Real-time updates via Kubernetes watch API
- In-memory caching with TTL for performance
- Structured JSON logging for observability
- TDD workflow with contract-first development

## Current Phase
Implementation planning complete. Ready for task breakdown and development.

## Recent Changes
- Technology research completed (Go + Bubble Tea chosen)
- Data model designed with 6 core entities
- API contracts defined for all 4 libraries
- Quickstart guide with test scenarios created
- Performance targets: <100ms navigation, <50MB memory

## Development Guidelines
- RED-GREEN-Refactor TDD cycle mandatory
- Each library must have CLI with --help/--version/--format
- Real Kubernetes clusters for integration tests
- No implementation before failing tests
- Constitutional principles followed (simplicity, observability, versioning)