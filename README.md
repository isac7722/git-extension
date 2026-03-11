# ge — Git Extension CLI

[![CI](https://github.com/isac7722/ge-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/isac7722/ge-cli/actions/workflows/ci.yml)
[![Release](https://github.com/isac7722/ge-cli/actions/workflows/release.yml/badge.svg)](https://github.com/isac7722/ge-cli/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/isac7722/ge-cli)](https://goreportcard.com/report/github.com/isac7722/ge-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A lightweight CLI that extends Git with **multi-account management**, **enhanced worktree support**, and **branch cleanup** — with seamless git passthrough.

## Features

- **Multi-account management** — Switch between Git identities (name, email, SSH key) globally or per-repo with an interactive TUI selector
- **Enhanced worktrees** — Create, list, and remove worktrees with auto branch creation and status indicators
- **Branch cleanup** — Interactively remove stale branches (gone, merged, local-only) with a multi-select TUI
- **Git passthrough** — Any unknown command is forwarded to git (`ge commit` = `git commit`)

## Installation

### Homebrew (recommended)

```bash
brew install isac7722/ge/ge
```

### curl

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/isac7722/ge-cli/main/install.sh)
```

### Build from source

```bash
git clone https://github.com/isac7722/ge-cli.git
cd ge-cli
go build -o ge ./cmd/ge
sudo mv ge /usr/local/bin/
```

### Shell integration

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
eval "$(ge init zsh)"    # for zsh
eval "$(ge init bash)"   # for bash
```

Shell integration enables commands that modify the current shell environment (e.g., `GIT_SSH_COMMAND` for SSH key switching).

## Quick Start

```bash
# 1. Add your first account
ge user add

# 2. Add another account
ge user add

# 3. Switch between accounts (interactive TUI)
ge user

# 4. Use git as usual
ge status
ge commit -m "hello from ge"
```

## Usage

### User Management — `ge user`

```bash
ge user                    # interactive TUI selector
ge user list               # list all accounts
ge user current            # show current git identity
ge user add                # interactively add a new account
ge user set <profile>      # set account for current repo (--local)
ge user switch <profile>   # switch global account
ge user ssh-key <profile>  # view or update SSH key path
ge user remove <profile>   # remove an account
ge user migrate            # migrate from legacy gituser format
```

Running `ge user` with no subcommand opens an interactive selector powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea).

### Worktree Management — `ge worktree` / `ge wt`

```bash
ge worktree add <branch> [dir]   # create worktree (auto-creates branch if needed)
ge worktree list                 # list worktrees with status indicators
ge worktree remove [path]        # remove worktree (interactive if no path given)
```

`ge wt` is an alias for `ge worktree`.

### Branch Cleanup — `ge clean`

```bash
ge clean                   # interactive multi-select of stale branches
ge clean --gone            # only branches whose remote is gone
ge clean --merged          # only merged branches
ge clean --local           # only local-only branches
ge clean --dry-run         # preview without deleting
ge clean --force           # skip confirmation, delete all
```

### Git Passthrough

Any command not recognized by `ge` is forwarded to `git`:

```bash
ge status                  # → git status
ge commit -m "msg"         # → git commit -m "msg"
ge push origin main        # → git push origin main
ge log --oneline -10       # → git log --oneline -10
```

## Architecture

`ge` uses a **hybrid shell-function + binary** approach:

- **Shell function** (loaded via `eval "$(ge init <shell>)"`): Intercepts commands that need to modify the current shell environment — specifically `GIT_SSH_COMMAND` export for account switching.
- **Go binary**: Handles everything else — user management, worktree operations, branch cleanup, and git passthrough.

This is the same pattern used by tools like `rbenv`, `pyenv`, and `direnv`.

## Configuration

Accounts are stored in `~/.ge/credentials` using INI format:

```ini
[work]
name = John Doe
email = john@company.com
ssh_key = ~/.ssh/work_ed25519

[personal]
name = John Doe
email = john@gmail.com
ssh_key = ~/.ssh/personal_ed25519
```

You can edit this file directly or use `ge user add` to create entries interactively.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, guidelines, and how to submit pull requests.

## License

[MIT](LICENSE)
