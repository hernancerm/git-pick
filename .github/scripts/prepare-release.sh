#!/usr/bin/env bash

# Opens a release in the working tree: sets the version in git-pick.zsh and turns the
# [Unreleased] section of CHANGELOG.md into a dated section for it. Run from the
# repo root. Commits nothing.

set -euo pipefail

version="${1:?usage: prepare-release.sh <version>}"
repo_url='https://github.com/hernancerm/git-pick'
today="$(date +%F)"

if [[ ! "${version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]
then
  echo "error: version must be MAJOR.MINOR.PATCH, got '${version}'" >&2
  exit 1
fi

# An empty [Unreleased] means there is nothing to release, which is almost
# always a mistake rather than an intent to publish an empty version.
if ! awk '
  /^## \[Unreleased\]/ { on = 1; next }
  on && /^## \[/ { exit }
  on && /^- / { found = 1 }
  END { exit !found }
' CHANGELOG.md
then
  echo 'error: the [Unreleased] section of CHANGELOG.md is empty' >&2
  exit 1
fi

# The newest released version, used as the left side of the compare link.
prev="$(grep -m1 -E '^## \[[0-9]' CHANGELOG.md | sed -E 's/^## \[([^]]+)\].*/\1/')"

awk \
  -v version="${version}" \
  -v today="${today}" \
  -v prev="${prev}" \
  -v url="${repo_url}" \
  '
    $0 == "## [Unreleased]" {
      print "## [Unreleased]"
      print ""
      print "## [" version "] - " today
      next
    }
    /^\[Unreleased\]:/ {
      print "[Unreleased]: " url "/compare/" version "...HEAD"
      print "[" version "]: " url "/compare/" prev "..." version
      next
    }
    { print }
  ' CHANGELOG.md > CHANGELOG.md.tmp

mv CHANGELOG.md.tmp CHANGELOG.md

.github/scripts/set-version.sh "${version}"

echo "prepared ${version} (previous: ${prev})"
