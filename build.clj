(ns build
  (:require [clojure.tools.build.api :as b]))

(def basis (delay (b/create-basis {:project "deps.edn"})))

(defn uber [_]
  (b/delete {:path "target"})
  (b/compile-clj {:basis @basis
                  :ns-compile '[git-pick.core]
                  :class-dir "target/classes"})
  (b/uber {:basis @basis
           :class-dir "target/classes"
           :uber-file "target/git-pick.jar"
           :main 'git-pick.core}))
