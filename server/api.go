package main

import (
	"encoding/json"
	"net/http"
	"os"
)

type MeteoInfo struct {
	Temperature float64 `json:"temperature"`
	Condition   string  `json:"condition"`
}
type CountryInfo struct {
	LeaderID           *string            `json:"leader_id"`
	AttackedBy         *string            `json:"attacked_by"`
	Meteo              *MeteoInfo         `json:"meteo"`
	ProducedRessources map[string]float64 `json:"produced_ressources"`
	Troops             map[string]float64 `json:"troops"`
}
type PlayerInfo struct {
	Couleur string `json:"couleur"`
}
type CountryMapInfo struct {
	Color      *string `json:"color"`
	IsAttacked bool    `json:"is_attacked"`
}

func MapInfosHandler(w http.ResponseWriter, r *http.Request) {
	countryFile, err1 := os.Open("server/data/countryInfos.json")
	if err1 != nil {
		println("ERREUR OPEN:", err1.Error())
		http.Error(w, "Cannot open file", http.StatusInternalServerError)
		return
	}
	defer countryFile.Close()
	playerFile, err2 := os.Open("server/data/playerInfos.json")
	if err2 != nil {
		println("ERREUR OPEN:", err2.Error())
		http.Error(w, "Cannot open file", http.StatusInternalServerError)
		return
	}
	defer playerFile.Close()

	var countries map[string]CountryInfo
	var players map[string]PlayerInfo
	json.NewDecoder(countryFile).Decode(&countries)
	json.NewDecoder(playerFile).Decode(&players)

	result := make(map[string]CountryMapInfo)

	for name, country := range countries {
		var color *string
		if country.LeaderID != nil {
			if player, ok := players[*country.LeaderID]; ok {
				color = &player.Couleur
			}
		}
		result[name] = CountryMapInfo{
			Color:      color,
			IsAttacked: country.AttackedBy != nil,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func StateHandler(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	println("country reçu =", country)

	file, err := os.Open("server/data/countryInfos.json")
	if err != nil {
		println("ERREUR OPEN:", err.Error())
		http.Error(w, "Cannot open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var data map[string]CountryInfo

	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		println("ERREUR JSON:", err.Error())
		http.Error(w, "Invalid JSON", http.StatusInternalServerError)
		return
	}

	println("JSON chargé OK")

	info, ok := data[country]
	if !ok {
		println("Pays non trouvé")
		http.Error(w, "Country not found", http.StatusNotFound)
		return
	}

	println("Pays trouvé, envoi réponse")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
