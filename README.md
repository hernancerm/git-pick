# git-pick

zsh (or PowerShell) + [fzf](https://github.com/junegunn/fzf) + git

## Usage

After sourcing the plugin these keybinds are set:

- <kbd>Ctrl-g</kbd><kbd>Ctrl-s</kbd> Add status files to the zsh buffer.
  - <kbd>Ctrl-f</kbd> Includes excluded items (`GP_STATUS_EXCLUDE`) in fzf's list.
  - <kbd>Ctrl-a</kbd> Toggles selecting all items from fzf's list.
  - <kbd>Tab</kbd> Toggle current item selection.
- <kbd>Ctrl-g</kbd><kbd>Ctrl-w</kbd> Add worktree path to the zsh buffer.
  - <kbd>Ctrl-a</kbd> Toggles selecting all items from fzf's list.
  - <kbd>Tab</kbd> Toggle current item selection.
- <kbd>Ctrl-g</kbd><kbd>Ctrl-h</kbd> Add the HEAD to the zsh buffer.

Usage examples:

```text
git add <Ctrl-g><Ctrl-s>
git push -u origin <Ctrl-g><Ctrl-h>
cd <Ctrl-g><Ctrl-w>
```

> [!WARNING]
> <kbd>Ctrl-g</kbd><kbd>Ctrl-s</kbd> does not work if flow control is enabled, `stty
> -ixon` disables flow control.

See also: [ADVANCED_USAGE.md](/ADVANCED_USAGE.md)

## Installation

### Without a plugin manager

1. Install [fzf](https://github.com/junegunn/fzf) and [babashka](https://babashka.org/)
   (`bb`).

2. Clone the git-pick repository by executing the below command:

    ```bash
    git clone 'https://github.com/hernancerm/git-pick.git' \
      "${HOME}/.local/share/git-pick"
    ```

#### Zsh

3. Source the plugin (also add command to your `~/.zshrc` to enable on future sessions):

    ```bash
    source "${HOME}/.local/share/git-pick/git-pick.plugin.zsh"
    ```

    OR: If you want to set custom keybinds then use:

    ```bash
    GP_SET_KEYBINDS=0
    source "${HOME}/.local/share/git-pick/git-pick.plugin.zsh"
    bindkey "^g^s" gp-status
    bindkey "^g^w" gp-worktrees
    bindkey "^g^h" gp-head
    ```

    OR: If you use [jeffreytse/zsh-vi-mode](https://github.com/jeffreytse/zsh-vi-mode):

    ```bash
    GP_SET_KEYBINDS=0
    source "${HOME}/.local/share/git-pick/git-pick.plugin.zsh"
    function zvm_after_init {
      zvm_define_widget gp-menu
      # ZVM does not handle chords well, prefer gp-menu.
      zvm_bindkey viins "^g" gp-menu
    }
    ```

#### PowerShell

3. Source the PowerShell layer from `$PROFILE`:

    ```powershell
    . "${HOME}/.local/share/git-pick/pwsh/git-pick.ps1"
    ```

    OR: If you want to set custom keybinds then use:

    ```powershell
    $GP_SET_KEYBINDS = 0
    . "${HOME}/.local/share/git-pick/pwsh/git-pick.ps1"
    Set-PSReadLineKeyHandler -Chord 'Ctrl+g,Ctrl+s' -ScriptBlock { Invoke-GpStatus }
    Set-PSReadLineKeyHandler -Chord 'Ctrl+g,Ctrl+w' -ScriptBlock { Invoke-GpWorktrees }
    Set-PSReadLineKeyHandler -Chord 'Ctrl+g,Ctrl+h' -ScriptBlock { Invoke-GpHead }
    ```

### With a plugin manager

#### Zsh

With [Sheldon](https://github.com/rossmacarthur/sheldon):

```toml
[plugins.git-pick]
github = "hernancerm/git-pick"

# File: ~/.config/sheldon/plugins.toml
```

```bash
# Source plugins.
eval "$(sheldon source)"

# File: ~/.zshrc
```

## Similar projects

- <https://github.com/junegunn/fzf-git.sh>
