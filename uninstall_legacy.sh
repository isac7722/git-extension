#!/usr/bin/env bash
# ============================================================
# ge-cli 레거시(셸 버전) 완전 제거 스크립트
#
# 제거 대상:
#   - npm 글로벌 패키지 (git-extension)
#   - 셸 RC 파일의 마커 블록 (gituser / git-extension / ge-cli)
#   - ~/.config/gituser/ 디렉토리 및 백업
#   - ~/ge-cli-sh/ 소스 디렉토리
#
# 보존:
#   - ~/.ge/credentials (Go 버전에서 사용)
# ============================================================

set -euo pipefail

# ── Utility functions ─────────────────────────────────────────

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
DIM='\033[2m'
RESET='\033[0m'

info()    { echo -e "  ${BOLD}$1${RESET}"; }
success() { echo -e "  ${GREEN}✔${RESET}  $1"; }
warn()    { echo -e "  ${YELLOW}⚠${RESET}  $1"; }
error()   { echo -e "  ${RED}✗${RESET}  $1"; }

# ── RC file targets ───────────────────────────────────────────

RC_FILES=()
[[ -f "$HOME/.zshrc" ]]        && RC_FILES+=("$HOME/.zshrc")
[[ -f "$HOME/.bashrc" ]]       && RC_FILES+=("$HOME/.bashrc")
[[ -f "$HOME/.bash_profile" ]] && RC_FILES+=("$HOME/.bash_profile")

MARKER_PAIRS=(
  "# >>> gituser >>>"        "# <<< gituser <<<"
  "# >>> git-extension >>>"  "# <<< git-extension <<<"
  "# >>> ge-cli >>>"         "# <<< ge-cli <<<"
)

# ── 1. Detect legacy installations ───────────────────────────

echo ""
echo -e "${BOLD}╔══════════════════════════════════════╗${RESET}"
echo -e "${BOLD}║   ge-cli Legacy Uninstaller          ║${RESET}"
echo -e "${BOLD}╚══════════════════════════════════════╝${RESET}"
echo ""

FOUND_ITEMS=()

# Check npm global package
if npm list -g git-extension --depth=0 &>/dev/null; then
  FOUND_ITEMS+=("npm global package: git-extension")
fi

# Check RC file marker blocks
for rc_file in "${RC_FILES[@]}"; do
  for ((i = 0; i < ${#MARKER_PAIRS[@]}; i += 2)); do
    start="${MARKER_PAIRS[$i]}"
    if grep -q "$start" "$rc_file" 2>/dev/null; then
      FOUND_ITEMS+=("marker block in $(basename "$rc_file"): ${start}")
    fi
  done
done

# Check legacy config directories
if [[ -d "$HOME/.config/gituser" ]]; then
  FOUND_ITEMS+=("legacy config: ~/.config/gituser/")
fi

# Check legacy config backups
legacy_backups=()
for bak in "$HOME/.config/gituser.bak."*; do
  if [[ -e "$bak" ]]; then
    legacy_backups+=("$bak")
    FOUND_ITEMS+=("legacy backup: $(basename "$bak")")
  fi
done

# Check legacy source directory
if [[ -d "$HOME/ge-cli-sh" ]]; then
  FOUND_ITEMS+=("legacy source: ~/ge-cli-sh/")
fi

# Nothing found → exit
if [[ ${#FOUND_ITEMS[@]} -eq 0 ]]; then
  success "No legacy ge-cli installation found. Nothing to do."
  echo ""
  exit 0
fi

# ── 2. Show summary and confirm ──────────────────────────────

info "Found legacy items to remove:"
echo ""
for item in "${FOUND_ITEMS[@]}"; do
  echo -e "    • $item"
done
echo ""
echo -e "  ${DIM}~/.ge/credentials will be preserved (used by Go version)${RESET}"
echo ""

read -rp "  Proceed with removal? [y/N] " answer
if [[ "$answer" != "y" && "$answer" != "Y" ]]; then
  echo ""
  info "Aborted."
  echo ""
  exit 0
fi
echo ""

# ── 3. Remove npm global package ─────────────────────────────

if npm list -g git-extension --depth=0 &>/dev/null; then
  info "Removing npm package..."
  npm uninstall -g git-extension 2>/dev/null && \
    success "Removed npm global package: git-extension" || \
    warn "Failed to remove npm package (may need sudo)"
fi

# ── 4. Remove marker blocks from RC files ────────────────────

for rc_file in "${RC_FILES[@]}"; do
  modified=false

  for ((i = 0; i < ${#MARKER_PAIRS[@]}; i += 2)); do
    start="${MARKER_PAIRS[$i]}"
    end="${MARKER_PAIRS[$((i + 1))]}"

    if grep -q "$start" "$rc_file" 2>/dev/null; then
      # Backup before first modification
      if [[ "$modified" == false ]]; then
        cp "$rc_file" "${rc_file}.bak"
      fi

      tmp_file="$(mktemp)"
      awk -v start="$start" -v end="$end" '
        $0 == start { skip=1; next }
        $0 == end   { skip=0; next }
        !skip { print }
      ' "$rc_file" > "$tmp_file"
      mv "$tmp_file" "$rc_file"

      modified=true
      success "Removed ${start} block from $(basename "$rc_file")"
    fi
  done

  if [[ "$modified" == true ]]; then
    info "Backup saved: ${rc_file}.bak"
  fi
done

# ── 5. Remove legacy config directory ─────────────────────────

if [[ -d "$HOME/.config/gituser" ]]; then
  rm -rf "$HOME/.config/gituser"
  success "Removed ~/.config/gituser/"
fi

# ── 6. Remove legacy config backups ──────────────────────────

for bak in "${legacy_backups[@]}"; do
  if [[ -e "$bak" ]]; then
    rm -rf "$bak"
    success "Removed $(basename "$bak")"
  fi
done

# ── 7. Remove legacy source directory ────────────────────────

if [[ -d "$HOME/ge-cli-sh" ]]; then
  rm -rf "$HOME/ge-cli-sh"
  success "Removed ~/ge-cli-sh/"
fi

# ── Done ──────────────────────────────────────────────────────

echo ""
echo -e "${BOLD}╔══════════════════════════════════════╗${RESET}"
echo -e "${BOLD}║       Legacy Removal Complete!       ║${RESET}"
echo -e "${BOLD}╚══════════════════════════════════════╝${RESET}"
echo ""
echo "  To clear legacy functions from the current shell session:"
echo ""
echo -e "    ${BOLD}exec \$SHELL${RESET}"
echo ""
echo "  Or open a new terminal window."
echo ""
echo -e "  ${DIM}~/.ge/credentials has been preserved for the Go version.${RESET}"
echo ""
