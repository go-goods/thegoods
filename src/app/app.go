package main

import (
	"log"
	"net/http"
	"path/filepath"
	"thegoods.biz/tmplmgr"
)

type Package struct {
	Name     string
	RepoPath string
	Package  string
	VCS      string
}

var (
	mode          = tmplmgr.Production
	assets_dir    = filepath.Join(env("APPROOT", ""), "assets")
	template_dir  = filepath.Join(env("APPROOT", ""), "templates")
	base_template = tmplmgr.Parse(tmpl_root("base.tmpl"))

	context = d{
		"css": []string{
			"bootstrap-slate.min.css",
			"bootstrap-responsive.min.css",
			"main.css",
		},
		"js": []string{
			"jquery.min.js",
			"jquery-ui.min.js",
			"bootstrap.js",
		},
		"packages": []Package{
			{"Template Manager", "git://github.com/goods/tmplmgr.git", "tmplmgr", "git"},
		},
	}
)

func handle_main(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-type", "text/html")
	if err := base_template.Execute(w, context); err != nil {
		log.Println(err)
	}
}

func init() {
	tmplmgr.CompileMode(mode)
	base_template.Blocks(tmpl_root("*.block"))
}

func main() {
	http.HandleFunc("/", handle_main)
	serve_static("/assets", asset_root(""))
	if err := http.ListenAndServe(":"+env("PORT", "9080"), nil); err != nil {
		log.Fatal(err)
	}
}
