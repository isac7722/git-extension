#!/usr/bin/env bash
# ============================================================
# git-extension postinstall — auto-inject shell integration
# Runs after `npm install -g git-extension`
# ============================================================

set -e

# ── Resolve GE_HOME ──────────────────────────────────────────

SOURCE="${BASH_SOURCE[0]}"
while [[ -L "$SOURCE" ]]; do
  DIR="$(cd -P "$(dirname "$SOURCE")" && pwd)"
  SOURCE="$(readlink "$SOURCE")"
  [[ "$SOURCE" != /* ]] && SOURCE="$DIR/$SOURCE"
done
GE_HOME="$(cd -P "$(dirname "$SOURCE")/.." && pwd)"

# ── Detect shell & RC file ───────────────────────────────────

case "$(uname -s)" in
  Darwin*)
    SHELL_TYPE="zsh"
    RC_FILE="$HOME/.zshrc"
    ;;
  Linux*)
    if [[ "$SHELL" == */zsh ]]; then
      SHELL_TYPE="zsh"
      RC_FILE="$HOME/.zshrc"
    else
      SHELL_TYPE="bash"
      RC_FILE="$HOME/.bashrc"
    fi
    ;;
  *)
    echo "  [ge] Unsupported OS — add shell integration manually:"
    echo "    eval \"\$(ge init <shell>)\""
    exit 0
    ;;
esac

# ── Inject eval block (idempotent) ───────────────────────────

MARKER_START="# >>> git-extension >>>"
MARKER_END="# <<< git-extension <<<"

if grep -qF "$MARKER_START" "$RC_FILE" 2>/dev/null; then
  echo "  [ge] Shell integration already exists in $RC_FILE"
  exit 0
fi

touch "$RC_FILE"

cat >> "$RC_FILE" <<EOF

${MARKER_START}
# git-extension shell integration (added by npm postinstall)
eval "\$(command ge init $SHELL_TYPE)"
${MARKER_END}
EOF

echo "  [ge] Shell integration added to $RC_FILE"
echo "  [ge] Run 'source $RC_FILE' or restart your terminal."
