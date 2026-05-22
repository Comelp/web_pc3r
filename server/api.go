package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Structures utilisées
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
	Couleur  string `json:"couleur"`
	Password string `json:"password"`
}
type CountryMapInfo struct {
	Color      *string `json:"color"`
	IsAttacked bool    `json:"is_attacked"`
}

// Sessions en mémoire qui sont sauvegardés
var sessions = map[string]string{}

func newSessionID(username string) string {
	// Produit l'id de session. Exemple : "Paul-124345786"
	return fmt.Sprintf("%s-%d", username, time.Now().UnixNano())
}

// API Handlers
func MeHandler(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Not connected", http.StatusUnauthorized)
		return
	}

	println("COOKIE:", cookie.Value)
	println("SESSIONS:")
	for k, v := range sessions {
		println(k, "=>", v)
	}

	username, ok := sessions[cookie.Value]
	if !ok {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"username": username,
	})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	couleur := r.FormValue("couleur")

	file, err := os.Open("server/data/playerInfos.json")
	if err != nil {
		http.Error(w, "Cannot open file : playerInfos.json", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var players map[string]PlayerInfo
	json.NewDecoder(file).Decode(&players)

	if _, exists := players[username]; exists {
		http.Error(w, "Utilisateur déjà existant", http.StatusConflict)
		return
	}

	players[username] = PlayerInfo{
		Couleur:  couleur,
		Password: password,
	}

	file.Close()

	file, err = os.Create("server/data/playerInfos.json")
	if err != nil {
		http.Error(w, "Cannot write file : playerInfos.json", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(players)

	sessionID := newSessionID(username)
	sessions[sessionID] = username

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// AFFICHAGE PAGE LOGIN : Get Login
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./dist/login.html")
		return
	}

	// GESTION LOGIN : Post Login
	username := r.FormValue("username")
	password := r.FormValue("password")

	file, err := os.Open("server/data/playerInfos.json")
	if err != nil {
		http.Error(w, "Cannot open file : playerInfos.json", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var players map[string]PlayerInfo
	json.NewDecoder(file).Decode(&players)

	player, ok := players[username]
	if !ok {
		http.Error(w, "Utilisateur non trouvé", http.StatusUnauthorized)
		return
	}
	if player.Password != password {
		http.Error(w, "Mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	sessionID := newSessionID(username)
	sessions[sessionID] = username

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id", // nom du cookie
		Value:    sessionID,    // l'id de session
		Path:     "/",          // il sera utilisé sur tous les appels d'api
		HttpOnly: true,         // cookie pas accessible depuis le JavaScript
		MaxAge:   3600,         // quand il expire
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session_id"); err == nil {
		// on supprime la session localement
		delete(sessions, cookie.Value)
	}
	// On supprime le cookie côté serveur
	http.SetCookie(w, &http.Cookie{Name: "session_id", Value: "", Path: "/", MaxAge: -1})

	http.Redirect(w, r, "/", http.StatusSeeOther)
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

	println("JSON chargé !")

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
