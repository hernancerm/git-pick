# git-pick advanced usage

## Transforming HEAD with `gp-head-map`

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

## Excluding files from `gp-status` (<kbd>Ctrl-g</kbd><kbd>Ctrl-s</kbd>)

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
<kbd>Ctrl-g</kbd><kbd>Ctrl-s</kbd> the closest such file is looked up: the lookup starts at
the cwd and walks up the parent dirs, stopping at `${HOME}` or `/`. The first file found
wins; files higher up are not merged in. When no file is found nothing is excluded.

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
