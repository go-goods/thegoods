package main

import (
	"github.com/goods/tmplmgr"
	"log"
	"net/http"
	"path/filepath"
)

type Package struct {
	Name string
	RepoPath string
	VCS string
}

var (
	mode          = tmplmgr.Development
	assets_dir    = filepath.Join(env("APPROOT", ""), "assets")
	template_dir  = filepath.Join(env("APPROOT", ""), "templates")
	base_template = tmplmgr.Parse(tmpl_root("base.tmpl"))

	packages = []Package{
		{"Template Manager", "github.com/goods/tmplmgr", "git"},
	}
)

func init() {
	//set our compiler mode
	tmplmgr.CompileMode(mode)

	//add blocks to base template
	base_template.Blocks(tmpl_root("*.block"))
}

func main() {
	http.HandleFunc("/", nil)

	serve_static("/assets", asset_root(""))
	if err := http.ListenAndServe(":"+env("PORT", "9080"), nil); err != nil {
		log.Fatal(err)
	}
}
