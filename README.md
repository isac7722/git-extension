# ge — Git Extension CLI

A lightweight CLI that extends git with multi-account management, enhanced worktree support, and seamless git passthrough.

## Features

- **`ge user`** — Switch between multiple git accounts (global/per-repo), manage SSH keys, directory auto-switch rules
- **`ge worktree`** — Enhanced worktree management with auto branch creation
- **Git passthrough** — Any unknown command is forwarded to git (`ge commit` → `git commit`)

## Installation

### npm (recommended)

```bash
npm install -g ge-cli

# Add to your shell RC file (~/.zshrc or ~/.bashrc):
eval "$(ge init zsh)"    # for zsh
eval "$(ge init bash)"   # for bash
```

### curl

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/isac7722/ge-cli/main/install.sh)
```

### Manual

```bash
git clone https://github.com/isac7722/ge-cli.git ~/.ge
~/.ge/install.sh
```

## Setup

After installation, register your git accounts:

```bash
# Edit the accounts file directly:
$EDITOR ~/.config/gituser/accounts

# Or use the interactive command:
ge user add
```

Account file format:
```
aliases:name:email:ssh_key_path
work,w:John Work:john@company.com:~/.ssh/work_ed25519
personal,p:John Personal:john@gmail.com:~/.ssh/personal_ed25519
```

## Usage

### User Management (`ge user`)

```bash
ge user                    # fzf selector (or help if fzf not installed)
ge user list               # list all accounts
ge user current            # show current global account
ge user <alias>            # switch globally
ge user set <alias>        # switch for current repo only (--local)
ge user add                # interactively register an account
ge user ssh-key <a> [path] # view or update SSH key path
ge user clone <a> <url>    # clone with a specific account
ge user rule add <a> <dir> # auto-switch rule for a directory
ge user rule list          # list rules
ge user rule remove <a>    # remove a rule
```

### Worktree Management (`ge worktree`)

```bash
ge worktree add <branch> [dir]  # create worktree (auto-creates branch if needed)
ge worktree list                # list worktrees (git passthrough)
ge worktree remove <path>       # remove worktree (git passthrough)
```

### Git Passthrough

```bash
ge status                  # → git status
ge commit -m "msg"         # → git commit -m "msg"
ge push origin main        # → git push origin main
```

### Other

```bash
ge update                  # self-update (git pull or npm update hint)
ge version                 # show version
ge help                    # show help
```

## Architecture

`ge` uses a hybrid shell-function + executable approach:

- **Shell function** (`shell/ge.zsh` / `shell/ge.bash`): Loaded via `eval "$(ge init <shell>)"`. Handles commands that modify the current shell environment (e.g., `GIT_SSH_COMMAND` export for account switching).
- **Standalone binary** (`bin/ge`): Handles `ge init`, `ge worktree`, and git passthrough. Works without shell integration.

This is the same pattern used by `rbenv`, `pyenv`, and `direnv`.

## Migrating from gituser

`ge` is backward compatible. The `gituser` command is aliased to `ge user` automatically. All config files remain in `~/.config/gituser/`.

| Before | After |
|--------|-------|
| `gituser list` | `ge user list` |
| `gituser <alias>` | `ge user <alias>` |
| `gituser set <alias>` | `ge user set <alias>` |
| `gituser clone <a> <url>` | `ge user clone <a> <url>` |
| `gituser update` | `ge update` |

## Requirements

- git
- bash 4+ or zsh
- fzf (optional, for interactive account selector)
- macOS or Linux

## License

MIT
