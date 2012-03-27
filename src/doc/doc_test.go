package doc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGrab(t *testing.T) {
	dir, updated, err := Grab("thegoods.biz/tmplmgr", "git://github.com/goods/tmplmgr.git")
	defer os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}
	if updated != true {
		t.Fatal("Not updated on the first grab")
	}

	_, updated, err = Grab("thegoods.biz/tmplmgr", "git://github.com/goods/tmplmgr.git")
	if err != nil {
		t.Fatal(err)
	}
	if updated != false {
		t.Fatal("Updated on the second grab")
	}
}

func TestBuild(t *testing.T) {
	imp := "thegoods.biz/tmplmgr"
	dir, _, err := Grab(imp, "git://github.com/goods/tmplmgr.git")
	defer os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}

	p, err := buildDoc(imp, files)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", p)
}
