package main

import (
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./dist")))

	// API
	go func() {
		StartWeatherPuller()
	}()
	http.HandleFunc("/getState", StateHandler)
	http.HandleFunc("/getMapInfos", MapInfosHandler)

	http.ListenAndServe(":8080", nil)
}
