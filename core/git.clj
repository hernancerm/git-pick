#!/usr/bin/env bb

;; Shell-independent core of zsh-git. The shell layers (git.zsh, pwsh/git.ps1) only bind keys, run
;; this script and add what it prints to the command line.
;;
;; Usage:
;;   bb git.clj status --shell <shell> [--exclude]  Pick status files in fzf. `--exclude` leaves
;;                                                  out the ones in `.zg_status_exclude`.
;;   bb git.clj worktrees --shell <shell>           Pick worktree paths in fzf.
;;   bb git.clj head [--map-regex <regex>]          Print the HEAD, or its first match of <regex>
;;                                                  (group 1 when the regex has one).
;;   bb git.clj menu --shell <shell> [--exclude] [--map-regex <regex>]
;;                                                  Pick one of the above in fzf, then run it.
;; <shell> is zsh or pwsh, for quoting the filepaths. Prints nothing when a pick is cancelled.

(require '[babashka.fs :as fs]
         '[babashka.process :as p]
         '[clojure.string :as str])

#_[:type Shell [:enum "zsh" "pwsh"]]

#_[:type Opts
   [:map [:shell [:maybe Shell]] [:exclude :boolean] [:map-regex [:maybe :string]]]]

;; `$HOME` first, as the shells use it for `~`. Java's home (`fs/home`) ignores it.
(def home #_[:- :string]
  (or (System/getenv "HOME") (str (fs/home))))

;; PROCESSES

(defn git
  "Stdout of `git` with `args`, or nil when it fails, e.g. outside a repository."
  #_[:=> [:cat [:vector :string] [:maybe :string]] [:maybe :string]]
  [args stdin]
  (let [{:keys [exit out]} @(p/process (into ["git"] args)
                                       {:in stdin :out :string :err :string})]
    (when (zero? exit)
      out)))

(defn fzf
  "Stdout of fzf with `args`, reading `input`. fzf draws on the terminal itself."
  #_[:=> [:cat [:vector :string] :string] :string]
  [args input]
  (:out @(p/process (into ["fzf"] args) {:in input :out :string :err :inherit})))

;; QUOTING

(defn quote-zsh
  "Quotes `s` for the zsh buffer only when needed, in single quotes, which also keep `!` from
  being history expansion."
  #_[:=> [:cat :string] :string]
  [s]
  (if (re-matches #"(?U)[\w./~@%+=:,-]+" s)
    s
    (str "'" (str/replace s "'" "'\\''") "'")))

(defn quote-pwsh
  #_[:=> [:cat :string] :string]
  [s]
  (if (re-matches #"(?U)[\w./\\:-]+" s)
    s
    (str "'" (str/replace s "'" "''") "'")))

(defn quote-for
  #_[:=> [:cat Shell] [:=> [:cat :string] :string]]
  [shell]
  (case shell "zsh" quote-zsh "pwsh" quote-pwsh))

;; HEAD

(defn head
  "The branch, or empty when detached."
  #_[:=> [:cat [:maybe :string]] :string]
  [map-regex]
  (let [branch (str/trim (or (git ["rev-parse" "--abbrev-ref" "HEAD"] nil) ""))
        branch (if (= "HEAD" branch) "" branch)]
    (if (and (seq map-regex) (seq branch))
      (let [m (re-find (re-pattern map-regex) branch)]
        (cond
          (nil? m) ""
          (string? m) m
          :else (or (second m) (first m))))
      branch)))

;; WORKTREES

(defn worktrees
  #_[:=> [:cat Shell] :string]
  [shell]
  ;; NUL delimited end to end: `--porcelain` prints paths raw, so a path with a newline would
  ;; otherwise split over two lines.
  (let [paths (->> (str/split (or (git ["worktree" "list" "--porcelain" "-z"] nil) "") #"\x00")
                   (keep #(second (re-matches #"(?s)worktree (.*)" %))))]
    (if (empty? paths)
      ""
      (->> (str/split (fzf ["--multi" "--read0" "--print0" "--bind" "ctrl-a:toggle-all"]
                           (str/join "\u0000" paths))
                      #"\x00")
           (remove str/blank?)
           (map (quote-for shell))
           (str/join " ")))))

;; STATUS

(defn strip-ansi
  #_[:=> [:cat :string] :string]
  [s]
  (str/replace s #"\e\[[0-9;]*m" ""))

(defn unquote-git
  "Decodes a filepath git wrapped in double quotes: C-style escapes and octal per byte for the
  non-ASCII characters, see `core.quotePath` in git-config(1)."
  #_[:=> [:cat :string] :string]
  [s]
  (let [out (java.io.ByteArrayOutputStream.)
        escapes {\a 7 \b 8 \t 9 \n 10 \v 11 \f 12 \r 13 \" 34 \\ 92}]
    (loop [i 1]
      (when (< i (dec (count s)))
        (let [c (.charAt s i)]
          (cond
            (not= c \\)
            (let [bytes (.getBytes (str c) "UTF-8")]
              (.write out bytes 0 (count bytes))
              (recur (inc i)))

            (Character/isDigit (.charAt s (inc i)))
            (do (.write out (Integer/parseInt (subs s (inc i) (+ i 4)) 8))
                (recur (+ i 4)))

            :else
            (do (.write out (int (escapes (.charAt s (inc i)))))
                (recur (+ i 2)))))))
    (String. (.toByteArray out) "UTF-8")))

(defn status-filepath
  "Filepath of a `git status --short` line, as it is on disk. For a rename or copy, the
  destination."
  #_[:=> [:cat :string] :string]
  [line]
  (let [line (strip-ansi line)
        filepath (subs line 3)
        filepath (if (#{\R \C} (first line))
                   (let [i (str/last-index-of filepath " -> ")]
                     (subs filepath (+ i 4)))
                   filepath)]
    (if (and (str/starts-with? filepath "\"") (str/ends-with? filepath "\""))
      (unquote-git filepath)
      filepath)))

(defn lexical-normalize
  "Resolves `.` and `..` segments without touching the filesystem: `a/b/../c` is `a/c`."
  #_[:=> [:cat :string] :string]
  [path]
  (->> (str/split path #"/")
       (reduce (fn [acc seg]
                 (case seg
                   ("" ".") acc
                   ".." (if (seq acc) (pop acc) acc)
                   (conj acc seg)))
               [])
       (str/join "/")))

(defn exclude-file
  "The closest `.zg_status_exclude`, from the cwd up to `$HOME` when the cwd is under it, else up
  to the root. Files higher up are not merged in."
  #_[:=> [:cat] [:maybe :string]]
  []
  (let [stop (fs/canonicalize home)]
    (loop [dir (fs/canonicalize ".")]
      (let [file (fs/path dir ".zg_status_exclude")]
        (cond
          (fs/readable? file) (str file)
          (or (= dir stop) (nil? (fs/parent dir))) nil
          :else (recur (fs/parent dir)))))))

(defn match-repo
  "An empty repository for `git check-ignore`, created once, or nil when it cannot be created.

  In the repository being worked on, a `.gitignore` match takes precedence over
  `core.excludesFile`, so a filepath which is tracked and gitignored could not be excluded. In an
  empty repository the exclusions file is the only exclude source. `init.templateDir` is emptied so
  the template brings no `info/exclude`."
  #_[:=> [:cat] [:maybe :string]]
  []
  (let [dir (str (fs/path (or (System/getenv "XDG_CACHE_HOME") (fs/path home ".cache"))
                          "zsh-git" "match_repo"))]
    (when-not (fs/exists? (fs/path dir ".git"))
      (git ["-c" "init.templateDir=" "init" "-q" dir] nil))
    (when (fs/exists? (fs/path dir ".git"))
      dir)))

(defn excluded-patterns
  "Maps each excluded filepath to the pattern of `file` which excluded it. `-v` also reports
  negated patterns (`!a.log`), which mean not excluded, so those are dropped. `--no-index` so
  tracked filepaths are checked too."
  #_[:=> [:cat :string [:maybe :string] [:sequential :string]] [:map-of :string :string]]
  [file repo filepaths]
  (let [out (git ["-C" (or repo ".") "-c" (str "core.excludesFile=" file)
                  "check-ignore" "-z" "-v" "--no-index" "--stdin"]
                 (str (str/join "\u0000" filepaths) "\u0000"))]
    (->> (str/split (or out "") #"\x00")
         (partition 4)
         (keep (fn [[source _line pattern filepath]]
                 ;; Without the match repository, `.gitignore` matches are reported too.
                 (when (and (= (str (fs/canonicalize source)) (str (fs/canonicalize file)))
                            (not (str/starts-with? pattern "!")))
                   [filepath pattern])))
         (into {}))))

(defn status-lines
  #_[:=> [:cat] [:vector :string]]
  []
  (->> (str/split-lines (or (git ["-c" "color.status=always" "status" "-su"] nil) ""))
       (remove str/blank?)
       vec))

(defn split-excluded
  "`[included header]`: the status lines `.zg_status_exclude` does not exclude, and a header line
  per pattern which excluded some, with how many."
  #_[:=> [:cat [:vector :string]] [:tuple [:vector :string] [:maybe :string]]]
  [lines]
  (if-let [file (exclude-file)]
    (let [repo (match-repo)
          ;; `git status` prints filepaths relative to the cwd, the match repository needs them
          ;; relative to the root. `--show-prefix` is the cwd from the root.
          prefix (if repo (str/trim (or (git ["rev-parse" "--show-prefix"] nil) "")) "")
          filepaths (mapv #(cond->> (status-filepath %)
                             (seq prefix) (str prefix)
                             (seq prefix) lexical-normalize)
                          lines)
          excluded (excluded-patterns file repo filepaths)
          patterns (keep excluded filepaths)
          counts (frequencies patterns)]
      [(vec (keep-indexed #(when-not (excluded (filepaths %1)) %2) lines))
       (when (seq patterns)
         (str/join "\n" (map #(str % "  (" (counts %) ")") (distinct patterns))))])
    [lines nil]))

(defn status
  #_[:=> [:cat Shell :boolean] :string]
  [shell exclude]
  (let [lines (status-lines)]
    (if (empty? lines)
      ""
      (let [[included header] (if exclude (split-excluded lines) [lines nil])
            base ["--multi" "--ansi" "--bind" "ctrl-a:toggle-all"]
            ;; Ctrl-f lists the full status. fzf hands the key back instead of running a command,
            ;; so this works the same on every OS.
            [query key & picked] (str/split-lines
                                  (fzf (cond-> (into base ["--print-query" "--expect" "ctrl-f"])
                                         header (into ["--header" header]))
                                       (str/join "\n" included)))
            picked (if (= "ctrl-f" key)
                     (str/split-lines (fzf (into base ["--query" (or query "")])
                                           (str/join "\n" lines)))
                     picked)]
        (->> picked
             (remove str/blank?)
             (map (comp (quote-for shell) status-filepath))
             (str/join " "))))))

;; MENU

(defn menu
  #_[:=> [:cat Opts] :string]
  [{:keys [shell exclude map-regex]}]
  (let [pick (fzf ["--query=^" "--bind" "one:accept"] "s -- status\nw -- worktrees\nh -- HEAD")]
    (case (first (str/trim pick))
      \s (status shell exclude)
      \w (worktrees shell)
      \h (head map-regex)
      "")))

;; MAIN

(defn parse-opts
  #_[:=> [:cat [:sequential :string]] Opts]
  [args]
  (loop [[arg & more] args
         opts {:shell nil :exclude false :map-regex nil}]
    (case arg
      nil opts
      "--shell" (recur (rest more) (assoc opts :shell (first more)))
      "--exclude" (recur more (assoc opts :exclude true))
      "--map-regex" (recur (rest more) (assoc opts :map-regex (first more)))
      (throw (ex-info (str "Unknown option: " arg) {})))))

(defn error
  #_[:=> [:cat :string] :nil]
  [msg]
  (binding [*out* *err*]
    (println (str "ERROR: " msg)))
  (System/exit 1))

(when (= *file* (System/getProperty "babashka.file"))
  (let [[cmd & args] *command-line-args*
        {:keys [shell exclude map-regex] :as opts} (parse-opts args)]
    (when (and (#{"status" "worktrees" "menu"} cmd) (not (#{"zsh" "pwsh"} shell)))
      (error (str "Usage: git.clj " cmd " --shell <zsh|pwsh>")))
    (print (case cmd
             "status" (status shell exclude)
             "worktrees" (worktrees shell)
             "head" (head map-regex)
             "menu" (menu opts)
             (error "Usage: git.clj <status|worktrees|head|menu>")))
    (flush)))
