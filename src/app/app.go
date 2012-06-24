package main

import (
	"doc"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"thegoods.biz/httpbuf"
	"thegoods.biz/tmplmgr"
)

type Package struct {
	Name       string
	RepoPath   string
	ImportPath string
}

func (p *Package) URL() string {
	return p.ImportPath[13:]
}

var (
	mode          = tmplmgr.Production
	assets_dir    = filepath.Join(env("APPROOT", ""), "assets")
	template_dir  = filepath.Join(env("APPROOT", ""), "templates")
	base_template = tmplmgr.Parse(tmpl_root("base.tmpl"))

	packages = []*Package{
		{"Template Manager", "git://github.com/goods/tmplmgr.git", "thegoods.biz/tmplmgr"},
		{"Forms", "git://github.com/goods/forms.git", "thegoods.biz/forms"},
		{"HTTP Buffer", "git://github.com/goods/httpbuf.git", "thegoods.biz/httpbuf"},
		{"Broadcaster", "git://github.com/goods/broadcast.git", "thegoods.biz/broadcast"},
	}

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
		"packages": packages,
	}
)

func search(imp string) (p *Package) {
	for _, pack := range packages {
		if pack.ImportPath == imp {
			p = pack
			return
		}
	}
	return
}

func cached_main(w http.ResponseWriter, req *http.Request) {
	if buf, ex := app_cache.get(req.URL.Path); ex {
		buf.Apply(w)
		return
	}

	buf := new(httpbuf.Buffer)
	handle_main(buf, req)
	app_cache.set(req.URL.Path, buf)
	buf.Apply(w)
}

func handle_main(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		//send the default
		w.Header().Set("Content-type", "text/html")
		if err := base_template.Execute(w, context, tmpl_root("app", "index.tmpl")); err != nil {
			log.Println(err)
		}
		return
	}

	//check the path
	pack := search(fmt.Sprintf("thegoods.biz%s", req.URL.Path))
	if pack == nil {
		//redirect to /
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusSeeOther)
		return
	}

	//grab the docs for the package
	p, err := doc.LoadDocs(pack.ImportPath, pack.RepoPath)
	if err != nil {
		log.Println(err)
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusSeeOther)
		return
	}

	//we have a package
	context["pdoc"] = p
	w.Header().Set("Content-type", "text/html")
	if err := base_template.Execute(w, context, tmpl_root("doc", "*.block")); err != nil {
		log.Println(err)
	}
}

func init() {
	tmplmgr.CompileMode(mode)
	base_template.Blocks(tmpl_root("*.block"))
	for name, fnc := range doc.Funcs {
		base_template.Call(name, fnc)
	}
}

func main() {
	http.HandleFunc("/", cached_main)
	serve_static("/assets", asset_root(""))
	if err := http.ListenAndServe(":"+env("PORT", "9080"), nil); err != nil {
		log.Fatal(err)
	}
}
