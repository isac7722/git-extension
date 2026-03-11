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
_ge_completion() {
  local cur prev
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"

  case "$COMP_CWORD" in
    1)
      COMPREPLY=($(compgen -W "user worktree wt clean branch version help $(git help -a 2>/dev/null | grep '^  ' | awk '{print $1}')" -- "$cur"))
      ;;
    2)
      case "$prev" in
        user)
          local profiles=""
          if [[ -f "$HOME/.ge/credentials" ]]; then
            profiles="$(grep '^\[' "$HOME/.ge/credentials" | tr -d '[]')"
          fi
          COMPREPLY=($(compgen -W "list current add set ssh-key migrate help $profiles" -- "$cur"))
          ;;
        worktree|wt)
          COMPREPLY=($(compgen -W "add list ls remove rm prune help" -- "$cur"))
          ;;
      esac
      ;;
  esac
}
complete -F _ge_completion ge
