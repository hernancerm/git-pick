# Do not source this script multiple times.
command -v zg_version > /dev/null && return

# CONFIGURATION

function zg_version {
  echo '0.1.2-SNAPSHOT'
}

# Set by git.plugin.zsh. The fallback covers sourcing this file directly.
ZSH_GIT_PATH="${ZSH_GIT_PATH:-${${(%):-%x}:A:h}}"
ZG_PREFIX="${ZG_PREFIX:-^g}"
ZG_SET_KEYBINDS="${ZG_SET_KEYBINDS:-1}"
ZG_STATUS_EXCLUDE="${ZG_STATUS_EXCLUDE:-0}"
ZG_HEAD_MAP_REGEX="${ZG_HEAD_MAP_REGEX:-}"

# CORE

# The git calls, fzf and quoting live in the core, shared with the PowerShell layer.

## @param $@ Subcommand of the core.
## @stdout Text for the zsh buffer.
function _zg_core {
  local args=("${@}")
  if [[ "${1}" != head ]]; then
    args+=(--shell zsh)
  fi
  if [[ ZG_STATUS_EXCLUDE -eq 1 && "${1}" = (status|menu) ]]; then
    args+=(--exclude)
  fi
  bb "${ZSH_GIT_PATH}/core/git.clj" "${args[@]}"
}

# WIDGETS

function zg-menu {
  # Fixes fzf process 2 hiding zsh prompt.
  zle -I
  local args=(menu)
  if [[ -n "${ZG_HEAD_MAP_REGEX}" ]]; then
    args+=(--map-regex "${ZG_HEAD_MAP_REGEX}")
  fi
  LBUFFER+="$(_zg_core "${args[@]}")"
  zle .reset-prompt
}

function zg-status {
  # Fixes fzf process 2 hiding zsh prompt.
  zle -I
  LBUFFER+="$(_zg_core status)"
  zle .reset-prompt
}

function zg-worktrees {
  LBUFFER+="$(_zg_core worktrees)"
  zle .reset-prompt
}

function zg-head-map {
  LBUFFER+="$(_zg_core head --map-regex "${ZG_HEAD_MAP_REGEX}")"
  zle .reset-prompt
}

function zg-head {
  LBUFFER+="$(_zg_core head)"
  zle .reset-prompt
}

zle -N zg-menu
zle -N zg-status
zle -N zg-worktrees
zle -N zg-head-map
zle -N zg-head

# Set keybinds.

if [[ ZG_SET_KEYBINDS -eq 1 ]]; then
  bindkey "${ZG_PREFIX}^s" zg-status
  bindkey "${ZG_PREFIX}^w" zg-worktrees
  bindkey "${ZG_PREFIX}^h" zg-head
fi
