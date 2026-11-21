// Shell-independent core of git-pick. The shell layers (git-pick.zsh, pwsh/git-pick.ps1) only bind
// keys, run the `git-pick` binary and add what it prints to the command line.
//
// Usage:
//
//	git-pick status --shell <shell> [--exclude]
//	  Pick status files in fzf. `--exclude` leaves out the ones in `.gitpickignore`.
//	git-pick worktrees --shell <shell>
//	  Pick worktree paths in fzf.
//	git-pick head [--map-regex <regex>]
//	  Print the HEAD, or its first match of <regex> (group 1 when the regex has one).
//	git-pick menu --shell <shell> [--exclude] [--map-regex <regex>]
//	  Pick one of the above in fzf, then run it.
//
// <shell> is zsh or pwsh, for quoting the filepaths. Prints nothing when a pick is cancelled.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type opts struct {
	shell    string
	exclude  bool
	mapRegex string
}

// `$HOME` first, as the shells use it for `~`.
func home() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	h, _ := os.UserHomeDir()
	return h
}

// splitLines splits on newlines and drops the trailing empty lines, so `"q\n\nf\n"` is
// `["q", "", "f"]`: the empty line in the middle is fzf's `--expect` key when Enter was pressed.
func splitLines(s string) []string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func removeBlank(items []string) []string {
	var out []string
	for _, s := range items {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

// PROCESSES

// git returns the stdout of `git` with `args`, and false when it fails, e.g. outside a repository.
func git(stdin string, args ...string) (string, bool) {
	cmd := exec.Command("git", args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return string(out), true
}

// fzf returns the stdout of fzf with `args`, reading `input`. fzf draws on the terminal itself.
// A cancelled pick exits non-zero with no output, which is not an error here.
func fzf(input string, args ...string) string {
	cmd := exec.Command("fzf", args...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		fail(err.Error())
	}
	return string(out)
}

// QUOTING

// Go's `\w` is ASCII only, so the Unicode classes are spelled out to leave `ñandú.md` unquoted.
var (
	safeZsh  = regexp.MustCompile(`^[\p{L}\p{M}\p{Nd}\p{Pc}./~@%+=:,-]+$`)
	safePwsh = regexp.MustCompile(`^[\p{L}\p{M}\p{Nd}\p{Pc}./\\:-]+$`)
)

// quoteZsh quotes `s` for the zsh buffer only when needed, in single quotes, which also keep `!`
// from being history expansion.
func quoteZsh(s string) string {
	if safeZsh.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func quotePwsh(s string) string {
	if safePwsh.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func quoteFor(shell string) func(string) string {
	if shell == "pwsh" {
		return quotePwsh
	}
	return quoteZsh
}

// HEAD

// mapBranch returns the first match of `mapRegex` in `branch` (group 1 when the regex has one),
// or empty when it does not match. `branch` as is when there is no regex.
func mapBranch(branch, mapRegex string) (string, error) {
	if mapRegex == "" || branch == "" {
		return branch, nil
	}
	re, err := regexp.Compile(mapRegex)
	if err != nil {
		return "", err
	}
	m := re.FindStringSubmatchIndex(branch)
	switch {
	case m == nil:
		return "", nil
	case len(m) > 2 && m[2] >= 0:
		return branch[m[2]:m[3]], nil
	default:
		return branch[m[0]:m[1]], nil
	}
}

// head returns the branch, or empty when detached.
func head(mapRegex string) (string, error) {
	out, _ := git("", "rev-parse", "--abbrev-ref", "HEAD")
	branch := strings.TrimSpace(out)
	if branch == "HEAD" {
		branch = ""
	}
	return mapBranch(branch, mapRegex)
}

// WORKTREES

// worktreePaths returns the paths in the output of `git worktree list --porcelain -z`. NUL
// delimited end to end: `--porcelain` prints paths raw, so a path with a newline would otherwise
// split over two lines.
func worktreePaths(out string) []string {
	var paths []string
	for _, field := range strings.Split(out, "\x00") {
		if path, ok := strings.CutPrefix(field, "worktree "); ok {
			paths = append(paths, path)
		}
	}
	return paths
}

func worktrees(shell string) string {
	out, _ := git("", "worktree", "list", "--porcelain", "-z")
	paths := worktreePaths(out)
	if len(paths) == 0 {
		return ""
	}
	picked := fzf(strings.Join(paths, "\x00"),
		"--multi", "--read0", "--print0", "--bind", "ctrl-a:toggle-all")
	return quoteAll(removeBlank(strings.Split(picked, "\x00")), quoteFor(shell))
}

func quoteAll(items []string, quote func(string) string) string {
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = quote(s)
	}
	return strings.Join(quoted, " ")
}

// STATUS

var ansi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// unquoteGit decodes a filepath git wrapped in double quotes: C-style escapes and octal per byte
// for the non-ASCII characters, see `core.quotePath` in git-config(1).
func unquoteGit(s string) string {
	escapes := map[byte]byte{
		'a': 7, 'b': 8, 't': 9, 'n': 10, 'v': 11, 'f': 12, 'r': 13, '"': '"', '\\': '\\',
	}
	var out []byte
	for i := 1; i < len(s)-1; {
		switch {
		case s[i] != '\\':
			out = append(out, s[i])
			i++
		case s[i+1] >= '0' && s[i+1] <= '7':
			b, _ := strconv.ParseUint(s[i+1:i+4], 8, 8)
			out = append(out, byte(b))
			i += 4
		default:
			out = append(out, escapes[s[i+1]])
			i += 2
		}
	}
	return string(out)
}

// statusFilepath returns the filepath of a `git status --short` line, as it is on disk. For a
// rename or copy, the destination.
func statusFilepath(line string) string {
	line = ansi.ReplaceAllString(line, "")
	path := line[3:]
	if line[0] == 'R' || line[0] == 'C' {
		path = path[strings.LastIndex(path, " -> ")+4:]
	}
	if len(path) >= 2 && strings.HasPrefix(path, `"`) && strings.HasSuffix(path, `"`) {
		return unquoteGit(path)
	}
	return path
}

// lexicalNormalize resolves `.` and `..` segments without touching the filesystem: `a/b/../c` is
// `a/c`, and `../c` is `c`, unlike `path.Clean`.
func lexicalNormalize(path string) string {
	var segs []string
	for _, seg := range strings.Split(path, "/") {
		switch seg {
		case "", ".":
		case "..":
			if len(segs) > 0 {
				segs = segs[:len(segs)-1]
			}
		default:
			segs = append(segs, seg)
		}
	}
	return strings.Join(segs, "/")
}

// canonicalize resolves symlinks when `path` exists, e.g. macOS's `/var` to `/private/var`.
func canonicalize(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		return real
	}
	return abs
}

// excludeFile returns the closest `.gitpickignore`, from the cwd up to `$HOME` when the cwd is
// under it, else up to the root. Files higher up are not merged in.
func excludeFile() string {
	stop := canonicalize(home())
	for dir := canonicalize("."); ; {
		file := filepath.Join(dir, ".gitpickignore")
		if f, err := os.Open(file); err == nil {
			f.Close()
			return file
		}
		parent := filepath.Dir(dir)
		if dir == stop || parent == dir {
			return ""
		}
		dir = parent
	}
}

// matchRepo returns an empty repository for `git check-ignore`, created once, or empty when it
// cannot be created.
//
// In the repository being worked on, a `.gitignore` match takes precedence over
// `core.excludesFile`, so a filepath which is tracked and gitignored could not be excluded. In an
// empty repository the exclusions file is the only exclude source. `init.templateDir` is emptied
// so the template brings no `info/exclude`.
func matchRepo() string {
	cache := os.Getenv("XDG_CACHE_HOME")
	if cache == "" {
		cache = filepath.Join(home(), ".cache")
	}
	dir := filepath.Join(cache, "git-pick", "match_repo")
	dotGit := filepath.Join(dir, ".git")
	if _, err := os.Stat(dotGit); err != nil {
		git("", "-c", "init.templateDir=", "init", "-q", dir)
	}
	if _, err := os.Stat(dotGit); err != nil {
		return ""
	}
	return dir
}

// excludedPatterns maps each excluded filepath to the pattern of `file` which excluded it. `-v`
// also reports negated patterns (`!a.log`), which mean not excluded, so those are dropped.
// `--no-index` so tracked filepaths are checked too.
func excludedPatterns(file, repo string, filepaths []string) map[string]string {
	if repo == "" {
		repo = "."
	}
	out, _ := git(strings.Join(filepaths, "\x00")+"\x00",
		"-C", repo, "-c", "core.excludesFile="+file,
		"check-ignore", "-z", "-v", "--no-index", "--stdin")
	fields := strings.Split(out, "\x00")
	excluded := map[string]string{}
	for i := 0; i+4 <= len(fields); i += 4 {
		source, pattern, path := fields[i], fields[i+2], fields[i+3]
		// Without the match repository, `.gitignore` matches are reported too.
		if canonicalize(source) == canonicalize(file) && !strings.HasPrefix(pattern, "!") {
			excluded[path] = pattern
		}
	}
	return excluded
}

func statusLines() []string {
	out, _ := git("", "-c", "color.status=always", "status", "-su")
	return removeBlank(splitLines(out))
}

// splitExcluded returns the status lines `.gitpickignore` does not exclude, and a header line per
// pattern which excluded some, with how many.
func splitExcluded(lines []string) ([]string, string) {
	file := excludeFile()
	if file == "" {
		return lines, ""
	}
	repo := matchRepo()
	// `git status` prints filepaths relative to the cwd, the match repository needs them relative
	// to the root. `--show-prefix` is the cwd from the root.
	prefix := ""
	if repo != "" {
		out, _ := git("", "rev-parse", "--show-prefix")
		prefix = strings.TrimSpace(out)
	}
	filepaths := make([]string, len(lines))
	for i, line := range lines {
		filepaths[i] = statusFilepath(line)
		if prefix != "" {
			filepaths[i] = lexicalNormalize(prefix + filepaths[i])
		}
	}
	excluded := excludedPatterns(file, repo, filepaths)
	var included, patterns []string
	counts := map[string]int{}
	for i, line := range lines {
		pattern, ok := excluded[filepaths[i]]
		if !ok {
			included = append(included, line)
			continue
		}
		if counts[pattern] == 0 {
			patterns = append(patterns, pattern)
		}
		counts[pattern]++
	}
	header := make([]string, len(patterns))
	for i, p := range patterns {
		header[i] = fmt.Sprintf("%s  (%d)", p, counts[p])
	}
	return included, strings.Join(header, "\n")
}

func status(shell string, exclude bool) string {
	lines := statusLines()
	if len(lines) == 0 {
		return ""
	}
	included, header := lines, ""
	if exclude {
		included, header = splitExcluded(lines)
	}
	base := []string{"--multi", "--ansi", "--bind", "ctrl-a:toggle-all"}
	// Ctrl-f lists the full status. fzf hands the key back instead of running a command, so this
	// works the same on every OS.
	args := append(append([]string{}, base...), "--print-query", "--expect", "ctrl-f")
	if header != "" {
		args = append(args, "--header", header)
	}
	out := append(splitLines(fzf(strings.Join(included, "\n"), args...)), "", "")
	query, key, picked := out[0], out[1], out[2:]
	if key == "ctrl-f" {
		picked = splitLines(fzf(strings.Join(lines, "\n"), append(base, "--query", query)...))
	}
	picked = removeBlank(picked)
	paths := make([]string, len(picked))
	for i, line := range picked {
		paths[i] = statusFilepath(line)
	}
	return quoteAll(paths, quoteFor(shell))
}

// MENU

func menu(o opts) (string, error) {
	pick := strings.TrimSpace(
		fzf("s -- status\nw -- worktrees\nh -- HEAD", "--query=^", "--bind", "one:accept"))
	switch {
	case strings.HasPrefix(pick, "s"):
		return status(o.shell, o.exclude), nil
	case strings.HasPrefix(pick, "w"):
		return worktrees(o.shell), nil
	case strings.HasPrefix(pick, "h"):
		return head(o.mapRegex)
	}
	return "", nil
}

// MAIN

func parseOpts(args []string) (opts, error) {
	var o opts
	next := func(i int) string {
		if i+1 < len(args) {
			return args[i+1]
		}
		return ""
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--shell":
			o.shell = next(i)
			i++
		case "--exclude":
			o.exclude = true
		case "--map-regex":
			o.mapRegex = next(i)
			i++
		default:
			return o, fmt.Errorf("Unknown option: %s", args[i])
		}
	}
	return o, nil
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "ERROR: "+msg)
	os.Exit(1)
}

func main() {
	cmd, args := "", []string{}
	if len(os.Args) > 1 {
		cmd, args = os.Args[1], os.Args[2:]
	}
	o, err := parseOpts(args)
	if err != nil {
		fail(err.Error())
	}
	switch cmd {
	case "status", "worktrees", "menu":
		if o.shell != "zsh" && o.shell != "pwsh" {
			fail("Usage: git-pick " + cmd + " --shell <zsh|pwsh>")
		}
	}
	var out string
	switch cmd {
	case "status":
		out = status(o.shell, o.exclude)
	case "worktrees":
		out = worktrees(o.shell)
	case "head":
		out, err = head(o.mapRegex)
	case "menu":
		out, err = menu(o)
	default:
		fail("Usage: git-pick <status|worktrees|head|menu>")
	}
	if err != nil {
		fail(err.Error())
	}
	fmt.Print(out)
}
