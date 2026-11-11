# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-11-11

### Added

- Prebuilt binaries for macOS arm64, Linux arm64 and x86_64, and Windows x86_64,
  published on release and installable with mise.
- `GP_CMD` to run the core from source with babashka.
- PowerShell support.
- Worktrees picker.
- Widgets for each action, so each can be bound to its own key.
- HEAD mapping with a regex (`GP_HEAD_MAP_REGEX`), e.g. to insert only the ticket ID of a
  branch.
- Status exclusions from a `.gitpickignore` file, in gitignore syntax, with the number of
  matches per pattern in fzf's header. `Ctrl-f` lists the full status.
- `Ctrl-a` toggles selecting all items in fzf's list.

### Changed

- Renamed the project from `git` to `git-pick`, and its widgets and variables to the `gp`
  prefix.
- The core is a native binary, so babashka is no longer needed.

### Fixed

- Worktree paths are quoted for the zsh buffer.
- `argument list too long` error with a large `git status`.
- The zsh prompt disappearing after picking, also with zsh-vi-mode.
- A file both staged as renamed and modified unstaged.

## [0.1.1] - 2026-01-11

### Fixed

- Globbing characters in status filepaths are escaped.

## [0.1.0] - 2025-03-23

### Added

- Initial release.

[Unreleased]: https://github.com/hernancerm/git-pick/compare/0.2.0...HEAD
[0.2.0]: https://github.com/hernancerm/git-pick/compare/0.1.1...0.2.0
[0.1.1]: https://github.com/hernancerm/git-pick/compare/0.1.0...0.1.1
[0.1.0]: https://github.com/hernancerm/git-pick/releases/tag/0.1.0
