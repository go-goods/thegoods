package doc

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var mu sync.RWMutex

func Grab(imp, clone string) (dir string, updated bool, err error) {
	mu.Lock()
	defer mu.Unlock()

	dir = filepath.Join(os.TempDir(), imp)

	if dir_exists(filepath.Join(dir, ".git")) {
		updated, err = do_update(dir)
	} else {
		updated, err = true, do_clone(clone, dir)
	}
	return
}

func dir_exists(dir string) (ex bool) {
	if _, err := os.Stat(dir); err == nil {
		ex = true
	}
	return
}

func do_clone(clone, dir string) (err error) {
	//run a git clone into the specified directory
	cmd := exec.Command("git", "clone", clone, dir)
	err = cmd.Run()
	return
}

func do_update(dir string) (updated bool, err error) {
	//run a git update in the specified directory and check to see if it was
	//already up to date or not.
	var out bytes.Buffer

	cmd := exec.Command("git", "pull", "origin", "master")
	cmd.Dir = dir
	cmd.Stdout = &out
	err = cmd.Run()

	updated = out.String() == "Already up-to-date."
	return
}

func LoadDocs(imp, clone string) (p *Package, err error) {
	dir, _, err := Grab(imp, clone)
	if err != nil {
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return
	}

	p, err = buildDoc(imp, files)
	return
}
