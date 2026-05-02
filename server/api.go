package main

import (
	"encoding/json"
	"net/http"
	"os"
)

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

	var data map[string]interface{}

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
