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
	go func() {
		StartPhaseWatcher()
	}()
	http.HandleFunc("/me", MeHandler)
	http.HandleFunc("/getPhase", PhaseHandler)
	http.HandleFunc("/getState", StateHandler)
	http.HandleFunc("/getMapInfos", MapInfosHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/logout", LogoutHandler)
	http.HandleFunc("/upgradeCountry", UpgradeHandler)
	http.HandleFunc("/conquerCountry", ConquerHandler)
	http.HandleFunc("/attackCountry", AttackHandler)
	http.HandleFunc("/getRules", RuleHandler)

	http.ListenAndServe(":8080", nil)
}
