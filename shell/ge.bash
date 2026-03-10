# ============================================================
# ge — Git Extension CLI (bash shell function)
# ============================================================
# This file is sourced into the user's shell via:
#   eval "$(ge init bash)"
# ============================================================

# Requires bash 4+ (associative array support)
if ((BASH_VERSINFO[0] < 4)); then
  echo "ge: bash 4.0 or later is required. (current: $BASH_VERSION)" >&2
  return 1
fi

# Determine GE_HOME if not already set
if [[ -z "$GE_HOME" ]]; then
  GE_HOME="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fi

# Source library files
source "$GE_HOME/lib/_ge_core.sh"

# ── Bash compat overrides ──
_ge_split_csv() { IFS=',' read -ra reply <<< "$1"; }
_ge_to_lower() { echo "${1,,}"; }
_ge_declare_map() { declare -gA "$1"; }
_ge_declare_array() { declare -ga "$1"; }
_ge_map_has_key() { local -n _map="$1"; [[ -n "${_map[$2]+x}" ]]; }
_ge_regex_match() {
  if [[ "$1" =~ $2 ]]; then _GE_MATCH=("${BASH_REMATCH[@]:1}"); return 0; fi; return 1
}
_ge_comment_check() { [[ "${1:0:1}" == "#" ]]; }

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
    echo "    source ~/.bashrc"
  fi
  echo ""
}

# ── Backward compatibility: gituser command ──────────────────

function gituser() {
  echo "$(_ge_dim 'Note: gituser is now ge user. Redirecting...')"
  _ge_user_dispatch "$@"
}
