# Contributing to bolt-monitor

## Commit Conventions

This project uses [Conventional Commits](https://www.conventionalcommits.org/). All commit messages must follow this format:

```
<type>(<scope>): <description>

[optional body]
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code (formatting)
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `build`: Changes that affect the build system or external dependencies
- `ci`: Changes to CI configuration and scripts
- `chore`: Other changes that don't modify src or test files
- `revert`: Reverts a previous commit

### Examples

```
feat(monitor-api): add endpoint to list probe locations
fix(dashboard): correct timezone display in incident list
docs(readme): update deployment instructions
ci: add golangci-lint configuration
```

## Pull Request Workflow

1. Create a feature branch from `main`
2. Make your changes with commits following conventional commits
3. Open a PR to `main`
4. Run the relevant local validation commands (lint, typecheck, tests)
5. After review, merge to `main`

## Code Quality

### Go Services

```bash
# Run linter
golangci-lint run ./services/... ./shared/...

# Run tests
go test ./services/... ./shared/...
```

### Dashboard

```bash
cd apps/dashboard

# Install dependencies
pnpm install --frozen-lockfile

# Lint
pnpm run lint

# Typecheck
pnpm exec tsc --noEmit

# Check formatting
pnpm run format:check
```

### Infrastructure

```bash
cd infra

# Typecheck
pnpm run check

# Check formatting
pnpm run format:check
```

## Local Development

See [README.md](./README.md#quick-start) for local development setup.
