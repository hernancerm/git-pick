(ns git-pick.core-test
  (:require [babashka.fs :as fs]
            [clojure.test :refer [deftest is testing]]
            [git-pick.core :as gp]))

(deftest quote-zsh-test
  (testing "safe filepaths are left as is"
    (is (= "src/a-b_c.clj" (gp/quote-zsh "src/a-b_c.clj")))
    (is (= "ñandú.md" (gp/quote-zsh "ñandú.md"))))
  (testing "others are single quoted"
    (is (= "'a b.txt'" (gp/quote-zsh "a b.txt")))
    (is (= "'a!b'" (gp/quote-zsh "a!b")))
    (is (= "'*.log'" (gp/quote-zsh "*.log")))
    (is (= "'it'\\''s'" (gp/quote-zsh "it's")))))

(deftest quote-pwsh-test
  (is (= "C:\\Users\\a.txt" (gp/quote-pwsh "C:\\Users\\a.txt")))
  (is (= "'a b.txt'" (gp/quote-pwsh "a b.txt")))
  (is (= "'it''s'" (gp/quote-pwsh "it's")))
  (is (= "'$a'" (gp/quote-pwsh "$a"))))

(deftest map-branch-test
  (testing "no regex keeps the branch"
    (is (= "ABC-123-feat" (gp/map-branch "ABC-123-feat" nil)))
    (is (= "ABC-123-feat" (gp/map-branch "ABC-123-feat" ""))))
  (testing "whole match without groups"
    (is (= "ABC-123" (gp/map-branch "feat/ABC-123-x" "[A-Z]+-[0-9]+"))))
  (testing "group 1 when the regex has one"
    (is (= "123" (gp/map-branch "ABC-123-x" "[A-Z]+-([0-9]+)"))))
  (testing "empty when it does not match"
    (is (= "" (gp/map-branch "main" "[A-Z]+-[0-9]+"))))
  (testing "detached HEAD stays empty"
    (is (= "" (gp/map-branch "" "[A-Z]+-[0-9]+")))))

(deftest worktree-paths-test
  (is (= ["/repo" "/repo/wt with\nnewline"]
         (gp/worktree-paths
          (str "worktree /repo\u0000HEAD abc\u0000branch refs/heads/main\u0000\u0000"
               "worktree /repo/wt with\nnewline\u0000HEAD def\u0000detached\u0000\u0000"))))
  (is (empty? (gp/worktree-paths ""))))

(deftest unquote-git-test
  (testing "C-style escapes"
    (is (= "a\"b\\c\td" (gp/unquote-git "\"a\\\"b\\\\c\\td\""))))
  (testing "octal bytes decode as UTF-8"
    (is (= "ñandú.md" (gp/unquote-git "\"\\303\\261and\\303\\272.md\"")))))

(deftest status-filepath-test
  (is (= "a.txt" (gp/status-filepath " M a.txt")))
  (is (= "a b.txt" (gp/status-filepath "?? \"a b.txt\"")))
  (testing "ANSI colors are stripped"
    (is (= "a.txt" (gp/status-filepath "\u001b[32mM\u001b[m  a.txt"))))
  (testing "a rename or copy gives the destination"
    (is (= "new.txt" (gp/status-filepath "R  old.txt -> new.txt")))
    (is (= "mv\"d.txt" (gp/status-filepath "R  plain.txt -> \"mv\\\"d.txt\"")))
    (is (= "b.txt" (gp/status-filepath "C  a -> b.txt")))))

(deftest lexical-normalize-test
  (is (= "a/c" (gp/lexical-normalize "a/b/../c")))
  (is (= "a/b" (gp/lexical-normalize "./a//b/.")))
  (is (= "c" (gp/lexical-normalize "../c"))))

(deftest excluded-patterns-test
  (let [dir (fs/create-temp-dir)
        repo (str (fs/path dir "repo"))
        file (str (fs/path dir ".gitpickignore"))]
    (try
      (gp/git ["-c" "init.templateDir=" "init" "-q" repo] nil)
      (spit file "# Comment.\n*.log\n!keep.log\nbuild/\n")
      (is (= {"a.log" "*.log" "src/b.log" "*.log" "build/x.txt" "build/"}
             (gp/excluded-patterns file repo
                                   ["a.log" "src/b.log" "keep.log" "build/x.txt" "c.txt"])))
      (finally
        (fs/delete-tree dir)))))

(deftest parse-opts-test
  (is (= {:shell "zsh" :exclude true :map-regex "x"}
         (gp/parse-opts ["--shell" "zsh" "--exclude" "--map-regex" "x"])))
  (is (= {:shell nil :exclude false :map-regex nil} (gp/parse-opts [])))
  (is (thrown-with-msg? clojure.lang.ExceptionInfo #"Unknown option: --nope"
                        (gp/parse-opts ["--nope"]))))
