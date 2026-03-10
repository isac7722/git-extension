# ============================================================
# ge clean — Remove stale local branches
# ============================================================
# Requires: _ge_core.sh (sourced before this file)
# ============================================================

_ge_clean_is_protected() {
  case "$1" in
    main|master|develop|dev|staging|production|release) return 0 ;;
    *) return 1 ;;
  esac
}

_ge_clean_run() {
  local mode="$1" force="$2" dry_run="$3"

  # Verify git repo
  if ! git rev-parse --git-dir &>/dev/null; then
    echo "$(_ge_red '✗') Current directory is not a git repository."
    return 1
  fi

  # Sync with remote
  echo "$(_ge_dim 'Fetching and pruning remote refs...')"
  if ! git fetch --prune 2>&1; then
    echo "$(_ge_red '✗') git fetch --prune failed. Check your network connection."
    return 1
  fi

  # Detect current branch (handle detached HEAD)
  local current_branch=""
  current_branch="$(git symbolic-ref --short HEAD 2>/dev/null)" || current_branch=""

  # Collect branches to delete
  local -a targets=()
  local -a skipped=()

  if [[ "$mode" == "merged" ]]; then
    # Detect default branch: origin/HEAD → main → master
    local default_branch=""
    default_branch="$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's|refs/remotes/origin/||')"
    if [[ -z "$default_branch" ]]; then
      if git rev-parse --verify origin/main &>/dev/null; then
        default_branch="main"
      elif git rev-parse --verify origin/master &>/dev/null; then
        default_branch="master"
      else
        echo "$(_ge_red '✗') Could not detect default branch."
        return 1
      fi
    fi

    while IFS= read -r branch; do
      branch="${branch#"${branch%%[![:space:]]*}"}"  # trim leading spaces
      branch="${branch#\* }"  # remove current branch marker
      [[ -z "$branch" ]] && continue

      if _ge_clean_is_protected "$branch"; then
        skipped+=("$branch (protected)")
        continue
      fi
      if [[ "$branch" == "$current_branch" ]]; then
        skipped+=("$branch (current)")
        continue
      fi
      targets+=("$branch")
    done < <(git branch --merged "$default_branch" 2>/dev/null)
  else
    # gone mode: branches whose upstream is gone
    while IFS= read -r line; do
      local branch="${line%% *}"
      local track="${line#* }"
      [[ "$track" != "[gone]" ]] && continue

      if _ge_clean_is_protected "$branch"; then
        skipped+=("$branch (protected)")
        continue
      fi
      if [[ "$branch" == "$current_branch" ]]; then
        skipped+=("$branch (current)")
        continue
      fi
      targets+=("$branch")
    done < <(git for-each-ref --format='%(refname:short) %(upstream:track)' refs/heads/)
  fi

  # Show skipped branches
  for s in "${skipped[@]}"; do
    echo "  $(_ge_yellow '⏭') Skipping: $s"
  done

  # Nothing to do
  if [[ ${#targets[@]} -eq 0 ]]; then
    echo "$(_ge_green '✔') Already clean — no stale branches found."
    return 0
  fi

  # Display targets
  echo ""
  echo "  Branches to remove:"
  for t in "${targets[@]}"; do
    echo "    $(_ge_red '•') $t"
  done
  echo ""

  # Dry run — stop here
  if [[ "$dry_run" == "true" ]]; then
    echo "$(_ge_dim "[dry-run] ${#targets[@]} branch(es) would be removed.")"
    return 0
  fi

  # Confirmation prompt
  if [[ "$force" != "true" ]]; then
    printf "  Remove %d branch(es)? [y/N] " "${#targets[@]}"
    local answer=""
    read -r answer
    case "$answer" in
      y|Y|yes|YES) ;;
      *) echo "  Cancelled."; return 0 ;;
    esac
  fi

  # Delete branches
  local count=0
  local delete_flag="-D"
  [[ "$mode" == "merged" ]] && delete_flag="-d"

  for t in "${targets[@]}"; do
    if git branch "$delete_flag" "$t" &>/dev/null; then
      count=$((count + 1))
    else
      echo "  $(_ge_red '✗') Failed to delete: $t"
    fi
  done

  echo "$(_ge_green '✔') Removed $count branch(es)."
}

_ge_clean_help() {
  echo ""
  echo "$(_ge_bold 'Usage:') ge clean [options]"
  echo ""
  echo "$(_ge_bold 'Options:')"
  printf "  %-20s %s\n" "(default)"      "Remove branches whose remote tracking is gone"
  printf "  %-20s %s\n" "--merged"        "Remove branches already merged into default branch"
  printf "  %-20s %s\n" "-f, --force"     "Skip confirmation prompt"
  printf "  %-20s %s\n" "--dry-run"       "Preview only, no branches deleted"
  printf "  %-20s %s\n" "help, -h"        "Show this help"
  echo ""
  echo "$(_ge_bold 'Protected branches') (never deleted):"
  echo "  main, master, develop, dev, staging, production, release"
  echo ""
}

_ge_clean_dispatch() {
  local mode="gone" force="false" dry_run="false"

  while [[ $# -gt 0 ]]; do
    case "$1" in
      help|-h|--help)  _ge_clean_help; return 0 ;;
      --merged)        mode="merged"; shift ;;
      -f|--force)      force="true"; shift ;;
      --dry-run)       dry_run="true"; shift ;;
      *)
        echo "$(_ge_red '✗') Unknown option: $1"
        _ge_clean_help
        return 1
        ;;
    esac
  done

  _ge_clean_run "$mode" "$force" "$dry_run"
}
