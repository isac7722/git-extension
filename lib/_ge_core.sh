# ============================================================
# git-extension core library
# ============================================================
# Shared utilities: color helpers, config loading, bash/zsh compat layer
# This file is sourced by both shell functions and bin/ge
# ============================================================

# ── Config ──────────────────────────────────────────────────

GE_CONFIG="${GE_CONFIG:-$HOME/.ge/credentials}"
GE_PROFILES_DIR="${GE_PROFILES_DIR:-$HOME/.ge/profiles}"

# ── Bash/Zsh compatibility layer ────────────────────────────

if [ -n "$ZSH_VERSION" ]; then
  _ge_split_csv() { reply=(${(s:,:)1}); }
  _ge_to_lower() { echo "${1:l}"; }
  _ge_declare_map() { typeset -gA "$1"; }
  _ge_declare_array() { typeset -ga "$1"; }
  _ge_map_has_key() {
    local -n _map="$1"
    [[ -n "${_map[$2]}" ]]
  }
  _ge_regex_match() {
    # $1=string $2=pattern; sets _GE_MATCH array
    # zsh: 1-indexed (match[1]=group1), bash: 0-indexed (MATCH[0]=group1)
    if [[ "$1" =~ $2 ]]; then
      _GE_MATCH=("${match[@]}")
      return 0
    fi
    return 1
  }
  _ge_comment_check() { [[ "$1" == \#* ]]; }
else
  _ge_split_csv() { IFS=',' read -ra reply <<< "$1"; }
  _ge_to_lower() { echo "${1,,}"; }
  _ge_declare_map() { declare -gA "$1"; }
  _ge_declare_array() { declare -ga "$1"; }
  _ge_map_has_key() {
    local -n _map="$1"
    [[ -n "${_map[$2]+x}" ]]
  }
  _ge_regex_match() {
    if [[ "$1" =~ $2 ]]; then
      _GE_MATCH=("${BASH_REMATCH[@]:1}")
      return 0
    fi
    return 1
  }
  _ge_comment_check() { [[ "${1:0:1}" == "#" ]]; }
fi

# ── Color helpers ────────────────────────────────────────────

_ge_green()  { printf "\033[32m%s\033[0m" "$*"; }
_ge_yellow() { printf "\033[33m%s\033[0m" "$*"; }
_ge_red()    { printf "\033[31m%s\033[0m" "$*"; }
_ge_bold()   { printf "\033[1m%s\033[0m" "$*"; }
_ge_dim()    { printf "\033[2m%s\033[0m" "$*"; }
_ge_cyan()   { printf "\033[36m%s\033[0m" "$*"; }

# ── Account loading ──────────────────────────────────────────
#
# _GE_USER_MAP:          profile -> "name:email:key_path"
# _GE_ACCOUNTS_RAW:      raw line list (used for list output)

_ge_load_accounts() {
  _ge_declare_map _GE_USER_MAP
  _ge_declare_array _GE_ACCOUNTS_RAW
  _GE_USER_MAP=()
  _GE_ACCOUNTS_RAW=()

  if [[ ! -f "$GE_CONFIG" ]]; then
    return 1
  fi

  local profile="" name="" email="" key_path=""

  _ge_flush_profile() {
    if [[ -n "$profile" && -n "$name" && -n "$email" && -n "$key_path" ]]; then
      local expanded_key="$key_path"
      expanded_key="${expanded_key/#\\\~/~}"
      expanded_key="${expanded_key/#\~/$HOME}"
      local value="${name}:${email}:${expanded_key}"
      _GE_USER_MAP["$profile"]="$value"
      _GE_ACCOUNTS_RAW+=("${profile}:${value}")
    fi
  }

  while IFS= read -r line; do
    # Skip empty lines and comments
    [[ -z "$line" ]] && continue
    _ge_comment_check "$line" && continue

    # Section header: [profile_name]
    if [[ "$line" == \[*\] ]]; then
      _ge_flush_profile
      profile="${line#\[}"
      profile="${profile%\]}"
      name="" email="" key_path=""
      continue
    fi

    # Key = value pairs
    local key="${line%%=*}"
    local val="${line#*=}"
    key="${key// /}"     # trim spaces
    val="${val#"${val%%[![:space:]]*}"}"  # trim leading spaces

    case "$key" in
      name)    name="$val" ;;
      email)   email="$val" ;;
      ssh_key) key_path="$val" ;;
    esac
  done < "$GE_CONFIG"

  # Flush the last profile
  _ge_flush_profile
  unset -f _ge_flush_profile
}

# ── Version ──────────────────────────────────────────────────

GE_VERSION="1.0.0"
