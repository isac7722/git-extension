# ============================================================
# ge user — Git user account management
# ============================================================
# Requires: _ge_core.sh (sourced before this file)
# ============================================================

# ── Internal: switch account ─────────────────────────────────

_ge_user_switch() {
  local name="$1"
  local email="$2"
  local key_path="$3"
  local scope="${4:---global}"

  if [[ ! -f "$key_path" ]]; then
    echo "$(_ge_red '✗') SSH key not found: $(_ge_yellow "$key_path")"
    echo "  On a new machine? See the 'SSH Key Setup' section in the README."
    return 1
  fi

  git config "$scope" user.name  "$name"
  git config "$scope" user.email "$email"
  git config "$scope" core.sshCommand "ssh -i $key_path"

  if [[ "$scope" == "--global" ]]; then
    unset GIT_SSH_COMMAND
    export GIT_SSH_COMMAND="ssh -i $key_path"
    eval "$(ssh-agent -s)" > /dev/null 2>&1
    ssh-add --apple-use-keychain "$key_path" 2>/dev/null || ssh-add "$key_path" 2>/dev/null
  fi

  local scope_label="global"
  [[ "$scope" == "--local" ]] && scope_label="local (this repo only)"

  echo ""
  echo "$(_ge_green '✔') $(_ge_bold 'Git account switched') $(_ge_dim "[$scope_label]")"
  printf "  %-10s %s\n" "Name:"    "$(_ge_bold "$name")"
  printf "  %-10s %s\n" "Email:"   "$email"
  printf "  %-10s %s\n" "SSH Key:" "$(_ge_dim "$key_path")"
  echo ""
}

# ── List accounts ────────────────────────────────────────────

_ge_user_list() {
  # zsh: make arrays 0-indexed like bash (scoped to this function only)
  [ -n "$ZSH_VERSION" ] && setopt localoptions KSH_ARRAYS

  _ge_load_accounts

  if [[ ${#_GE_ACCOUNTS_RAW[@]} -eq 0 ]]; then
    echo ""
    echo "  $(_ge_yellow '⚠') No accounts registered."
    echo "  Check $GE_CONFIG or run 'ge user add' to add one."
    echo ""
    return
  fi

  local current_name current_email
  current_name="$(git config --global user.name 2>/dev/null)"
  current_email="$(git config --global user.email 2>/dev/null)"

  # Build arrays of display data
  local profiles=() names=() emails=()
  local entry profile rest name email key_path
  for entry in "${_GE_ACCOUNTS_RAW[@]}"; do
    profile="${entry%%:*}"
    rest="${entry#*:}"
    name="${rest%%:*}"
    rest="${rest#*:}"
    email="${rest%%:*}"
    profiles+=("$profile")
    names+=("$name")
    emails+=("$email")
  done

  local total=${#profiles[@]}
  local selected=0

  # Find current user index for initial cursor position
  local i
  for (( i=0; i<total; i++ )); do
    if [[ "${names[$i]}" == "$current_name" && "${emails[$i]}" == "$current_email" ]]; then
      selected=$i
      break
    fi
  done

  # Hide cursor and set up cleanup
  tput civis 2>/dev/null
  trap 'tput cnorm 2>/dev/null; trap - INT; return' INT

  _ge_user_list_render() {
    # Move cursor up to redraw (skip on first render)
    if [[ "${1:-}" == "redraw" ]]; then
      local lines_up=$(( 2 + total * 2 + 2 ))
      printf "\033[%dA\033[J" "$lines_up"
    fi

    printf "  %s\n" "$(_ge_bold 'Git User Accounts')"
    printf "  %s\n" "$(_ge_dim '──────────────────────────────────────────────')"

    local idx pad_len
    for (( idx=0; idx<total; idx++ )); do
      local raw_name="${names[$idx]}"
      # Calculate padding based on raw name length (no ANSI)
      pad_len=$(( 20 - ${#raw_name} ))
      (( pad_len < 1 )) && pad_len=1
      local padding=""
      padding="$(printf '%*s' "$pad_len" "")"

      if [[ $idx -eq $selected ]]; then
        printf "  %s %s%s%s" "$(_ge_cyan '❯')" "$(_ge_bold "$raw_name")" "$padding" "$(_ge_dim "${emails[$idx]}")"
      else
        printf "    %s%s%s" "$raw_name" "$padding" "$(_ge_dim "${emails[$idx]}")"
      fi

      if [[ "${names[$idx]}" == "$current_name" && "${emails[$idx]}" == "$current_email" ]]; then
        printf "  %s" "$(_ge_green '✔ current')"
      fi
      printf "\n"

      printf "      %s %s\n" "$(_ge_dim 'profile:')" "$(_ge_dim "${profiles[$idx]}")"
    done

    printf "  %s\n" "$(_ge_dim '──────────────────────────────────────────────')"
    printf "  %s\n" "$(_ge_dim '↑↓ navigate  ⏎ select  q cancel')"
  }

  # Initial render
  echo ""
  _ge_user_list_render

  # Read input loop — use zsh-compatible read flags
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
    # Read a single character (raw mode)
    _read_char key

    case "$key" in
      $'\x1b')
        # Escape sequence — read next two chars
        _read_chars key
        case "$key" in
          '[A') # Up arrow
            (( selected > 0 )) && (( selected-- ))
            ;;
          '[B') # Down arrow
            (( selected < total - 1 )) && (( selected++ ))
            ;;
        esac
        ;;
      'k') # vim up
        (( selected > 0 )) && (( selected-- ))
        ;;
      'j') # vim down
        (( selected < total - 1 )) && (( selected++ ))
        ;;
      ''|$'\n') # Enter
        tput cnorm 2>/dev/null
        trap - INT
        _ge_user_do "${profiles[$selected]}"
        return
        ;;
      'q'|$'\x03') # q or Ctrl-C
        tput cnorm 2>/dev/null
        trap - INT
        echo ""
        return 0
        ;;
    esac

    _ge_user_list_render "redraw"
  done
}

# ── Help ─────────────────────────────────────────────────────

_ge_user_help() {
  _ge_load_accounts

  echo ""
  echo "$(_ge_bold 'Usage:') ge user <profile | subcommand>"
  echo ""
  echo "$(_ge_bold 'Subcommands:')"
  printf "  %-28s %s\n" "list"                  "List all registered accounts"
  printf "  %-28s %s\n" "current"               "Show the current git account"
  printf "  %-28s %s\n" "add"                   "Interactively register an account"
  printf "  %-28s %s\n" "set <profile>"         "Apply account to current repo only (--local)"
  printf "  %-28s %s\n" "ssh-key <p> [path]"    "View or update SSH key path"
  printf "  %-28s %s\n" "migrate"               "Migrate from legacy config format"
  printf "  %-28s %s\n" "<profile>"             "Switch to the specified account (--global)"
  echo ""

  if [[ ${#_GE_ACCOUNTS_RAW[@]} -gt 0 ]]; then
    echo "$(_ge_bold 'Available profiles:')"
    local entry profile rest name
    for entry in "${_GE_ACCOUNTS_RAW[@]}"; do
      profile="${entry%%:*}"
      rest="${entry#*:}"
      name="${rest%%:*}"
      printf "  %-24s → %s\n" "$(_ge_cyan "$profile")" "$name"
    done
    echo ""
  else
    echo "$(_ge_yellow '⚠') No accounts found in config: $GE_CONFIG"
    echo "  Run 'ge user add' to add an account."
    echo ""
  fi
}

# ── fzf interactive selection ────────────────────────────────

_ge_user_fzf() {
  _ge_load_accounts

  if [[ ${#_GE_ACCOUNTS_RAW[@]} -eq 0 ]]; then
    echo "$(_ge_yellow '⚠') No accounts registered. Run 'ge user add' to add one."
    return 1
  fi

  local current_name
  current_name="$(git config --global user.name 2>/dev/null)"

  local options=()
  local entry profile rest name email key_path mark
  for entry in "${_GE_ACCOUNTS_RAW[@]}"; do
    profile="${entry%%:*}"
    rest="${entry#*:}"
    name="${rest%%:*}"
    rest="${rest#*:}"
    email="${rest%%:*}"
    key_path="${rest#*:}"
    mark=""
    [[ "$name" == "$current_name" ]] && mark=" ✔"
    options+=("${profile}  ${name}  <${email}>${mark}")
  done

  local selected
  selected="$(printf '%s\n' "${options[@]}" | fzf \
    --prompt="  Git User > " \
    --header="Select an account (Enter: switch, Esc: cancel)" \
    --height=40% \
    --reverse \
    --no-info)"

  [[ -z "$selected" ]] && return 0

  local chosen_alias="${selected%% *}"
  _ge_user_do "$chosen_alias"
}

# ── Perform switch by alias ──────────────────────────────────

_ge_user_do() {
  local alias_key="$1"
  local scope="${2:---global}"
  _ge_load_accounts

  if ! _ge_map_has_key _GE_USER_MAP "$alias_key"; then
    echo "$(_ge_red '✗') Unknown profile: '$(_ge_yellow "$alias_key")'"
    _ge_user_help
    return 1
  fi

  local entry="${_GE_USER_MAP[$alias_key]}"
  local name="${entry%%:*}"
  local rest="${entry#*:}"
  local email="${rest%%:*}"
  local key_path="${rest#*:}"

  _ge_user_switch "$name" "$email" "$key_path" "$scope"
}

# ── Interactive account registration ─────────────────────────

_ge_user_add() {
  echo ""
  echo "$(_ge_bold 'Register account') → $GE_CONFIG"
  echo "$(_ge_dim '──────────────────────────────')"
  echo ""

  local gu_profile gu_name gu_email gu_key

  printf "  Profile name $(_ge_dim '(e.g. work, personal)'): "
  read -r gu_profile
  [[ -z "$gu_profile" ]] && echo "$(_ge_red '✗') Profile name is required." && return 1

  printf "  Name $(_ge_dim '(git user.name)'): "
  read -r gu_name
  [[ -z "$gu_name" ]] && echo "$(_ge_red '✗') Name is required." && return 1

  printf "  Email $(_ge_dim '(git user.email)'): "
  read -r gu_email
  [[ -z "$gu_email" ]] && echo "$(_ge_red '✗') Email is required." && return 1

  printf "  SSH key path $(_ge_dim '(e.g. ~/.ssh/work_ed25519)'): "
  read -r gu_key
  gu_key="${gu_key/#\~/$HOME}"
  if [[ ! -f "$gu_key" ]]; then
    echo "$(_ge_yellow '⚠') SSH key not found: $gu_key"
    printf "  Continue anyway? $(_ge_dim '[y/N]'): "
    read -r confirm
    [[ "$(_ge_to_lower "$confirm")" != "y" ]] && echo "Cancelled." && return 0
  fi

  mkdir -p "$(dirname "$GE_CONFIG")"
  cat >> "$GE_CONFIG" <<EOF

[$gu_profile]
name = $gu_name
email = $gu_email
ssh_key = ${gu_key/#$HOME/\~}
EOF

  echo ""
  echo "$(_ge_green '✔') Account added"
  printf "  %-10s %s\n" "Profile:" "$(_ge_bold "$gu_profile")"
  printf "  %-10s %s\n" "Name:"    "$(_ge_bold "$gu_name")"
  printf "  %-10s %s\n" "Email:"   "$gu_email"
  echo ""
}

# ── Update SSH key path ──────────────────────────────────────

_ge_user_ssh_key() {
  local alias_key="$1"
  local new_key="$2"

  if [[ -z "$alias_key" ]]; then
    echo "$(_ge_bold 'Usage:') ge user ssh-key <profile> [new_path]"
    echo ""
    echo "  Providing only a profile shows the current SSH key path."
    echo "  Providing new_path updates it to that path."
    echo ""
    return 1
  fi

  _ge_load_accounts

  if ! _ge_map_has_key _GE_USER_MAP "$alias_key"; then
    echo "$(_ge_red '✗') Unknown profile: '$(_ge_yellow "$alias_key")'"
    return 1
  fi

  local entry="${_GE_USER_MAP[$alias_key]}"
  local name="${entry%%:*}"
  local rest="${entry#*:}"
  local email="${rest%%:*}"
  local old_key="${rest#*:}"

  # View path only
  if [[ -z "$new_key" ]]; then
    echo ""
    printf "  %-10s %s\n" "Account:"  "$(_ge_bold "$name") $(_ge_dim "<$email>")"
    printf "  %-10s %s\n" "SSH Key:"  "$old_key"
    if [[ -f "$old_key" ]]; then
      printf "  %-10s %s\n" "Status:" "$(_ge_green 'File exists')"
    else
      printf "  %-10s %s\n" "Status:" "$(_ge_red 'File not found')"
    fi
    echo ""
    return 0
  fi

  # Handle new path
  new_key="${new_key/#\~/$HOME}"
  if [[ ! -f "$new_key" ]]; then
    echo "$(_ge_yellow '⚠') SSH key not found: $new_key"
    printf "  Continue anyway? $(_ge_dim '[y/N]'): "
    read -r confirm
    [[ "$(_ge_to_lower "$confirm")" != "y" ]] && echo "Cancelled." && return 0
  fi

  local new_key_display="${new_key/#$HOME/\~}"

  # Replace ssh_key line in the matching profile section
  local tmp_file
  tmp_file="$(mktemp)"
  local matched=false
  local in_target=false

  while IFS= read -r line; do
    # Section header
    if [[ "$line" == \[*\] ]]; then
      local section="${line#\[}"
      section="${section%\]}"
      in_target=false
      [[ "$section" == "$alias_key" ]] && in_target=true
    fi

    # Replace ssh_key in target section
    local trimmed="${line// /}"
    if $in_target && [[ "$trimmed" == ssh_key=* ]]; then
      echo "ssh_key = $new_key_display" >> "$tmp_file"
      matched=true
    else
      echo "$line" >> "$tmp_file"
    fi
  done < "$GE_CONFIG"

  if ! $matched; then
    rm -f "$tmp_file"
    echo "$(_ge_red '✗') Account not found in config file."
    return 1
  fi

  mv "$tmp_file" "$GE_CONFIG"


  echo ""
  echo "$(_ge_green '✔') SSH key path updated"
  printf "  %-10s %s\n" "Account:" "$(_ge_bold "$name")"
  printf "  %-10s %s\n" "Before:"  "$(_ge_dim "$old_key")"
  printf "  %-10s %s\n" "After:"   "$(_ge_bold "$new_key")"
  echo ""
}

# ── Per-repo local config ────────────────────────────────────

_ge_user_set() {
  local alias_key="$1"

  if [[ -z "$alias_key" ]]; then
    echo "$(_ge_red '✗') Usage: ge user set <profile>"
    return 1
  fi

  if ! git rev-parse --git-dir &>/dev/null; then
    echo "$(_ge_red '✗') Current directory is not a git repository."
    return 1
  fi

  _ge_user_do "$alias_key" "--local"
}

# ── Current account ──────────────────────────────────────────

_ge_user_current() {
  local name email
  name="$(git config --global user.name 2>/dev/null)"
  email="$(git config --global user.email 2>/dev/null)"
  echo ""
  echo "$(_ge_bold 'Current Git Account (global)')"
  echo "$(_ge_dim '──────────────────────────────')"
  printf "  %-10s %s\n" "Name:"  "$(_ge_bold "${name:-(not set)}")"
  printf "  %-10s %s\n" "Email:" "${email:-(not set)}"
  echo ""
}

# ── Migrate from legacy config ────────────────────────────────

_ge_user_migrate() {
  local legacy_config="$HOME/.config/gituser/accounts"
  local legacy_profiles="$HOME/.config/gituser/profiles"

  if [[ ! -f "$legacy_config" ]]; then
    echo "$(_ge_yellow '⚠') Legacy config not found: $legacy_config"
    echo "  Nothing to migrate."
    return 1
  fi

  echo ""
  echo "$(_ge_bold 'Migrating legacy config') → $GE_CONFIG"
  echo "$(_ge_dim '──────────────────────────────────────────────')"

  mkdir -p "$(dirname "$GE_CONFIG")"

  local count=0
  local ini_output=""

  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    _ge_comment_check "$line" && continue

    local raw_aliases="${line%%:*}"
    local rest="${line#*:}"
    local name="${rest%%:*}"
    rest="${rest#*:}"
    local email="${rest%%:*}"
    local key_path="${rest#*:}"

    # Use first alias as profile name
    local profile="${raw_aliases%%,*}"
    profile="${profile// /}"

    if [[ -n "$ini_output" ]]; then
      ini_output="${ini_output}
"
    fi
    ini_output="${ini_output}[$profile]
name = $name
email = $email
ssh_key = $key_path"

    count=$((count + 1))
  done < "$legacy_config"

  if [[ $count -eq 0 ]]; then
    echo "  $(_ge_yellow '⚠') No accounts found in legacy config."
    return 1
  fi

  printf '%s\n' "$ini_output" > "$GE_CONFIG"

  # Copy profiles directory if it exists
  if [[ -d "$legacy_profiles" ]]; then
    mkdir -p "$GE_PROFILES_DIR"
    cp -r "$legacy_profiles"/* "$GE_PROFILES_DIR/" 2>/dev/null || true
    echo "  $(_ge_dim "Profiles copied: $legacy_profiles → $GE_PROFILES_DIR")"
  fi

  # Backup legacy config
  cp "$legacy_config" "${legacy_config}.bak"
  echo "  $(_ge_dim "Legacy config backed up: ${legacy_config}.bak")"

  echo ""
  echo "$(_ge_green '✔') Migrated $count account(s) to $GE_CONFIG"
  echo ""
}

# ── User subcommand dispatcher ───────────────────────────────

_ge_user_dispatch() {
  case "$1" in
    "")
      if command -v fzf &>/dev/null; then
        _ge_user_fzf
      else
        _ge_user_help
      fi
      ;;
    list)           _ge_user_list ;;
    current|now)    _ge_user_current ;;
    add)            _ge_user_add ;;
    set)            shift; _ge_user_set "$@" ;;
    ssh-key)        shift; _ge_user_ssh_key "$@" ;;
    migrate)        _ge_user_migrate ;;
    help|-h|--help) _ge_user_help ;;
    *)              _ge_user_do "$1" ;;
  esac
}
