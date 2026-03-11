function ge() {
  case "$1" in
    user|worktree|wt)
      local output
      output="$(command ge __run "$@")"
      local rc=$?
      local eval_lines="" print_lines=""
      while IFS= read -r line; do
        case "$line" in
          __GE_EVAL:*) eval_lines+="${line#__GE_EVAL:}"$'\n' ;;
          *) print_lines+="${line}"$'\n' ;;
        esac
      done <<< "$output"
      [[ -n "$print_lines" ]] && printf '%s' "$print_lines"
      [[ -n "$eval_lines" ]] && eval "$eval_lines"
      return $rc ;;
    *) command ge "$@" ;;
  esac
}

# Completion
function _ge() {
  local -a commands user_commands worktree_commands
  commands=(user worktree wt clean branch version help)
  user_commands=(list current add set ssh-key migrate help)
  worktree_commands=(add list ls remove rm prune help)

  case "$CURRENT" in
    2)
      compadd -a commands
      # Also complete git commands
      local -a git_commands
      git_commands=(${(f)"$(git help -a 2>/dev/null | grep '^  ' | awk '{print $1}')"})
      compadd -a git_commands
      ;;
    3)
      case "${words[2]}" in
        user)
          compadd -a user_commands
          # Also complete profile names
          if [[ -f "$HOME/.ge/credentials" ]]; then
            local -a profiles
            profiles=(${(f)"$(grep '^\[' "$HOME/.ge/credentials" | tr -d '[]')"})
            compadd -a profiles
          fi
          ;;
        worktree|wt) compadd -a worktree_commands ;;
      esac
      ;;
  esac
}
compdef _ge ge
