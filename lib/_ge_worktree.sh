# ============================================================
# ge worktree — Enhanced git worktree management
# ============================================================
# Requires: _ge_core.sh (sourced before this file)
# ============================================================

_ge_worktree_branch_selector() {
  # zsh: make arrays 0-indexed like bash (scoped to this function only)
  [ -n "$ZSH_VERSION" ] && setopt localoptions KSH_ARRAYS

  # Must be run from a terminal
  [[ -t 0 ]] || return 1

  # Collect branches already checked out in worktrees
  local -A wt_branches
  local line wt_path wt_hash wt_branch_raw wt_branch
  while IFS= read -r line; do
    read -r wt_path wt_hash wt_branch_raw <<< "$line"
    wt_branch="${wt_branch_raw#\[}"
    wt_branch="${wt_branch%\]}"
    wt_branches["$wt_branch"]=1
  done < <(git worktree list)

  # Collect available branches
  local _sel_names=() _sel_labels=()
  local ref

  # Local branches
  while IFS= read -r ref; do
    [[ -z "$ref" ]] && continue
    [[ -n "${wt_branches[$ref]+x}" ]] && continue
    _sel_names+=("$ref")
    _sel_labels+=("local")
  done < <(git branch --format='%(refname:short)')

  # Remote branches (skip if already listed or in worktree)
  local short
  while IFS= read -r ref; do
    [[ -z "$ref" ]] && continue
    short="${ref#*/}"
    # Skip if already listed as local or checked out in worktree
    local already=0
    local n
    for n in "${_sel_names[@]}"; do
      [[ "$n" == "$short" ]] && { already=1; break; }
    done
    [[ $already -eq 1 ]] && continue
    [[ -n "${wt_branches[$short]+x}" ]] && continue
    _sel_names+=("$short")
    _sel_labels+=("remote")
  done < <(git branch -r --format='%(refname:short)' | grep -v '/HEAD$')

  local total=${#_sel_names[@]}
  if [[ $total -eq 0 ]]; then
    echo "No branches available." >/dev/tty
    return 1
  fi

  local selected=0

  # Hide cursor and set up cleanup — output to /dev/tty since stdout is captured
  tput civis 2>/dev/null </dev/tty >/dev/tty
  trap 'tput cnorm 2>/dev/null </dev/tty >/dev/tty; trap - INT; return 1' INT

  _ge_worktree_branch_render() {
    if [[ "${1:-}" == "redraw" ]]; then
      local lines_up=$(( 2 + total + 2 ))
      printf "\033[%dA\033[J" "$lines_up" >/dev/tty
    fi

    printf "  %s\n" "$(_ge_bold 'Select a branch')" >/dev/tty
    printf "  %s\n" "$(_ge_dim '──────────────────────────────────────────────')" >/dev/tty

    local idx name label pad_len padding
    for (( idx=0; idx<total; idx++ )); do
      name="${_sel_names[$idx]}"
      label="${_sel_labels[$idx]}"

      pad_len=$(( 30 - ${#name} ))
      (( pad_len < 1 )) && pad_len=1
      padding="$(printf '%*s' "$pad_len" "")"

      if [[ $idx -eq $selected ]]; then
        printf "  %s %s%s%s\n" "$(_ge_cyan '❯')" "$(_ge_bold "$name")" "$padding" "$(_ge_dim "$label")" >/dev/tty
      else
        printf "    %s%s%s\n" "$name" "$padding" "$(_ge_dim "$label")" >/dev/tty
      fi
    done

    printf "  %s\n" "$(_ge_dim '──────────────────────────────────────────────')" >/dev/tty
    printf "  %s\n" "$(_ge_dim '↑↓ navigate  ⏎ select  esc/q cancel')" >/dev/tty
  }

  # Initial render
  echo "" >/dev/tty
  _ge_worktree_branch_render

  # Read input loop
  local _read_char _read_chars
  if [ -n "$ZSH_VERSION" ]; then
    _read_char()  { IFS= read -rsk1 "$1" </dev/tty; }
    _read_chars() { read -rsk2 -t 0.1 "$1" </dev/tty; }
  else
    _read_char()  { IFS= read -rsn1 "$1" </dev/tty; }
    _read_chars() { read -rsn2 -t 0.1 "$1" </dev/tty; }
  fi

  local key
  while true; do
    _read_char key

    case "$key" in
      $'\x1b')
        _read_chars key
        case "$key" in
          '[A') (( selected > 0 )) && (( selected-- )) ;;
          '[B') (( selected < total - 1 )) && (( selected++ )) ;;
          *)  # ESC alone or unknown sequence → cancel
            tput cnorm 2>/dev/null </dev/tty >/dev/tty
            trap - INT
            echo "" >/dev/tty
            return 1
            ;;
        esac
        ;;
      'k') (( selected > 0 )) && (( selected-- )) ;;
      'j') (( selected < total - 1 )) && (( selected++ )) ;;
      ''|$'\n')
        tput cnorm 2>/dev/null </dev/tty >/dev/tty
        trap - INT
        echo "" >/dev/tty
        echo "${_sel_names[$selected]}"
        return 0
        ;;
      'q'|$'\x03'|$'\x1a') # q, Ctrl-C, Ctrl-Z
        tput cnorm 2>/dev/null </dev/tty >/dev/tty
        trap - INT
        echo "" >/dev/tty
        return 1
        ;;
    esac

    _ge_worktree_branch_render "redraw"
  done
}

_ge_worktree_add() {
  local branch="$1"
  local dir="${2:-../$branch}"

  if [[ -z "$branch" ]]; then
    if ! branch="$(_ge_worktree_branch_selector)"; then
      return 0  # user cancelled
    fi
    dir="../$branch"
  fi

  if ! git rev-parse --git-dir &>/dev/null; then
    echo "$(_ge_red '✗') Current directory is not a git repository."
    return 1
  fi

  if git rev-parse --verify "$branch" 2>/dev/null; then
    git worktree add "$dir" "$branch"
  else
    if [[ -e "$dir" ]]; then
      echo "$(_ge_red '✗') Directory '$dir' already exists."
      return 1
    fi
    git worktree add -b "$branch" "$dir"
  fi

  local rc=$?
  if [[ $rc -eq 0 ]]; then
    local abs_dir
    abs_dir="$(cd "$dir" && pwd)"
    cd "$abs_dir"
    echo ""
    echo "$(_ge_green '✔') Worktree created & switched"
    printf "  %-10s %s\n" "Branch:" "$(_ge_bold "$branch")"
    printf "  %-10s %s\n" "Path:"   "$(_ge_dim "$abs_dir")"
    echo ""
  fi
  return $rc
}

_ge_worktree_list() {
  # zsh: make arrays 0-indexed like bash (scoped to this function only)
  [ -n "$ZSH_VERSION" ] && setopt localoptions KSH_ARRAYS

  if ! git rev-parse --git-dir &>/dev/null; then
    echo "$(_ge_red '✗') Current directory is not a git repository."
    return 1
  fi

  local paths=() branches=() hashes=()
  local line wt_path wt_hash wt_branch_raw wt_branch
  while IFS= read -r line; do
    read -r wt_path wt_hash wt_branch_raw <<< "$line"
    # Strip brackets: [main] → main
    wt_branch="${wt_branch_raw#\[}"
    wt_branch="${wt_branch%\]}"
    paths+=("$wt_path")
    hashes+=("$wt_hash")
    branches+=("$wt_branch")
  done < <(git worktree list)

  local total=${#paths[@]}

  if [[ $total -eq 0 ]]; then
    echo ""
    echo "  $(_ge_yellow '⚠') No worktrees found."
    echo ""
    return
  fi

  # Find current worktree index
  local current_path selected=0
  current_path="$(pwd -P 2>/dev/null || pwd)"
  local i
  for (( i=0; i<total; i++ )); do
    if [[ "$current_path" == "${paths[$i]}"* ]]; then
      selected=$i
      break
    fi
  done

  # Hide cursor and set up cleanup
  tput civis 2>/dev/null
  trap 'tput cnorm 2>/dev/null; trap - INT; return' INT

  _ge_worktree_list_render() {
    if [[ "${1:-}" == "redraw" ]]; then
      local lines_up=$(( 2 + total + 2 ))
      printf "\033[%dA\033[J" "$lines_up"
    fi

    printf "  %s\n" "$(_ge_bold 'Git Worktrees')"
    printf "  %s\n" "$(_ge_dim '──────────────────────────────────────────────')"

    local idx branch_display path_display pad_len padding
    for (( idx=0; idx<total; idx++ )); do
      branch_display="${branches[$idx]}"
      path_display="${paths[$idx]/#$HOME/~}"

      pad_len=$(( 24 - ${#branch_display} ))
      (( pad_len < 1 )) && pad_len=1
      padding="$(printf '%*s' "$pad_len" "")"

      if [[ $idx -eq $selected ]]; then
        printf "  %s %s%s%s" "$(_ge_cyan '❯')" "$(_ge_bold "$branch_display")" "$padding" "$(_ge_dim "$path_display")"
      else
        printf "    %s%s%s" "$branch_display" "$padding" "$(_ge_dim "$path_display")"
      fi

      if [[ "$current_path" == "${paths[$idx]}"* ]]; then
        printf "  %s" "$(_ge_green '✔ here')"
      fi
      printf "\n"
    done

    printf "  %s\n" "$(_ge_dim '──────────────────────────────────────────────')"
    printf "  %s\n" "$(_ge_dim '↑↓ navigate  ⏎ select  esc/q cancel')"
  }

  # Initial render
  echo ""
  _ge_worktree_list_render

  # Read input loop
  local _read_char _read_chars
  if [ -n "$ZSH_VERSION" ]; then
    _read_char()  { IFS= read -rsk1 "$1"; }
    _read_chars() { read -rsk2 -t 0.1 "$1"; }
  else
    _read_char()  { IFS= read -rsn1 "$1"; }
    _read_chars() { read -rsn2 -t 0.1 "$1"; }
  fi

  local key
  while true; do
    _read_char key

    case "$key" in
      $'\x1b')
        _read_chars key
        case "$key" in
          '[A') (( selected > 0 )) && (( selected-- )) ;;
          '[B') (( selected < total - 1 )) && (( selected++ )) ;;
          *)  # ESC alone or unknown sequence → cancel
            tput cnorm 2>/dev/null
            trap - INT
            echo ""
            return 0
            ;;
        esac
        ;;
      'k') (( selected > 0 )) && (( selected-- )) ;;
      'j') (( selected < total - 1 )) && (( selected++ )) ;;
      ''|$'\n')
        tput cnorm 2>/dev/null
        trap - INT
        local target="${paths[$selected]}"
        if [[ -d "$target" ]]; then
          cd "$target"
          echo ""
          echo "$(_ge_green '✔') Switched to worktree"
          printf "  %-10s %s\n" "Branch:" "$(_ge_bold "${branches[$selected]}")"
          printf "  %-10s %s\n" "Path:"   "$(_ge_dim "$target")"
          echo ""
        else
          echo ""
          echo "$(_ge_red '✗') Directory not found: $target"
          echo ""
        fi
        return
        ;;
      'q'|$'\x03'|$'\x1a') # q, Ctrl-C, Ctrl-Z
        tput cnorm 2>/dev/null
        trap - INT
        echo ""
        return 0
        ;;
    esac

    _ge_worktree_list_render "redraw"
  done
}

_ge_worktree_help() {
  echo ""
  echo "$(_ge_bold 'Usage:') ge worktree <subcommand>"
  echo ""
  echo "$(_ge_bold 'Subcommands:')"
  printf "  %-28s %s\n" "add [branch] [dir]" "Create worktree (interactive selector if no branch given)"
  printf "  %-28s %s\n" "list"               "Interactive worktree selector"
  printf "  %-28s %s\n" "remove <branch|path>" "Remove a worktree by branch name or path"
  printf "  %-28s %s\n" "prune"              "Clean up stale worktree info"
  echo ""
  echo "  Any other subcommand is passed directly to 'git worktree'."
  echo ""
}

_ge_worktree_remove() {
  local target="$1"
  local force_flag="$2"

  if [[ -z "$target" ]]; then
    echo "$(_ge_red '✗') Usage: ge worktree remove <branch|path> [--force]"
    return 1
  fi

  # Try to resolve branch name to worktree path
  local wt_path=""
  local line wt_hash wt_branch_raw wt_branch
  while IFS= read -r line; do
    read -r wt_path wt_hash wt_branch_raw <<< "$line"
    wt_branch="${wt_branch_raw#\[}"
    wt_branch="${wt_branch%\]}"
    if [[ "$wt_branch" == "$target" ]]; then
      break
    fi
    wt_path=""
  done < <(git worktree list)

  local branch_to_delete=""
  local remove_rc=0

  if [[ -n "$wt_path" ]]; then
    branch_to_delete="$wt_branch"
    git worktree remove $force_flag "$wt_path" || remove_rc=$?
  else
    # Fallback: treat as path directly
    # Look up the branch name for this path
    local abs_target
    abs_target="$(cd "$target" 2>/dev/null && pwd)" || abs_target="$target"
    local fl fl_path fl_hash fl_branch_raw
    while IFS= read -r fl; do
      read -r fl_path fl_hash fl_branch_raw <<< "$fl"
      if [[ "$fl_path" == "$abs_target" ]]; then
        branch_to_delete="${fl_branch_raw#\[}"
        branch_to_delete="${branch_to_delete%\]}"
        break
      fi
    done < <(git worktree list)
    git worktree remove $force_flag "$target" || remove_rc=$?
  fi

  if [[ $remove_rc -ne 0 ]]; then
    return $remove_rc
  fi

  # Also delete the local branch
  if [[ -n "$branch_to_delete" ]]; then
    local branch_flag="-d"
    if [[ "$force_flag" == "--force" ]]; then
      branch_flag="-D"
    fi
    if git branch $branch_flag "$branch_to_delete" 2>/dev/null; then
      echo "$(_ge_bold '✓') Branch '$branch_to_delete' deleted"
    fi
  fi
}

_ge_worktree_dispatch() {
  case "$1" in
    add)            shift; _ge_worktree_add "$@" ;;
    remove|rm)      shift; _ge_worktree_remove "$@" ;;
    list|ls)        _ge_worktree_list ;;
    help|-h|--help) _ge_worktree_help ;;
    "")             _ge_worktree_list ;;
    *)              git worktree "$@" ;;
  esac
}
