package main

import (
	"net/http"
	"os"
	"path/filepath"
)

type d map[string]interface{}

func tmpl_root(path ...string) string {
	elems := append([]string{template_dir}, path...)
	return filepath.Join(elems...)
}

func asset_root(path ...string) string {
	elems := append([]string{assets_dir}, path...)
	return filepath.Join(elems...)
}

func env(key, def string) string {
	if k := os.Getenv(key); k != "" {
		return k
	}
	return def
}

// Serves static files from filesystemDir when any request is made matching
// requestPrefix
func serve_static(requestPrefix, filesystemDir string) {
	fileServer := http.FileServer(http.Dir(filesystemDir))
	handler := http.StripPrefix(requestPrefix, fileServer)
	http.Handle(requestPrefix+"/", handler)
}
