# Contributing to ge-cli

Thanks for your interest in contributing! This guide will help you get started.

## Development Environment

### Prerequisites

- [Go](https://go.dev/dl/) (version specified in `go.mod`)
- Git
- macOS or Linux

### Setup

```bash
git clone https://github.com/isac7722/ge-cli.git
cd ge-cli
go mod download
```

### Build

```bash
go build -o ge ./cmd/ge
```

### Test

```bash
go test ./...
```

### Lint

We use [golangci-lint](https://golangci-lint.run/):

```bash
golangci-lint run
```

CI runs both `go test ./...` and `golangci-lint` on every pull request.

## Project Structure

```
ge-cli/
├── cmd/ge/              # Entry point (main.go)
├── internal/
│   ├── cmd/             # Cobra command definitions
│   │   ├── user/        # ge user subcommands
│   │   ├── worktree/    # ge worktree subcommands
│   │   └── clean/       # ge clean command
│   ├── config/          # Credentials file parsing (~/.ge/credentials)
│   ├── git/             # Git operations (exec wrappers)
│   ├── migration/       # Legacy format migration
│   ├── shell/           # Shell init and wrapper scripts
│   └── tui/             # Interactive TUI components (Bubble Tea)
├── scripts/
│   ├── dev-install.sh   # Local development installer
│   └── uninstall_legacy.sh  # Legacy uninstaller
├── ge-cli-sh/           # Legacy shell implementation (deprecated)
├── install.sh           # curl installer
└── .goreleaser.yml      # Release configuration
```

## Development Workflow

1. **Fork** the repository
2. **Create a branch** from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```
3. **Make your changes** and ensure tests pass:
   ```bash
   go test ./...
   go build ./cmd/ge/
   ```
4. **Commit** using [Conventional Commits](#commit-conventions)
5. **Push** and open a Pull Request against `main`

## Commit Conventions

This project uses [Conventional Commits](https://www.conventionalcommits.org/) and [release-please](https://github.com/googleapis/release-please) for automated releases. Your commit messages determine the next version number.

| Prefix | Purpose | Version bump |
|--------|---------|-------------|
| `feat:` | New feature | Minor |
| `fix:` | Bug fix | Patch |
| `docs:` | Documentation only | — |
| `chore:` | Maintenance (deps, CI, etc.) | — |
| `refactor:` | Code restructuring | — |
| `test:` | Adding or updating tests | — |

**Breaking changes**: Add `!` after the type (e.g., `feat!: remove legacy support`) or include `BREAKING CHANGE:` in the commit body.

Examples:
```
feat: add ge user remove command
fix: handle missing SSH key gracefully
docs: update installation instructions
chore: bump golangci-lint to v1.60
```

## Pull Request Guidelines

- **One concern per PR** — keep changes focused
- **CI must pass** — tests and lint are required
- **Describe what and why** — the PR description should explain the motivation
- **Update docs if needed** — if your change affects user-facing behavior, update README.md

## Code Style

- Run `go fmt ./...` before committing
- Follow standard Go conventions ([Effective Go](https://go.dev/doc/effective_go))
- Keep packages focused — `internal/` is intentionally organized by domain
- TUI components use [Bubble Tea](https://github.com/charmbracelet/bubbletea) / [Lip Gloss](https://github.com/charmbracelet/lipgloss)

## Reporting Issues

### Bugs

When filing a bug report, please include:

- `ge version` output
- OS and shell (`echo $SHELL`)
- Steps to reproduce
- Expected vs actual behavior

### Feature Requests

Open an issue describing:

- The problem you're trying to solve
- Your proposed solution (if any)
- Any alternatives you've considered

## Questions?

Open an issue with the **question** label — happy to help!
