# Contributing to Savannah Backend API

Thank you for considering contributing to the Savannah Backend API! This document provides guidelines and information for contributors.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)

## 🤝 Code of Conduct

This project adheres to a code of conduct that ensures a welcoming environment for everyone. By participating, you agree to:

- Be respectful and inclusive in all interactions
- Focus on constructive feedback and collaboration
- Respect differing viewpoints and experiences
- Show empathy towards other community members

## 🚀 Getting Started

### Prerequisites

- **Go 1.21+** - Latest stable version recommended
- **PostgreSQL 15+** - For database development
- **Redis 7+** - For job queue functionality
- **Docker** - For containerized development (optional)
- **Git** - Version control

### Development Environment Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/savannah-backend.git
   cd savannah-backend
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   ```

3. **Set up Environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Start Dependencies**
   ```bash
   docker-compose up -d postgres redis
   ```

5. **Run Migrations**
   ```bash
   go run cmd/migrate.go -action=up
   ```

6. **Start Development Server**
   ```bash
   go run main.go
   ```

## 🔄 Development Workflow

### Branch Naming Convention

Use descriptive branch names with prefixes:

- `feature/` - New features
- `bugfix/` - Bug fixes
- `hotfix/` - Critical production fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring

Examples:
- `feature/add-user-authentication`
- `bugfix/fix-sms-queue-timeout`
- `docs/update-api-documentation`

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code formatting changes
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Build process or auxiliary tool changes

**Examples:**
```
feat(auth): add OIDC authentication middleware
fix(sms): resolve timeout issue in job processing
docs(api): update customer endpoint documentation
```

## 📝 Coding Standards

### Go Code Standards

1. **Follow Go conventions**
   - Use `gofmt` for code formatting
   - Use `golint` for style checking
   - Use `go vet` for static analysis

2. **Package Organization**
   ```
   internal/
   ├── api/        # HTTP handlers and routes
   ├── service/    # Business logic
   ├── repository/ # Data access layer
   ├── domain/     # Business entities
   └── config/     # Configuration
   ```

3. **Error Handling**
   - Always handle errors explicitly
   - Use custom error types for business logic errors
   - Include context in error messages

4. **Documentation**
   - Add Go doc comments for all public functions and types
   - Use meaningful variable and function names
   - Include usage examples in documentation

### Code Quality Tools

Run these tools before submitting:

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run security checks (if gosec is installed)
gosec ./...

# Run linter (if golangci-lint is installed)
golangci-lint run
```

### API Design Standards

1. **RESTful Principles**
   - Use appropriate HTTP methods (GET, POST, PUT, DELETE)
   - Use proper HTTP status codes
   - Follow consistent URL patterns

2. **Response Format**
   ```json
   {
     "success": true,
     "message": "Operation completed successfully",
     "data": { ... },
     "error": null
   }
   ```

3. **Validation**
   - Validate all input data
   - Return meaningful validation error messages
   - Use appropriate HTTP status codes for validation errors

## 🧪 Testing Guidelines

### Test Structure

1. **Unit Tests**
   - Test individual functions and methods
   - Mock external dependencies
   - Aim for 70%+ code coverage

2. **Integration Tests**
   - Test API endpoints end-to-end
   - Use test database for integration tests
   - Test error scenarios and edge cases

3. **Test Naming**
   ```go
   func TestServiceName_MethodName_ExpectedBehavior(t *testing.T) {
       // Test implementation
   }
   ```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./internal/service/...

# Run tests with verbose output
go test -v ./...
```

### Test Requirements

- All new features must include tests
- Bug fixes should include regression tests
- Tests should be fast and reliable
- Use table-driven tests for multiple scenarios

## 🔍 Pull Request Process

### Before Submitting

1. **Run Quality Checks**
   ```bash
   # Run tests
   go test ./...
   
   # Check formatting
   go fmt ./...
   
   # Run vet
   go vet ./...
   ```

2. **Update Documentation**
   - Update API documentation if endpoints changed
   - Update README.md if necessary
   - Add/update Go doc comments

3. **Test Your Changes**
   - Ensure all tests pass
   - Test manually in development environment
   - Verify no breaking changes

### Pull Request Template

When creating a PR, include:

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that causes existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] Added tests for new functionality
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project coding standards
- [ ] Self-review of code completed
- [ ] Documentation updated
- [ ] No unnecessary dependencies added
```

### Review Process

1. **Automated Checks**
   - All CI/CD checks must pass
   - Code coverage requirements must be met
   - Security scans must pass

2. **Code Review**
   - At least one reviewer approval required
   - Address all reviewer comments
   - Update PR if requested

3. **Merge Requirements**
   - All checks passing
   - Up-to-date with main branch
   - No merge conflicts

## 🐛 Issue Reporting

### Bug Reports

Include the following information:

```markdown
## Bug Description
Clear description of the bug

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: [e.g., Ubuntu 20.04]
- Go Version: [e.g., 1.21.0]
- Project Version: [e.g., commit hash]

## Additional Context
Any other relevant information
```

### Feature Requests

Use this template:

```markdown
## Feature Description
Clear description of the proposed feature

## Use Case
Why is this feature needed?

## Proposed Solution
How should this feature work?

## Alternatives Considered
Any alternative solutions considered

## Additional Context
Any other relevant information
```

## 📚 Resources

### Documentation
- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [Redis Go Client](https://redis.uptrace.dev/)

### Tools
- [Go Style Guide](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [golangci-lint](https://golangci-lint.run/)
- [gosec](https://github.com/securecodewarrior/gosec)

## ❓ Getting Help

- **GitHub Issues** - For bug reports and feature requests
- **GitHub Discussions** - For questions and general discussion
- **Code Comments** - For implementation-specific questions

## 🙏 Recognition

Contributors will be recognized in:
- PROJECT_SHOWCASE.md acknowledgments
- GitHub contributor statistics
- Release notes for significant contributions

Thank you for contributing to Savannah Backend API! 🚀