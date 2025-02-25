package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type treeEntry struct {
	hash string
	path string
	t    string
}

type Commit struct {
	Hash    string
	Tree    string
	Message string
	Parent  string
}

func tag(name, hash string) {
	setRef("refs/tags/"+name, hash)
}

func commit(message string) string {
	head := getRef("HEAD")
	tree := writeTree(".")
	data := fmt.Sprintf("tree %s", tree)
	if head != "" {
		data += fmt.Sprintf("\nparent %s", head)
	}
	data += fmt.Sprintf("\n\n%s", message)

	commithash := hashObject([]byte(data), "commit")
	setRef("HEAD", commithash)
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
	return &Commit{Hash: hash, Message: parts[len(parts)-1], Parent: parent, Tree: parts[0][5:]}
}

func log(hash string) {
	for hash != "" {
		commit := getCommit(hash)

		fmt.Printf("commit %s\ntree %s\n\n%s\n\n", commit.Hash, commit.Tree, commit.Message)
		hash = commit.Parent
	}
}

func k() {
	refs := getRefs()

	hashes := []string{}
	dot := "digraph commits {\n"
	for _, ref := range refs {
		hashes = append(hashes, ref[1])
		dot += fmt.Sprintf("\"%s\" [shape=note]\n", ref[0])
		dot += fmt.Sprintf("\"%s\" -> \"%s\"\n", ref[0], ref[1])
	}

	for hash := range commitsAndParents(hashes) {
		commit := getCommit(hash)
		dot += fmt.Sprintf("\"%s\" [shape=box style=filled label=\"%s\"]\n", hash, hash[:10])
		if commit.Parent != "" {
			dot += fmt.Sprintf("\"%s\" -> \"%s\"\n", hash, commit.Parent)
		}
	}
	dot += "}"

	outputFile, err := os.Create("output.png")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	cmd := exec.Command("dot", "-Tpng", "/dev/stdin")
	cmd.Stdin = strings.NewReader(dot)
	cmd.Stdout = outputFile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("Error running dot command:", err)
	}

	var openCmd *exec.Cmd
	switch os := runtime.GOOS; os {
	case "darwin":
		openCmd = exec.Command("open", "output.png")
	case "linux":
		openCmd = exec.Command("xdg-open", "output.png")
	case "windows":
		openCmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", "output.png")
	default:
		fmt.Println("Unsupported platform")
		return
	}

	if err := openCmd.Start(); err != nil {
		fmt.Println("Error opening image:", err)
	}
}

func commitsAndParents(hashes []string) chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		visited := make(map[string]bool)
		for len(hashes) > 0 {
			hash := hashes[len(hashes)-1]
			hashes = hashes[:len(hashes)-1]
			if visited[hash] == true {
				continue
			}
			visited[hash] = true
			ch <- hash
			commit := getCommit(hash)
			if commit.Parent != "" {
				hashes = append(hashes, commit.Parent)
			}
		}
	}()
	return ch
}

func getRefs() [][]string {
	refs := []string{"HEAD"}
	if err := filepath.Walk(".dgit/refs", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			refs = append(refs, strings.ReplaceAll(path, ".dgit/", ""))
		}
		return nil
	}); err != nil {
		panic(err)
	}

	result := make([][]string, len(refs))
	for i, ref := range refs {
		result[i] = []string{ref, getRef(ref)}
	}

	return result
}

func checkout(hash string) {
	readTree(getCommit(hash).Tree)
	setRef("HEAD", hash)
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
