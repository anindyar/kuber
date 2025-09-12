# Contributing to kUber

We welcome contributions to kUber! This document provides guidelines for contributing to the project.

## Getting Started

### Prerequisites
- Go 1.24 or later
- kubectl configured with cluster access
- Git
- Make

### Development Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/your-username/kuber.git
   cd kuber
   ```

3. Set up development tools:
   ```bash
   make dev-setup
   ```

4. Build and test:
   ```bash
   make build
   make test
   ```

## Development Workflow

### 1. Create a Branch
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number
```

### 2. Make Changes
- Follow Go best practices and the existing code style
- Write tests for new functionality
- Update documentation as needed
- Keep commits focused and atomic

### 3. Test Your Changes
```bash
# Run all tests
make test

# Run linting
make lint

# Format code
make format

# Build to ensure it compiles
make build
```

### 4. Commit Guidelines
We follow conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```bash
git commit -m "feat(logs): add real-time log streaming support"
git commit -m "fix(ui): resolve navigation issue in resource view"
git commit -m "docs: update installation instructions"
```

### 5. Push and Create Pull Request
```bash
git push origin your-branch-name
```

Then create a Pull Request through the GitHub interface.

## Code Style

### Go Style
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` and `goimports` (run `make format`)
- Follow the existing naming conventions
- Write clear, descriptive function and variable names

### File Organization
- Keep files focused on single concerns
- Group related functionality together
- Use clear directory structure following the existing pattern

### Comments and Documentation
- Document all public functions and types
- Include examples for complex functionality
- Update README.md for user-facing changes
- Add inline comments for complex logic

## Testing

### Test Types
1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **Contract Tests**: Verify API contracts between libraries

### Writing Tests
- Use table-driven tests where appropriate
- Test both happy path and error cases
- Mock external dependencies
- Keep tests fast and reliable

### Running Tests
```bash
# All tests
make test

# Specific test types
make test-unit
make test-integration
make test-contract
```

## Pull Request Process

### Before Submitting
- [ ] Tests pass locally
- [ ] Code is formatted (`make format`)
- [ ] Linting passes (`make lint`)
- [ ] Build succeeds (`make build`)
- [ ] Documentation is updated
- [ ] Commit messages follow conventions

### PR Requirements
1. **Title**: Clear, descriptive title following conventional commit format
2. **Description**: 
   - What changes were made and why
   - Any breaking changes
   - Screenshots for UI changes
   - Testing instructions
3. **Size**: Keep PRs focused and reasonably sized
4. **Tests**: Include appropriate test coverage

### Review Process
1. Automated checks must pass
2. At least one maintainer review required
3. All conversations must be resolved
4. No merge conflicts

## Issue Reporting

### Bug Reports
Include:
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, kubectl version)
- Screenshots if applicable
- Logs or error messages

### Feature Requests
Include:
- Use case and problem being solved
- Proposed solution
- Alternative approaches considered
- Mockups or examples if helpful

## Documentation

### Types of Documentation
- **README.md**: User-facing getting started guide
- **CONTRIBUTING.md**: This file - contributor guidelines
- **Code Comments**: Inline documentation
- **API Documentation**: Generated from code comments

### Documentation Guidelines
- Write for your audience (users vs developers)
- Include examples and code snippets
- Keep content up to date with code changes
- Use clear, concise language

## Release Process

Releases are handled by maintainers:

1. Version bump in appropriate files
2. Update CHANGELOG.md
3. Create and push git tag
4. GitHub Actions builds and publishes release
5. Update documentation as needed

## Community Guidelines

### Code of Conduct
- Be respectful and inclusive
- Focus on constructive feedback
- Help newcomers feel welcome
- Assume positive intent

### Communication
- Use GitHub issues for bug reports and feature requests
- Use GitHub discussions for questions and general discussion
- Be patient with responses - maintainers are volunteers

## Getting Help

- **Documentation**: Check README.md and existing issues
- **Discussions**: Use GitHub Discussions for questions
- **Issues**: Create an issue for bugs or feature requests
- **Discord/Slack**: [Add community links if available]

## Recognition

Contributors are recognized in:
- CONTRIBUTORS.md file (maintained automatically)
- Release notes for significant contributions
- Special recognition for major features or fixes

Thank you for contributing to kUber! 🚀