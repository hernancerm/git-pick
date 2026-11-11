# git-pick

zsh (or PowerShell) + [fzf](https://github.com/junegunn/fzf) + git

## Usage

After sourcing the plugin these keybinds are set:

- <kbd>Ctrl-g</kbd><kbd>Ctrl-s</kbd> Add status files to the command line.
  - <kbd>Ctrl-f</kbd> Includes excluded items (`GP_STATUS_EXCLUDE`) in fzf's list.
  - <kbd>Ctrl-a</kbd> Toggles selecting all items from fzf's list.
  - <kbd>Tab</kbd> Toggle current item selection.
- <kbd>Ctrl-g</kbd><kbd>Ctrl-w</kbd> Add worktree path to the command line.
  - <kbd>Ctrl-a</kbd> Toggles selecting all items from fzf's list.
  - <kbd>Tab</kbd> Toggle current item selection.
- <kbd>Ctrl-g</kbd><kbd>Ctrl-h</kbd> Add the HEAD to the command line.

Usage examples:

```text
git add <Ctrl-g><Ctrl-s>
git push -u origin <Ctrl-g><Ctrl-h>
cd <Ctrl-g><Ctrl-w>
```

> [!WARNING]
> <kbd>Ctrl-g</kbd><kbd>Ctrl-s</kbd> does not work if flow control is enabled, `stty
> -ixon` disables flow control.

## Advanced usage

<details>

<summary>Excluding files from <code>gp-status</code> (<kbd>Ctrl-g</kbd><kbd>Ctrl-s</kbd>)</summary>

Set `GP_STATUS_EXCLUDE=1` before sourcing the plugin so
<kbd>Ctrl-g</kbd><kbd>Ctrl-s</kbd> leaves specific files out of fzf's list:

```bash
GP_STATUS_EXCLUDE=1
source "${HOME}/.local/share/git-pick/git-pick.plugin.zsh"

# Use case:
# Exclude from `git add` files modified for app startup purposes.
# $ git add <Ctrl-g><Ctrl-s>

# File: ~/.zshrc
```

In PowerShell, set `$GP_STATUS_EXCLUDE = 1` before dot-sourcing `pwsh/git-pick.ps1`.

The exclusions are defined per directory, in a `.gitpickignore` file. On each
<kbd>Ctrl-g</kbd><kbd>Ctrl-s</kbd> the closest such file is looked up: the lookup starts
at the cwd and walks up the parent dirs, stopping at `${HOME}` or `/`. The first file
found wins; files higher up are not merged in. When no file is found nothing is excluded.

The file is in [gitignore](https://git-scm.com/docs/gitignore) syntax, so its comment,
escaping and pattern rules are the ones from `.gitignore`. Example contents of a
`.gitpickignore` file placed at the root of a repository:

```gitignore
# Files modified for app startup purposes.
Makefile
compose.yaml
src/main/resources/application-*.yml
!src/main/resources/application-prod.yml
src/main/resources/db/migration/

# File: ./.gitpickignore
```

Only the patterns which excluded at least one file are listed in fzf's header, each
followed by how many files it excluded, as in `src/main/resources/application-*.yml  (2)`.
Press <kbd>Ctrl-f</kbd> to list the full status.

The matching runs in an empty repository cached at `${XDG_CACHE_HOME:-~/.cache}/git-pick`,
created on first use, so that `.gitpickignore` is the only exclude source Git
considers. Otherwise the repo's `.gitignore` would take precedence over it, leaving a
filepath which is tracked _and_ gitignored impossible to exclude.

Add `.gitpickignore` to your global `.gitignore`.

</details>

<details>

<summary>Transforming HEAD with <code>gp-head-map</code></summary>

Set `GP_HEAD_MAP_REGEX` and use `gp-head-map` to put only the HEAD's first match of the
regex in the zsh buffer (its group 1 when it has one, else the whole match). Nothing is put
when it does not match. The regex is Java's syntax. Example:

```bash
# Jira ticket key, e.g., ABC-123.
GP_HEAD_MAP_REGEX='[A-Z]+-[0-9]+'
source "${HOME}/.local/share/git-pick/git-pick.plugin.zsh"
bindkey "^g^k" gp-head-map

# Use case:
# Prefix the ticket key to each commit message:
# $ git commit -m '<Ctrl-g><Ctrl-k>

# File: ~/.zshrc
```

In PowerShell, set the variable before dot-sourcing and bind `Invoke-GpHeadMap`:

```powershell
$GP_HEAD_MAP_REGEX = '[A-Z]+-[0-9]+'
. "${HOME}/.local/share/git-pick/pwsh/git-pick.ps1"
Set-PSReadLineKeyHandler -Chord 'Ctrl+g,Ctrl+k' -ScriptBlock { Invoke-GpHeadMap }

# File: $PROFILE
```

</details>

## Installation

First, install [fzf](https://github.com/junegunn/fzf) and the `git-pick` binary. With
[mise](https://mise.jdx.dev/):

```bash
mise use -g github:hernancerm/git-pick
```

OR: Download it from [releases](https://github.com/hernancerm/git-pick/releases) and add
to `$PATH`.

### Zsh

With [Sheldon](https://github.com/rossmacarthur/sheldon) (Zsh plugin manager):

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

Alternatively, if not using Sheldon:

1. Clone the git-pick repository by executing the below command:

    ```bash
    git clone 'https://github.com/hernancerm/git-pick.git' \
      "${HOME}/.local/share/git-pick"
    ```

2. Source the plugin (also add the command to your `~/.zshrc` to enable on future
   sessions):

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

### PowerShell

1. Clone the git-pick repository by executing the below command:

    ```powershell
    git clone 'https://github.com/hernancerm/git-pick.git' "${HOME}/.local/share/git-pick"
    ```

2. Source the PowerShell layer from `$PROFILE`:

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

## Development

To run [src/git_pick/core.clj](/src/git_pick/core.clj) from source with
[babashka](https://babashka.org/) instead of the binary, set this before sourcing the
plugin:

- Zsh: `GP_CMD=(bb ~/dev/repos/git-pick/src/git_pick/core.clj)`
- PowerShell: `$GP_CMD = 'bb', "$HOME/dev/repos/git-pick/src/git_pick/core.clj"`

The core must run on both bb and the JVM. `bb test` runs the tests on both. Build the
binary with `bb bin`.

## Similar projects

- <https://github.com/junegunn/fzf-git.sh>
