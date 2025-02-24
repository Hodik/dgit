package main

import (
	"fmt"
	"os"
	"strings"
)

type treeEntry struct {
	hash string
	path string
	t    string
}

type Commit struct {
  Hash    string
  Tree string
  Message string
  Parent string
}

func commit(message string) string {
	head := getHEAD()
	tree := writeTree(".")
	data := fmt.Sprintf("tree %s", tree)
  if head != "" {
    data += fmt.Sprintf("\nparent %s", head)
  }
  data += fmt.Sprintf("\n\n%s", message)

	commithash := hashObject([]byte(data), "commit")
	setHEAD(commithash)
	return commithash
}

func getCommit(hash string) *Commit {
  data := catObject(hash, "commit")
  parts := strings.Split(data, "\n")
  var parent string
  if len(parts) > 3 {
    parent = parts[1][7:]
  } else {
    parent = ""
  }
  return &Commit{Hash: hash, Message: parts[len(parts) - 1], Parent: parent, Tree: parts[0][5:]}
}

func log() {
  hash := getHEAD()

  for hash != "" {
    commit := getCommit(hash)

    fmt.Printf("commit %s\ntree %s\n\n%s\n\n", commit.Hash, commit.Tree, commit.Message)
    hash = commit.Parent
  }
}

func checkout(hash string) {
  readTree(getCommit(hash).Tree)
  setHEAD(hash)
}

func setHEAD(hash string) {
	os.WriteFile(".dgit/HEAD", []byte(hash), 0644)
}

func getHEAD() string {
	data, err := os.ReadFile(".dgit/HEAD")
	if err != nil {
    return ""
	}
	return string(data)
}

func getTreeEntries(hash string) []treeEntry {

	treestr := catObject(hash, "tree")
	var entries []treeEntry
	for _, line := range strings.Split(treestr, "\n") {
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		entries = append(entries, treeEntry{hash: parts[1], path: parts[2], t: parts[0]})
	}

	return entries
}

func getTree(hash, basepath string) map[string]string {

	entries := getTreeEntries(hash)
	tree := make(map[string]string)
	for _, entry := range entries {
		var path = basepath + "/" + entry.path
		if entry.t == "blob" {
			tree[path] = entry.hash
		} else {
			inner := getTree(entry.hash, path)
			for k, v := range inner {
				tree[k] = v
			}
		}
	}

	return tree

}

func readTree(hash string) {
	tree := getTree(hash, ".")
	emptyDirectory(".")
	for k, v := range tree {
		dirs := strings.Join(strings.Split(k, "/")[:len(strings.Split(k, "/"))-1], "/")

		if err := os.MkdirAll(dirs, 0755); err != nil {
			panic(err)
		}

		err := os.WriteFile(k, []byte(catObject(v, "blob")), 0644)

		if err != nil {
			panic(err)
		}
	}
}

func emptyDirectory(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if isIgnored(entry.Name()) {
			continue
		}

		if entry.IsDir() {
			emptyDirectory(path + "/" + entry.Name())
		} else {
			err := os.Remove(path + "/" + entry.Name())
			if err != nil {
				panic(err)
			}
		}
	}
}

func writeTree(path string) string {
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	var data string
	for _, entry := range entries {
		if isIgnored(entry.Name()) {
			continue
		}

		if !entry.IsDir() {
			hash := hashFile(path + "/" + entry.Name())

			data += fmt.Sprintf("blob %s %s\n", hash, entry.Name())
		} else {
			data += fmt.Sprintf("tree %s %s\n", writeTree(path+"/"+entry.Name()), entry.Name())
		}
	}

	return hashObject([]byte(data), "tree")
}

func hashFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return hashObject(data, "blob")
}

func isIgnored(name string) bool {
	ignored := []string{".dgit", ".git", "dgit"}
	for _, i := range ignored {
		if i == name {
			return true
		}
	}
	return false
}
