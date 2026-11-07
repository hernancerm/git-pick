# zsh-git

zsh (or PowerShell) + [fzf](https://github.com/junegunn/fzf) + git

## Usage

After sourcing the plugin these keybinds are set:

- <kbd>Ctrl-g</kbd><kbd>Ctrl-s</kbd> Add status files to the zsh buffer.
  - <kbd>Ctrl-f</kbd> Includes excluded items (`ZG_STATUS_EXCLUDE`) in fzf's list.
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

### Transforming HEAD with `zg-head-map`

Set `ZG_HEAD_MAP_REGEX` and use `zg-head-map` to put only the HEAD's first match of the
regex in the zsh buffer (its group 1 when it has one, else the whole match). Nothing is put
when it does not match. The regex is Java's syntax. Example:

```bash
# Jira ticket key, e.g., ABC-123.
ZG_HEAD_MAP_REGEX='[A-Z]+-[0-9]+'
source "${HOME}/.zsh-git/zsh-git/git.plugin.zsh"
bindkey "^g^k" zg-head-map

# Use case:
# Prefix the ticket key to each commit message:
# $ git commit -m '<Ctrl-g><Ctrl-k>

# File: ~/.zshrc
```

In PowerShell, set the variable before dot-sourcing and bind `Invoke-ZgHeadMap`:

```powershell
$ZG_HEAD_MAP_REGEX = '[A-Z]+-[0-9]+'
. "${HOME}/.zsh-git/zsh-git/pwsh/git.ps1"
Set-PSReadLineKeyHandler -Chord 'Ctrl+g,Ctrl+k' -ScriptBlock { Invoke-ZgHeadMap }

# File: $PROFILE
```

### Excluding files from `zg-status` (<kbd>Ctrl-g</kbd><kbd>Ctrl-s</kbd>)

Set `ZG_STATUS_EXCLUDE=1` before sourcing the plugin so
<kbd>Ctrl-g</kbd><kbd>Ctrl-s</kbd> leaves specific files out of fzf's list:

```bash
ZG_STATUS_EXCLUDE=1
source "${HOME}/.zsh-git/zsh-git/git.plugin.zsh"

# Use case:
# Exclude from `git add` files modified for app startup purposes.
# $ git add <Ctrl-g><Ctrl-s>

# File: ~/.zshrc
```

In PowerShell, set `$ZG_STATUS_EXCLUDE = 1` before dot-sourcing `pwsh/git.ps1`.

The exclusions are defined per directory, in a `.zg_status_exclude` file. On each
<kbd>Ctrl-g</kbd><kbd>Ctrl-s</kbd> the closest such file is looked up: the lookup starts at
the cwd and walks up the parent dirs, stopping at `${HOME}` or `/`. The first file found
wins; files higher up are not merged in. When no file is found nothing is excluded.

The file is in [gitignore](https://git-scm.com/docs/gitignore) syntax, so its comment,
escaping and pattern rules are the ones from `.gitignore`. Example contents of a
`.zg_status_exclude` file placed at the root of a repository:

```gitignore
# Files modified for app startup purposes.
Makefile
compose.yaml
src/main/resources/application-*.yml
!src/main/resources/application-prod.yml
src/main/resources/db/migration/

# File: ./.zg_status_exclude
```

Only the patterns which excluded at least one file are listed in fzf's header, each
followed by how many files it excluded, as in `src/main/resources/application-*.yml  (2)`.
Press <kbd>Ctrl-f</kbd> to list the full status.

The matching runs in an empty repository cached at `${XDG_CACHE_HOME:-~/.cache}/zsh-git`,
created on first use, so that `.zg_status_exclude` is the only exclude source Git
considers. Otherwise the repo's `.gitignore` would take precedence over it, leaving a
filepath which is tracked _and_ gitignored impossible to exclude.

Add `.zg_status_exclude` to your global `.gitignore`.

## Installation

### Without a plugin manager

1. Install [fzf](https://github.com/junegunn/fzf) and [babashka](https://babashka.org/)
   (`bb`).

2. Clone the zsh-git repository by executing the below command:

    ```bash
    git clone 'https://github.com/hernancerm/zsh-git.git' \
      "${HOME}/.zsh-git/zsh-git"
    ```

#### Zsh

3. Source the plugin (also add command to your `~/.zshrc` to enable on future sessions):

    ```bash
    source "${HOME}/.zsh-git/zsh-git/git.plugin.zsh"
    ```

    OR: If you want to set custom keybinds then use:

    ```bash
    ZG_SET_KEYBINDS=0
    source "${HOME}/.zsh-git/zsh-git/git.plugin.zsh"
    bindkey "^g^s" zg-status
    bindkey "^g^w" zg-worktrees
    bindkey "^g^h" zg-head
    ```

    OR: If you use [jeffreytse/zsh-vi-mode](https://github.com/jeffreytse/zsh-vi-mode):

    ```bash
    ZG_SET_KEYBINDS=0
    source "${HOME}/.zsh-git/zsh-git/git.plugin.zsh"
    function zvm_after_init {
      zvm_define_widget zg-menu
      # ZVM does not handle chords well, prefer zg-menu.
      zvm_bindkey viins "^g" zg-menu
    }
    ```

#### PowerShell

3. Source the PowerShell layer from `$PROFILE`:

    ```powershell
    . "${HOME}/.zsh-git/zsh-git/pwsh/git.ps1"
    ```

    OR: If you want to set custom keybinds then use:

    ```powershell
    $ZG_SET_KEYBINDS = 0
    . "${HOME}/.zsh-git/zsh-git/pwsh/git.ps1"
    Set-PSReadLineKeyHandler -Chord 'Ctrl+g,Ctrl+s' -ScriptBlock { Invoke-ZgStatus }
    Set-PSReadLineKeyHandler -Chord 'Ctrl+g,Ctrl+w' -ScriptBlock { Invoke-ZgWorktrees }
    Set-PSReadLineKeyHandler -Chord 'Ctrl+g,Ctrl+h' -ScriptBlock { Invoke-ZgHead }
    ```

### With a plugin manager

#### Zsh

With [Sheldon](https://github.com/rossmacarthur/sheldon):

```toml
[plugins.zsh-git]
github = "hernancerm/zsh-git"

# File: ~/.config/sheldon/plugins.toml
```

```bash
# Source plugins.
eval "$(sheldon source)"

# File: ~/.zshrc
```

## Similar projects

- <https://github.com/junegunn/fzf-git.sh>
