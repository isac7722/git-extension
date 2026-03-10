#!/usr/bin/env bash
# ============================================================
# git-extension installer
# Usage: ./install.sh [--dry-run]
#
# Install methods:
#   1. Local:  ./install.sh        (from cloned repo)
#   2. Remote: bash <(curl -fsSL https://raw.githubusercontent.com/.../install.sh)
#
# What it does:
#   1. Clone repo to ~/.ge (if remote install)
#   2. Detect OS and shell
#   3. Inject eval "$(ge init <shell>)" into RC file
#   4. Initialize Git user account config file
# ============================================================

set -e

GE_REPO="https://github.com/isac7722/git-extension.git"
GE_DEFAULT_DIR="$HOME/.ge"
GE_CONFIG_DIR="$HOME/.ge"

# Detect if running from cloned repo or remote
GE_HOME="$(cd "$(dirname "${BASH_SOURCE[0]}")" 2>/dev/null && pwd)"

if [[ -z "$GE_HOME" || ! -f "$GE_HOME/bin/ge" ]]; then
  echo ""
  echo "╔══════════════════════════════════════╗"
  echo "║       git-extension Remote Install     ║"
  echo "╚══════════════════════════════════════╝"
  echo ""

  if [[ -d "$GE_DEFAULT_DIR/.git" ]]; then
    echo "  Existing git-extension found: $GE_DEFAULT_DIR"
    echo "  Updating to latest version..."
    git -C "$GE_DEFAULT_DIR" pull --ff-only
  else
    echo "  Cloning repository: $GE_REPO"
    git clone "$GE_REPO" "$GE_DEFAULT_DIR"
  fi

  echo ""
  exec bash "$GE_DEFAULT_DIR/install.sh" "$@"
fi

DRY_RUN=false

# ── Option parsing ──────────────────────────────────────────

for arg in "$@"; do
  case $arg in
    --dry-run) DRY_RUN=true ;;
  esac
done

# ── Utility functions ───────────────────────────────────────

log()     { echo "  $1"; }
success() { echo "✔  $1"; }
warn()    { echo "⚠  $1"; }
error()   { echo "✗  $1"; exit 1; }
section() { echo ""; echo "── $1 ──────────────────────────────"; }

# ── Start ───────────────────────────────────────────────────

echo ""
echo "╔══════════════════════════════════════╗"
echo "║       git-extension Installer          ║"
echo "╚══════════════════════════════════════╝"
$DRY_RUN && warn "DRY-RUN mode: no changes will be made"

# ── 1. Detect OS & Shell ──────────────────────────────────────

section "Detecting OS & Shell"

case "$(uname -s)" in
  Darwin*)
    OS="macOS"
    RC_FILE="$HOME/.zshrc"
    SHELL_TYPE="zsh"
    ;;
  Linux*)
    OS="Linux"
    if [[ -n "$ZSH_VERSION" ]] || [[ "$SHELL" == */zsh ]]; then
      RC_FILE="$HOME/.zshrc"
      SHELL_TYPE="zsh"
    else
      RC_FILE="$HOME/.bashrc"
      SHELL_TYPE="bash"
    fi
    ;;
  *)
    error "Unsupported OS: $(uname -s)"
    ;;
esac

success "Detected: $OS / $SHELL_TYPE → $RC_FILE"

# ── 2. Make bin/ge executable ─────────────────────────────────

chmod +x "$GE_HOME/bin/ge"

# ── 3. Add bin to PATH (if not already) ──────────────────────

section "PATH setup"

GE_BIN_DIR="$GE_HOME/bin"

if echo "$PATH" | tr ':' '\n' | grep -qx "$GE_BIN_DIR"; then
  success "bin/ge already in PATH"
else
  log "bin/ge will be available after shell integration"
fi

# ── 3.5. Migrate legacy gituser data ─────────────────────────

LEGACY_DIR="$HOME/.config/gituser"
LEGACY_CONFIG="$LEGACY_DIR/accounts"
LEGACY_PROFILES="$LEGACY_DIR/profiles"

if [[ -d "$LEGACY_DIR" ]]; then
  section "Legacy gituser migration"

  # a. Account migration
  if [[ -f "$LEGACY_CONFIG" ]]; then
    has_real_data=false
    if [[ -f "$GE_CONFIG_DIR/credentials" ]]; then
      while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        [[ "${line:0:1}" == "#" ]] && continue
        has_real_data=true
        break
      done < "$GE_CONFIG_DIR/credentials"
    fi

    if $has_real_data; then
      log "Credentials already exist, skipping account migration"
    else
      count=0
      ini_output=""
      while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        [[ "${line:0:1}" == "#" ]] && continue
        raw_aliases="${line%%:*}"
        rest="${line#*:}"
        name="${rest%%:*}"
        rest="${rest#*:}"
        email="${rest%%:*}"
        key_path="${rest#*:}"
        profile="${raw_aliases%%,*}"
        profile="${profile// /}"
        [[ -n "$ini_output" ]] && ini_output="${ini_output}
"
        ini_output="${ini_output}[$profile]
name = $name
email = $email
ssh_key = $key_path"
        count=$((count + 1))
      done < "$LEGACY_CONFIG"

      if [[ $count -gt 0 ]]; then
        if $DRY_RUN; then
          log "[dry-run] Would migrate $count account(s) to $GE_CONFIG_DIR/credentials"
        else
          mkdir -p "$GE_CONFIG_DIR"
          printf '%s\n' "$ini_output" > "$GE_CONFIG_DIR/credentials"
          success "Migrated $count account(s) from legacy config"
        fi
      fi
    fi
  fi

  # b. Profile copy
  if [[ -d "$LEGACY_PROFILES" ]]; then
    if $DRY_RUN; then
      log "[dry-run] Would copy profiles to $GE_CONFIG_DIR/profiles"
    else
      mkdir -p "$GE_CONFIG_DIR/profiles"
      cp -n "$LEGACY_PROFILES"/* "$GE_CONFIG_DIR/profiles/" 2>/dev/null || true
      success "Profiles copied from legacy directory"
    fi
  fi

  # c. gitconfig marker conversion
  GITCONFIG="$HOME/.gitconfig"
  if [[ -f "$GITCONFIG" ]] && grep -q "gituser" "$GITCONFIG" 2>/dev/null; then
    if $DRY_RUN; then
      log "[dry-run] Would update gitconfig markers"
    else
      tmp_gc="$(mktemp)"
      sed -e 's/# gituser:rule:/# ge:rule:/g' \
          -e 's|\.config/gituser/profiles/|.ge/profiles/|g' \
          "$GITCONFIG" > "$tmp_gc"
      mv "$tmp_gc" "$GITCONFIG"
      success "Updated gitconfig markers"
    fi
  fi

  # d. Backup legacy directory (RC block cleanup happens in step 4)
  if $DRY_RUN; then
    log "[dry-run] Would backup $LEGACY_DIR"
  else
    backup_name="${LEGACY_DIR}.bak.$(date +%Y%m%d)"
    mv "$LEGACY_DIR" "$backup_name"
    success "Backed up legacy data → $backup_name"
  fi

  # e. gituser command removal hint
  if command -v gituser &>/dev/null; then
    echo ""
    warn "Legacy 'gituser' command is still installed."
    log "  To remove it: npm uninstall -g gituser"
  fi
fi

# ── 4. Inject shell integration into RC file ──────────────────

section "Shell integration ($RC_FILE)"

MARKER_START="# >>> git-extension >>>"
MARKER_END="# <<< git-extension <<<"

SOURCE_BLOCK="
${MARKER_START}
# git-extension shell integration (generated by install.sh)
export GE_HOME=\"$GE_HOME\"
export PATH=\"\$GE_HOME/bin:\$PATH\"
eval \"\$(ge init $SHELL_TYPE)\"
${MARKER_END}"

# Also remove legacy gituser block if present
LEGACY_START="# >>> gituser >>>"
LEGACY_END="# <<< gituser <<<"

if grep -q "$MARKER_START" "$RC_FILE" 2>/dev/null; then
  success "Already integrated: $RC_FILE"
elif $DRY_RUN; then
  log "[dry-run] Source block will be added to $RC_FILE"
  log "[dry-run] Block contents:"
  echo "$SOURCE_BLOCK" | sed 's/^/    /'
else
  touch "$RC_FILE"

  # Remove legacy gituser block if present
  if grep -q "$LEGACY_START" "$RC_FILE" 2>/dev/null; then
    tmp_file="$(mktemp)"
    awk -v start="$LEGACY_START" -v end="$LEGACY_END" '
      $0 ~ start { skip=1; next }
      $0 ~ end   { skip=0; next }
      !skip { print }
    ' "$RC_FILE" > "$tmp_file"
    mv "$tmp_file" "$RC_FILE"
    log "Removed legacy gituser block"
  fi

  printf '%s\n' "$SOURCE_BLOCK" >> "$RC_FILE"
  success "Source block added: $RC_FILE"
fi

# ── 5. Initialize Git user config file ────────────────────────

section "Git User Config File"

GE_CONFIG="$GE_CONFIG_DIR/credentials"
GE_EXAMPLE="$GE_HOME/config/credentials.example"

if [[ ! -f "$GE_CONFIG" ]]; then
  if $DRY_RUN; then
    log "[dry-run] Will create $GE_CONFIG"
    log "[dry-run] Template: $GE_EXAMPLE"
  else
    mkdir -p "$GE_CONFIG_DIR"
    cp "$GE_EXAMPLE" "$GE_CONFIG"
    success "Config file created: $GE_CONFIG"
    echo ""
    log "▶ Edit the file below to register your accounts:"
    log "     \$EDITOR $GE_CONFIG"
    log ""
    log "  Format:"
    log "    [profile_name]"
    log "    name = Your Name"
    log "    email = you@example.com"
    log "    ssh_key = ~/.ssh/key_name"
  fi
else
  success "Config file already exists: $GE_CONFIG"
fi

# ── Done ────────────────────────────────────────────────────

echo ""
echo "╔══════════════════════════════════════╗"
echo "║       Installation Complete!          ║"
echo "╚══════════════════════════════════════╝"
echo ""
echo "  Apply changes immediately:"
echo "    source $RC_FILE"
echo ""
echo "  Usage:"
echo "    ge user list      # list accounts"
echo "    ge user <profile> # switch account"
echo "    ge worktree add   # create worktree"
echo "    ge commit -m msg  # git passthrough"
echo ""
