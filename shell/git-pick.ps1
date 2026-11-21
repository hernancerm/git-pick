# PowerShell layer of git-pick, printed by `git-pick init pwsh`. Load it from $PROFILE:
#   Invoke-Expression (& git-pick init pwsh | Out-String)
# Set these before that line, with the same meaning as in zsh (see the README):
#   $GP_SET_KEYBINDS = 0, $GP_STATUS_EXCLUDE = 1, $GP_HEAD_MAP_REGEX = '...'
#   $GP_PREFIX, in PSReadLine chord syntax (see `Get-Help Set-PSReadLineKeyHandler`).

# The git calls, fzf and quoting live in the core, shared with the zsh layer.
function Invoke-GpCore([string[]] $CoreArgs) {
  if ($CoreArgs[0] -ne 'head') {
    $CoreArgs += '--shell', 'pwsh'
  }
  if ($GP_STATUS_EXCLUDE -eq 1 -and $CoreArgs[0] -in 'status', 'menu') {
    $CoreArgs += '--exclude'
  }
  $out = & git-pick @CoreArgs
  # Redraw the prompt, which fzf drew over.
  [Microsoft.PowerShell.PSConsoleReadLine]::InvokePrompt()
  if ($out) {
    [Microsoft.PowerShell.PSConsoleReadLine]::Insert($out)
  }
}

function Invoke-GpMenu {
  $coreArgs = @('menu')
  if ($GP_HEAD_MAP_REGEX) {
    $coreArgs += '--map-regex', $GP_HEAD_MAP_REGEX
  }
  Invoke-GpCore $coreArgs
}

function Invoke-GpStatus {
  Invoke-GpCore @('status')
}

function Invoke-GpWorktrees {
  Invoke-GpCore @('worktrees')
}

function Invoke-GpHeadMap {
  Invoke-GpCore @('head', '--map-regex', $GP_HEAD_MAP_REGEX)
}

function Invoke-GpHead {
  Invoke-GpCore @('head')
}

if ($null -eq $GP_PREFIX) {
  $GP_PREFIX = 'Ctrl+g'
}
if ($GP_SET_KEYBINDS -ne 0) {
  Set-PSReadLineKeyHandler -Chord "$GP_PREFIX,Ctrl+s" -ScriptBlock { Invoke-GpStatus }
  Set-PSReadLineKeyHandler -Chord "$GP_PREFIX,Ctrl+w" -ScriptBlock { Invoke-GpWorktrees }
  Set-PSReadLineKeyHandler -Chord "$GP_PREFIX,Ctrl+h" -ScriptBlock { Invoke-GpHead }
}
