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
	http.HandleFunc("/getPhasePopup", PhasePopupHandler)
	http.HandleFunc("/ackPhasePopup", AckPhasePopupHandler)
	http.HandleFunc("/getState", StateHandler)
	http.HandleFunc("/getMapInfos", MapInfosHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/login", LoginHandler) // Ici il y a un Get ET un Post
	http.HandleFunc("/logout", LogoutHandler)
	http.HandleFunc("/upgradeCountry", UpgradeHandler)
	http.HandleFunc("/conquerCountry", ConquerHandler)
	http.HandleFunc("/attackCountry", AttackHandler)
	http.HandleFunc("/getRules", RuleHandler)
	http.HandleFunc("/buyTroop", BuyTroopHandler)
	http.HandleFunc("/deployTroop", DeployTroopHandler)

	http.ListenAndServe(":8080", nil)
}
