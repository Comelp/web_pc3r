package main

import (
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./dist")))

	// API
	http.HandleFunc("/getState", StateHandler)

	http.ListenAndServe(":8080", nil)
}
