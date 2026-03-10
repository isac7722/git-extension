# ============================================================
# ge worktree — Enhanced git worktree management
# ============================================================
# Requires: _ge_core.sh (sourced before this file)
# ============================================================

_ge_worktree_add() {
  local branch="$1"
  local dir="${2:-./$branch}"

  if [[ -z "$branch" ]]; then
    echo "$(_ge_red '✗') Usage: ge worktree add <branch> [directory]"
    return 1
  fi

  if ! git rev-parse --git-dir &>/dev/null; then
    echo "$(_ge_red '✗') Current directory is not a git repository."
    return 1
  fi

  if git rev-parse --verify "$branch" 2>/dev/null; then
    git worktree add "$dir" "$branch"
  else
    git worktree add -b "$branch" "$dir"
  fi

  local status=$?
  if [[ $status -eq 0 ]]; then
    echo ""
    echo "$(_ge_green '✔') Worktree created"
    printf "  %-10s %s\n" "Branch:" "$(_ge_bold "$branch")"
    printf "  %-10s %s\n" "Path:"   "$(_ge_dim "$dir")"
    echo ""
  fi
  return $status
}

_ge_worktree_help() {
  echo ""
  echo "$(_ge_bold 'Usage:') ge worktree <subcommand>"
  echo ""
  echo "$(_ge_bold 'Subcommands:')"
  printf "  %-28s %s\n" "add <branch> [dir]" "Create worktree (auto-creates branch if needed)"
  printf "  %-28s %s\n" "list"               "List active worktrees (git worktree list)"
  printf "  %-28s %s\n" "remove <path>"      "Remove a worktree (git worktree remove)"
  printf "  %-28s %s\n" "prune"              "Clean up stale worktree info"
  echo ""
  echo "  Any other subcommand is passed directly to 'git worktree'."
  echo ""
}

_ge_worktree_dispatch() {
  case "$1" in
    add)            shift; _ge_worktree_add "$@" ;;
    help|-h|--help) _ge_worktree_help ;;
    "")             _ge_worktree_help ;;
    *)              git worktree "$@" ;;
  esac
}
