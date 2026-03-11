#!/usr/bin/env bash
# ============================================================
# git-extension installer
# Usage: curl -fsSL https://raw.githubusercontent.com/isac7722/git-extension/main/ge-cli-go/install.sh | bash
#
# Downloads the latest ge binary from GitHub Releases.
# ============================================================

set -e

REPO="isac7722/git-extension"
INSTALL_DIR="${GE_INSTALL_DIR:-/usr/local/bin}"
GE_CONFIG_DIR="$HOME/.ge"

# ── Utility functions ───────────────────────────────────────

log()     { echo "  $1"; }
success() { echo "✔  $1"; }
warn()    { echo "⚠  $1"; }
error()   { echo "✗  $1"; exit 1; }

# ── Detect OS and Architecture ──────────────────────────────

detect_platform() {
  local os arch

  case "$(uname -s)" in
    Darwin*) os="darwin" ;;
    Linux*)  os="linux" ;;
    *)       error "Unsupported OS: $(uname -s)" ;;
  esac

  case "$(uname -m)" in
    x86_64|amd64)  arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *)             error "Unsupported architecture: $(uname -m)" ;;
  esac

  echo "${os}_${arch}"
}

# ── Detect Shell ────────────────────────────────────────────

detect_shell() {
  if [[ -n "$ZSH_VERSION" ]] || [[ "$SHELL" == */zsh ]]; then
    echo "zsh"
  else
    echo "bash"
  fi
}

detect_rc_file() {
  local shell_type="$1"
  case "$shell_type" in
    zsh)  echo "$HOME/.zshrc" ;;
    bash)
      if [[ -f "$HOME/.bashrc" ]]; then
        echo "$HOME/.bashrc"
      elif [[ -f "$HOME/.bash_profile" ]]; then
        echo "$HOME/.bash_profile"
      else
        echo "$HOME/.bashrc"
      fi
      ;;
    *)    echo "$HOME/.profile" ;;
  esac
}

# ── Clean legacy installations ────────────────────────────────

clean_legacy() {
  log "Cleaning legacy installations..."

  # Remove legacy marker blocks from all RC files
  local rc_files=("$HOME/.zshrc" "$HOME/.bashrc" "$HOME/.bash_profile")
  local markers=("# >>> gituser >>>" "# >>> git-extension >>>" "# >>> ge-cli >>>")
  local end_markers=("# <<< gituser <<<" "# <<< git-extension <<<" "# <<< ge-cli <<<")

  for rc in "${rc_files[@]}"; do
    [[ -f "$rc" ]] || continue
    for i in "${!markers[@]}"; do
      if grep -q "${markers[$i]}" "$rc" 2>/dev/null; then
        local tmp_file
        tmp_file="$(mktemp)"
        awk -v start="${markers[$i]}" -v end="${end_markers[$i]}" '
          $0 == start { skip=1; next }
          $0 == end   { skip=0; next }
          !skip { print }
        ' "$rc" > "$tmp_file"
        mv "$tmp_file" "$rc"
      fi
    done
  done

  # npm uninstall -g git-extension
  if command -v npm &>/dev/null; then
    npm uninstall -g git-extension &>/dev/null || true
  fi

  # Remove legacy config directories
  rm -rf "$HOME/.config/gituser" 2>/dev/null || true
  rm -rf "$HOME/.config"/gituser.bak.* 2>/dev/null || true

  # Remove legacy shell project directory
  rm -rf "$HOME/ge-cli-sh" 2>/dev/null || true

  success "Legacy cleanup complete"
}

# ── Start ───────────────────────────────────────────────────

echo ""
echo "╔══════════════════════════════════════╗"
echo "║      git-extension Installer         ║"
echo "╚══════════════════════════════════════╝"
echo ""

clean_legacy

# ── 1. Download binary ──────────────────────────────────────

PLATFORM="$(detect_platform)"
log "Detected platform: $PLATFORM"

# Get latest release tag
LATEST_TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [[ -z "$LATEST_TAG" ]]; then
  error "Failed to determine latest release"
fi
log "Latest version: $LATEST_TAG"

VERSION="${LATEST_TAG#v}"
ARCHIVE_NAME="git-extension_${VERSION}_${PLATFORM}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${ARCHIVE_NAME}"

log "Downloading ${ARCHIVE_NAME}..."
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$ARCHIVE_NAME"
tar -xzf "$TMP_DIR/$ARCHIVE_NAME" -C "$TMP_DIR"

# Install binary
if [[ -w "$INSTALL_DIR" ]]; then
  cp "$TMP_DIR/ge" "$INSTALL_DIR/ge"
else
  log "Requesting sudo to install to $INSTALL_DIR..."
  sudo cp "$TMP_DIR/ge" "$INSTALL_DIR/ge"
fi
chmod +x "$INSTALL_DIR/ge"
success "Installed ge to $INSTALL_DIR/ge"

# ── 2. Shell integration ────────────────────────────────────

SHELL_TYPE="$(detect_shell)"
RC_FILE="$(detect_rc_file "$SHELL_TYPE")"

MARKER_START="# >>> git-extension >>>"
MARKER_END="# <<< git-extension <<<"

if grep -q "$MARKER_START" "$RC_FILE" 2>/dev/null; then
  success "Shell integration already configured in $RC_FILE"
else
  SOURCE_BLOCK="
${MARKER_START}
eval \"\$(command ge init ${SHELL_TYPE})\"
${MARKER_END}"

  touch "$RC_FILE"
  printf '%s\n' "$SOURCE_BLOCK" >> "$RC_FILE"
  success "Shell integration added to $RC_FILE"
fi

# ── 3. Initialize config ────────────────────────────────────

GE_CONFIG="$GE_CONFIG_DIR/credentials"

if [[ ! -f "$GE_CONFIG" ]]; then
  mkdir -p "$GE_CONFIG_DIR"
  cat > "$GE_CONFIG" << 'EXAMPLE'
# ge-cli credentials
# Add your git accounts below:
#
# [profile_name]
# name = Your Name
# email = you@example.com
# ssh_key = ~/.ssh/key_name
EXAMPLE
  success "Config template created: $GE_CONFIG"
  log "  Edit it to add your accounts: \$EDITOR $GE_CONFIG"
else
  success "Config file already exists: $GE_CONFIG"
fi

# ── Done ────────────────────────────────────────────────────

echo ""
echo "╔══════════════════════════════════════╗"
echo "║       Installation Complete!         ║"
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
