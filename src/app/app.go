package main

import (
	"github.com/zeebo/bencode"
	"hello"
	"log"
	"net/http"
	"os"
)

func benc(w http.ResponseWriter, req *http.Request) {
	enc := bencode.NewEncoder(w)
	if err := enc.Encode(map[string]interface{}{
		"foo": "bar",
		"baz": 2,
		"buf": []string{"baz", "bloof"},
	}); err != nil {
		log.Println(err)
	}
	log.Printf("Handled request from %s :)", req.RemoteAddr)
}

func main() {
	http.HandleFunc("/", hello.Hello)
	http.HandleFunc("/benc", benc)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
