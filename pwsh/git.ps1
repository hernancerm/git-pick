# PowerShell layer of zsh-git. Dot-source it from $PROFILE:
#   . ~/dev/repos/zsh-git/pwsh/git.ps1
# Set these before dot-sourcing, with the same meaning as in zsh (see the README):
#   $ZG_SET_KEYBINDS = 0, $ZG_STATUS_EXCLUDE = 1, $ZG_HEAD_MAP_REGEX = '...'
#   $ZG_PREFIX, in PSReadLine chord syntax (see `Get-Help Set-PSReadLineKeyHandler`).

$script:ZgCore = Join-Path $PSScriptRoot '../core/git.clj'

# The git calls, fzf and quoting live in the core, shared with the zsh layer.
function Invoke-ZgCore([string[]] $CoreArgs) {
  if ($CoreArgs[0] -ne 'head') {
    $CoreArgs += '--shell', 'pwsh'
  }
  if ($ZG_STATUS_EXCLUDE -eq 1 -and $CoreArgs[0] -in 'status', 'menu') {
    $CoreArgs += '--exclude'
  }
  $out = bb $script:ZgCore @CoreArgs
  # Redraw the prompt, which fzf drew over.
  [Microsoft.PowerShell.PSConsoleReadLine]::InvokePrompt()
  if ($out) {
    [Microsoft.PowerShell.PSConsoleReadLine]::Insert($out)
  }
}

function Invoke-ZgMenu {
  $coreArgs = @('menu')
  if ($ZG_HEAD_MAP_REGEX) {
    $coreArgs += '--map-regex', $ZG_HEAD_MAP_REGEX
  }
  Invoke-ZgCore $coreArgs
}

function Invoke-ZgStatus {
  Invoke-ZgCore @('status')
}

function Invoke-ZgWorktrees {
  Invoke-ZgCore @('worktrees')
}

function Invoke-ZgHeadMap {
  Invoke-ZgCore @('head', '--map-regex', $ZG_HEAD_MAP_REGEX)
}

function Invoke-ZgHead {
  Invoke-ZgCore @('head')
}

if ($null -eq $ZG_PREFIX) {
  $ZG_PREFIX = 'Ctrl+g'
}
if ($ZG_SET_KEYBINDS -ne 0) {
  Set-PSReadLineKeyHandler -Chord "$ZG_PREFIX,Ctrl+s" -ScriptBlock { Invoke-ZgStatus }
  Set-PSReadLineKeyHandler -Chord "$ZG_PREFIX,Ctrl+w" -ScriptBlock { Invoke-ZgWorktrees }
  Set-PSReadLineKeyHandler -Chord "$ZG_PREFIX,Ctrl+h" -ScriptBlock { Invoke-ZgHead }
}
