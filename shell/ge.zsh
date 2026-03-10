# ============================================================
# ge — Git Extension CLI (zsh shell function)
# ============================================================
# This file is sourced into the user's shell via:
#   eval "$(ge init zsh)"
# ============================================================

# Determine GE_HOME if not already set
if [[ -z "$GE_HOME" ]]; then
  # Try to find from this file's location
  GE_HOME="${0:A:h:h}"
fi

# Source library files
source "$GE_HOME/lib/_ge_core.sh"

# ── Zsh compat overrides (guarantee correct implementations) ──
_ge_split_csv() { reply=(${(s:,:)1}); }
_ge_to_lower() { echo "${1:l}"; }
_ge_declare_map() { typeset -gA "$1"; }
_ge_declare_array() { typeset -ga "$1"; }
_ge_map_has_key() { eval "(( \${+${1}[$2]} ))"; }
_ge_regex_match() {
  if [[ "$1" =~ $2 ]]; then _GE_MATCH=("${match[@]}"); return 0; fi; return 1
}
_ge_comment_check() { [[ "$1" == \#* ]]; }

source "$GE_HOME/lib/_ge_user.sh"
source "$GE_HOME/lib/_ge_worktree.sh"
source "$GE_HOME/lib/_ge_clean.sh"

# ── Main dispatcher ──────────────────────────────────────────

function ge() {
  case "$1" in
    init)
      shift
      command ge init "$@"
      ;;
    user)
      shift
      _ge_user_dispatch "$@"
      ;;
    worktree|wt)
      shift
      _ge_worktree_dispatch "$@"
      ;;
    clean)
      shift
      _ge_clean_dispatch "$@"
      ;;
    update)
      _ge_update
      ;;
    help|--help|-h)
      _ge_help
      ;;
    version|--version|-v)
      _ge_version
      ;;
    "")
      _ge_help
      ;;
    *)
      # Git passthrough
      git "$@"
      ;;
  esac
}

# ── Help ─────────────────────────────────────────────────────

_ge_help() {
  echo ""
  echo "$(_ge_bold 'ge') — Git Extension CLI v${GE_VERSION}"
  echo ""
  echo "$(_ge_bold 'Usage:') ge <command> [args...]"
  echo ""
  echo "$(_ge_bold 'Commands:')"
  printf "  %-20s %s\n" "user [sub]"     "Manage git user accounts"
  printf "  %-20s %s\n" "worktree [sub]" "Enhanced worktree management"
  printf "  %-20s %s\n" "clean [opts]"   "Remove stale local branches"
  printf "  %-20s %s\n" "update"         "Update ge to the latest version"
  printf "  %-20s %s\n" "version"        "Show version"
  printf "  %-20s %s\n" "help"           "Show this help"
  echo ""
  echo "$(_ge_bold 'Git Passthrough:')"
  echo "  Any unrecognized command is passed directly to git."
  echo "  e.g. 'ge commit -m \"msg\"' → 'git commit -m \"msg\"'"
  echo ""
  echo "$(_ge_bold 'Aliases:')"
  printf "  %-20s %s\n" "wt" "Shorthand for 'worktree'"
  echo ""
}

_ge_version() {
  echo "ge ${GE_VERSION}"
}

# ── Self-update ──────────────────────────────────────────────

_ge_update() {
  if [[ -z "$GE_HOME" ]]; then
    echo "$(_ge_red '✗') GE_HOME is not set."
    return 1
  fi

  if [[ ! -d "$GE_HOME/.git" ]]; then
    # npm install — suggest npm update
    echo "$(_ge_bold 'Update git-extension')"
    echo ""
    echo "  Installed via npm. Run:"
    echo "    npm update -g git-extension"
    echo ""
    return 0
  fi

  echo ""
  echo "$(_ge_bold 'Update git-extension')"
  echo "$(_ge_dim '──────────────────────────────')"
  printf "  %-10s %s\n" "Path:" "$(_ge_dim "$GE_HOME")"
  echo ""

  local pull_output
  pull_output="$(git -C "$GE_HOME" pull 2>&1)"
  local pull_status=$?

  if [[ $pull_status -ne 0 ]]; then
    echo "$(_ge_red '✗') git pull failed"
    echo "$pull_output" | sed 's/^/  /'
    return $pull_status
  fi

  if echo "$pull_output" | grep -q "Already up to date"; then
    echo "$(_ge_green '✔') Already up to date."
  else
    echo "$(_ge_green '✔') Updated"
    echo "$pull_output" | sed 's/^/  /'
    echo ""
    echo "  $(_ge_dim 'To apply changes:')"
    echo "    source ~/.zshrc"
  fi
  echo ""
}

# ── Backward compatibility: gituser command ──────────────────

function gituser() {
  echo "$(_ge_dim 'Note: gituser is now ge user. Redirecting...')"
  _ge_user_dispatch "$@"
}

# ── Tab completion ────────────────────────────────────────────

_ge() {
  local cmd="${words[2]}"

  case $CURRENT in
    2) # ge <tab>
      compadd user worktree wt clean update version help
      ;;
    3) # ge <cmd> <tab>
      case "$cmd" in
        user)
          compadd list current add set ssh-key migrate help
          _ge_complete_profiles
          ;;
        worktree|wt)
          compadd add list ls remove rm prune help
          ;;
        clean)
          compadd -- --merged --gone -f --force --dry-run help
          ;;
      esac
      ;;
    *) # deeper args
      case "$cmd" in
        user)
          case "${words[3]}" in
            set|ssh-key) _ge_complete_profiles ;;
          esac
          ;;
        worktree|wt)
          case "${words[3]}" in
            add)       _ge_complete_branches ;;
            remove|rm) _ge_complete_branches ;;
          esac
          ;;
        clean)
          compadd -- --merged --gone -f --force --dry-run
          ;;
      esac
      ;;
  esac
}

_ge_complete_profiles() {
  local config="${GE_CONFIG:-$HOME/.ge/credentials}"
  [[ -f "$config" ]] || return
  local -a profiles
  profiles=(${(f)"$(sed -n 's/^\[\(.*\)\]$/\1/p' "$config")"})
  compadd "${profiles[@]}"
}

_ge_complete_branches() {
  local -a branches
  branches=(${(f)"$(git branch --list --format='%(refname:short)' 2>/dev/null)"})
  compadd "${branches[@]}"
}

compdef _ge ge
