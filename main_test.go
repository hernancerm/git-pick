package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func check[T any](t *testing.T, got, want T) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestQuoteZsh(t *testing.T) {
	// Safe filepaths are left as is.
	check(t, quoteZsh("src/a-b_c.go"), "src/a-b_c.go")
	check(t, quoteZsh("ñandú.md"), "ñandú.md")
	// Others are single quoted.
	check(t, quoteZsh("a b.txt"), "'a b.txt'")
	check(t, quoteZsh("a!b"), "'a!b'")
	check(t, quoteZsh("*.log"), "'*.log'")
	check(t, quoteZsh("it's"), `'it'\''s'`)
}

func TestQuotePwsh(t *testing.T) {
	check(t, quotePwsh(`C:\Users\a.txt`), `C:\Users\a.txt`)
	check(t, quotePwsh("a b.txt"), "'a b.txt'")
	check(t, quotePwsh("it's"), "'it''s'")
	check(t, quotePwsh("$a"), "'$a'")
}

func TestMapBranch(t *testing.T) {
	cases := []struct{ branch, regex, want string }{
		{"ABC-123-feat", "", "ABC-123-feat"},           // No regex keeps the branch.
		{"feat/ABC-123-x", "[A-Z]+-[0-9]+", "ABC-123"}, // Whole match without groups.
		{"ABC-123-x", "[A-Z]+-([0-9]+)", "123"},        // Group 1 when the regex has one.
		{"main", "[A-Z]+-[0-9]+", ""},                  // Empty when it does not match.
		{"", "[A-Z]+-[0-9]+", ""},                      // Detached HEAD stays empty.
	}
	for _, c := range cases {
		got, err := mapBranch(c.branch, c.regex)
		check(t, err, nil)
		check(t, got, c.want)
	}
	_, err := mapBranch("main", "(")
	if err == nil {
		t.Error("want an error for an invalid regex")
	}
}

func TestWorktreePaths(t *testing.T) {
	check(t, worktreePaths("worktree /repo\x00HEAD abc\x00branch refs/heads/main\x00\x00"+
		"worktree /repo/wt with\nnewline\x00HEAD def\x00detached\x00\x00"),
		[]string{"/repo", "/repo/wt with\nnewline"})
	check(t, worktreePaths(""), nil)
}

func TestUnquoteGit(t *testing.T) {
	check(t, unquoteGit(`"a\"b\\c\td"`), "a\"b\\c\td")           // C-style escapes.
	check(t, unquoteGit(`"\303\261and\303\272.md"`), "ñandú.md") // Octal bytes decode as UTF-8.
}

func TestStatusFilepath(t *testing.T) {
	check(t, statusFilepath(" M a.txt"), "a.txt")
	check(t, statusFilepath(`?? "a b.txt"`), "a b.txt")
	check(t, statusFilepath("\x1b[32mM\x1b[m  a.txt"), "a.txt") // ANSI colors are stripped.
	// A rename or copy gives the destination.
	check(t, statusFilepath("R  old.txt -> new.txt"), "new.txt")
	check(t, statusFilepath(`R  plain.txt -> "mv\"d.txt"`), `mv"d.txt`)
	check(t, statusFilepath("C  a -> b.txt"), "b.txt")
}

func TestLexicalNormalize(t *testing.T) {
	check(t, lexicalNormalize("a/b/../c"), "a/c")
	check(t, lexicalNormalize("./a//b/."), "a/b")
	check(t, lexicalNormalize("../c"), "c")
}

func TestExcludedPatterns(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	file := filepath.Join(dir, ".gitpickignore")
	if _, ok := git("", "-c", "init.templateDir=", "init", "-q", repo); !ok {
		t.Fatal("git init failed")
	}
	if err := os.WriteFile(file, []byte("# Comment.\n*.log\n!keep.log\nbuild/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	check(t, excludedPatterns(file, repo,
		[]string{"a.log", "src/b.log", "keep.log", "build/x.txt", "c.txt"}),
		map[string]string{"a.log": "*.log", "src/b.log": "*.log", "build/x.txt": "build/"})
}

func TestParseOpts(t *testing.T) {
	o, err := parseOpts([]string{"--shell", "zsh", "--exclude", "--map-regex", "x"})
	check(t, err, nil)
	check(t, o, opts{shell: "zsh", exclude: true, mapRegex: "x"})
	o, err = parseOpts(nil)
	check(t, err, nil)
	check(t, o, opts{})
	_, err = parseOpts([]string{"--nope"})
	check(t, err.Error(), "Unknown option: --nope")
}
