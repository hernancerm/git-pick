# Do not source this script multiple times.
command -v gp_version > /dev/null && return

# CONFIGURATION

function gp_version {
  echo '0.2.1-SNAPSHOT'
}

GP_PREFIX="${GP_PREFIX:-^g}"
GP_SET_KEYBINDS="${GP_SET_KEYBINDS:-1}"
GP_STATUS_EXCLUDE="${GP_STATUS_EXCLUDE:-0}"
GP_HEAD_MAP_REGEX="${GP_HEAD_MAP_REGEX:-}"
# Command running the core. Set it to run from source with bb, e.g.:
#   GP_CMD=(bb ~/dev/repos/git-pick/src/git_pick/core.clj)
if (( ! ${+GP_CMD} )); then
  GP_CMD=(git-pick)
fi

# CORE

# The git calls, fzf and quoting live in the core, shared with the PowerShell layer.

## @param $@ Subcommand of the core.
## @stdout Text for the zsh buffer.
function _gp_core {
  local args=("${@}")
  if [[ "${1}" != head ]]; then
    args+=(--shell zsh)
  fi
  if [[ GP_STATUS_EXCLUDE -eq 1 && "${1}" = (status|menu) ]]; then
    args+=(--exclude)
  fi
  "${GP_CMD[@]}" "${args[@]}"
}

# WIDGETS

function gp-menu {
  # Fixes fzf process 2 hiding zsh prompt.
  zle -I
  local args=(menu)
  if [[ -n "${GP_HEAD_MAP_REGEX}" ]]; then
    args+=(--map-regex "${GP_HEAD_MAP_REGEX}")
  fi
  LBUFFER+="$(_gp_core "${args[@]}")"
  zle .reset-prompt
}

function gp-status {
  # Fixes fzf process 2 hiding zsh prompt.
  zle -I
  LBUFFER+="$(_gp_core status)"
  zle .reset-prompt
}

function gp-worktrees {
  LBUFFER+="$(_gp_core worktrees)"
  zle .reset-prompt
}

function gp-head-map {
  LBUFFER+="$(_gp_core head --map-regex "${GP_HEAD_MAP_REGEX}")"
  zle .reset-prompt
}

function gp-head {
  LBUFFER+="$(_gp_core head)"
  zle .reset-prompt
}

zle -N gp-menu
zle -N gp-status
zle -N gp-worktrees
zle -N gp-head-map
zle -N gp-head

# Set keybinds.

if [[ GP_SET_KEYBINDS -eq 1 ]]; then
  bindkey "${GP_PREFIX}^s" gp-status
  bindkey "${GP_PREFIX}^w" gp-worktrees
  bindkey "${GP_PREFIX}^h" gp-head
fi
