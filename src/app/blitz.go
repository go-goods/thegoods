package main

import "net/http"

func init() {
	http.HandleFunc("/mu-d3b56281-842dfd17-15dd38ac-e0734cdf", handle_blitz)
}

func handle_blitz(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(`42`))
}
