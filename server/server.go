package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
)

type GameState struct {
	mut       sync.RWMutex
	countries map[string]CountryInfo
	players   map[string]PlayerInfo
}

var gameState = GameState{
	countries: make(map[string]CountryInfo),
	players:   make(map[string]PlayerInfo),
}

func main() {
	// charger l'état initial depuis le disque
	if err := loadStateFromDisk(); err != nil {
		log.Printf("Warning: cannot load initial state: %v", err)
	}

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

	log.Println("Server starting on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// loadStateFromDisk lit les fichiers JSON initiaux et peuple gameState.
func loadStateFromDisk() error {
	// countries
	if data, err := os.ReadFile("server/data/countryInfos.json"); err == nil {
		var countries map[string]CountryInfo
		if err := json.Unmarshal(data, &countries); err == nil {
			gameState.SetCountries(countries)
		} else {
			return err
		}
	}

	// players
	if data, err := os.ReadFile("server/data/playerInfos.json"); err == nil {
		var players map[string]PlayerInfo
		if err := json.Unmarshal(data, &players); err == nil {
			gameState.SetPlayers(players)
		} else {
			return err
		}
	}

	return nil
}

// - - - - - - - - - - - - - - - - - -
// Getters et setters avec mutex
// - - - - - - - - - - - - - - - - - -

// On utilise le read lock et une copie pour relacher le mutex rapidement
func (gs *GameState) GetCountries() map[string]CountryInfo {
	gs.mut.RLock()
	defer gs.mut.RUnlock() // defer libère le mutex à la fin (même en cas d'erreur)

	copy := make(map[string]CountryInfo)
	for k, v := range gs.countries {
		copy[k] = v
	}
	return copy
}

func (gs *GameState) GetPlayers() map[string]PlayerInfo {
	gs.mut.RLock()
	defer gs.mut.RUnlock()

	copy := make(map[string]PlayerInfo)
	for k, v := range gs.players {
		copy[k] = v
	}
	return copy
}

func (gs *GameState) SetCountries(countries map[string]CountryInfo) {
	gs.mut.Lock()
	defer gs.mut.Unlock()
	gs.countries = countries
}

func (gs *GameState) SetPlayers(players map[string]PlayerInfo) {
	gs.mut.Lock()
	defer gs.mut.Unlock()
	gs.players = players
}
