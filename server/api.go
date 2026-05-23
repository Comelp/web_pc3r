package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"
)

// Structures utilisées
type MeteoInfo struct {
	Temperature float64 `json:"temperature"`
	Condition   string  `json:"condition"`
}
type TroopData struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}
type CountryInfo struct {
	LeaderID        *string    `json:"leader_id"`
	AttackedBy      *string    `json:"attacked_by"`
	ConqueredBy     *string    `json:"conquered_by"`
	Meteo           *MeteoInfo `json:"meteo"`
	ProducedGold    int        `json:"produced_gold"`
	Level           int        `json:"level"`
	TroopsAttacking TroopData  `json:"troops_attacking"`
	TroopsDefending TroopData  `json:"troops_defending"`
}
type PlayerInfo struct {
	Couleur  string         `json:"couleur"`
	Password string         `json:"password"`
	Gold     int            `json:"gold"`
	Troops   map[string]int `json:"troops"`
}
type CountryMapInfo struct {
	Color       *string `json:"color"`
	IsAttacked  bool    `json:"is_attacked"`
	IsConquered bool    `json:"is_conquered"`
	AttackedBy  *string `json:"attacked_by"`
	ConqueredBy *string `json:"conquered_by"`
}

var troopCosts = map[string]int{
	"soldiers": 10,
	"tanks":    20,
	"planes":   30,
}

var validInput = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

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

	username, ok := sessions[cookie.Value]
	if !ok {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	file, err := os.Open("server/data/playerInfos.json")
	if err != nil {
		http.Error(w, "Cannot open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var players map[string]PlayerInfo
	json.NewDecoder(file).Decode(&players)

	player, ok := players[username]
	if !ok {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"username": username,
		"troops":   player.Troops,
		"gold":     player.Gold,
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

	if !validInput.MatchString(username) {
		http.Error(w, "Username invalide (a-z A-Z 0-9 _ uniquement)", http.StatusBadRequest)
		return
	}

	if !validInput.MatchString(password) {
		http.Error(w, "Password invalide (a-z A-Z 0-9 _ uniquement)", http.StatusBadRequest)
		return
	}

	if username == "" || password == "" {
		http.Error(w, "Champs manquants", http.StatusBadRequest)
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
			Color:       color,
			IsAttacked:  country.AttackedBy != nil,
			IsConquered: country.ConqueredBy != nil,
			AttackedBy:  country.AttackedBy,
			ConqueredBy: country.ConqueredBy,
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

const (
	Attack       = "Attaque 🪖"
	Distribution = "Paix 🤝"
)

// Donne la phase actuelle du jeu en fonction du temps réel.
func GetCurrentPhase() string {
	minutes := time.Now().Unix() / 60
	if (minutes/10)%2 == 1 {
		return Attack
	}
	return Distribution
}

func PhaseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"phase": GetCurrentPhase()})
}

// Renvoie le dernier popup de fin de phase (généré par onPhaseEnd)
func PhasePopupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, err := os.ReadFile("server/data/phasePopup.json")
	if err != nil {
		// si il n'existe pas, renvoyer une structure vide
		json.NewEncoder(w).Encode(map[string]any{
			"phase":            "",
			"lost_to_war":      []string{},
			"lost_to_weather":  []string{},
			"conquered":        []string{},
			"gained_by_attack": []string{},
			"improved":         []string{},
		})
		return
	}

	w.Write(data)
}

// Acquitte et réinitialise le popup de fin de phase (client appelle après affichage)
func AckPhasePopupHandler(w http.ResponseWriter, r *http.Request) {
	empty := map[string]any{
		"phase":            "",
		"lost_to_war":      []string{},
		"lost_to_weather":  []string{},
		"conquered":        []string{},
		"gained_by_attack": []string{},
		"improved":         []string{},
	}
	data, _ := json.MarshalIndent(empty, "", "  ")
	os.WriteFile("server/data/phasePopup.json", data, 0644)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Amélioration d'un pays. Un pays amélioré produit plus de ressources,
// mais il faut être le leader du pays pour pouvoir l'améliorer,
// et on ne peut améliorer que pendant la phase de distribution.
func UpgradeHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	println("----------------------\nEntering upgrade Handler")

	if GetCurrentPhase() != Distribution {
		http.Error(w, "Upgrade only allowed during Distribution phase", http.StatusBadRequest)
		return
	}

	println("-- phase de jeu correct")

	country := r.URL.Query().Get("country")

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Not connected", http.StatusUnauthorized)
		return
	}
	playerID, ok := sessions[cookie.Value]
	if !ok {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	println("-- utilisateur bien connecté")

	// --- Charger les pays ---
	countryFile, err := os.Open("server/data/countryInfos.json")
	if err != nil {
		http.Error(w, "Cannot open file", http.StatusInternalServerError)
		return
	}
	defer countryFile.Close()

	var countries map[string]CountryInfo
	json.NewDecoder(countryFile).Decode(&countries)

	countryInfo, ok := countries[country]
	if !ok {
		http.Error(w, "Country not found", http.StatusNotFound)
		return
	}

	println("-- pays trouvé")

	if countryInfo.LeaderID == nil || *countryInfo.LeaderID != playerID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	println("-- user est le leader")

	if countryInfo.Level >= 3 {
		http.Error(w, "Level maximum atteint", http.StatusBadRequest)
		return
	}

	println("-- possible de level up")

	// --- Charger les joueurs ---
	playerFile, err := os.Open("server/data/playerInfos.json")
	if err != nil {
		http.Error(w, "Cannot open file", http.StatusInternalServerError)
		return
	}
	defer playerFile.Close()

	var players map[string]PlayerInfo
	json.NewDecoder(playerFile).Decode(&players)

	playerInfo, ok := players[playerID]
	if !ok {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	// --- Vérifier l'or ---
	upgradeCost := countryInfo.ProducedGold * 1000 * 2 * (countryInfo.Level + 1)
	if playerInfo.Gold < upgradeCost {
		http.Error(w, "Insufficient gold", http.StatusPaymentRequired)
		return
	}
	println("-- vérification de l'or passée !")
	println("-- applique l'upgrade!")

	// --- Appliquer l'upgrade ---
	playerInfo.Gold -= upgradeCost
	countryInfo.Level += 1

	players[playerID] = playerInfo
	countries[country] = countryInfo

	// --- Sauvegarder les deux fichiers ---
	countryData, _ := json.MarshalIndent(countries, "", "  ")
	os.WriteFile("server/data/countryInfos.json", countryData, 0644)

	playerData, _ := json.MarshalIndent(players, "", "  ")
	os.WriteFile("server/data/playerInfos.json", playerData, 0644)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":         "upgraded",
		"new_level":      countryInfo.Level,
		"gold_remaining": playerInfo.Gold,
	})

}

func AttackHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	country := r.URL.Query().Get("country")

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Not connected", http.StatusUnauthorized)
		return
	}

	playerID, ok := sessions[cookie.Value]
	if !ok {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	file, err := os.Open("server/data/countryInfos.json")
	if err != nil {
		http.Error(w, "Cannot open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var countries map[string]CountryInfo
	json.NewDecoder(file).Decode(&countries)

	// Vérifier que le joueur n'attaque pas déjà un autre pays
	for name, info := range countries {
		if info.AttackedBy != nil && *info.AttackedBy == playerID {
			http.Error(w, fmt.Sprintf("You're already attacking a country : %s", name), http.StatusBadRequest)
			return
		}
	}

	countryInfo, ok := countries[country]
	if !ok {
		http.Error(w, "Country not found", http.StatusNotFound)
		return
	}

	if countryInfo.LeaderID == nil {
		http.Error(w, "Country is not occupied", http.StatusBadRequest)
		return
	}

	if countryInfo.AttackedBy != nil {
		http.Error(w, "Country already under attack", http.StatusBadRequest)
		return
	}

	if countryInfo.ConqueredBy != nil {
		http.Error(w, "Country already being conquered", http.StatusBadRequest)
		return
	}

	if *countryInfo.LeaderID == playerID {
		http.Error(w, "Cannot attack your own country", http.StatusForbidden)
		return
	}

	countryInfo.AttackedBy = &playerID
	countries[country] = countryInfo

	data, _ := json.MarshalIndent(countries, "", "  ")
	os.WriteFile("server/data/countryInfos.json", data, 0644)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":     "attacking",
		"country":    country,
		"attackedBy": playerID,
	})
}

func ConquerHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	country := r.URL.Query().Get("country")

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Not connected", http.StatusUnauthorized)
		return
	}

	playerID, ok := sessions[cookie.Value]
	if !ok {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	file, err := os.Open("server/data/countryInfos.json")
	if err != nil {
		http.Error(w, "Cannot open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var countries map[string]CountryInfo
	json.NewDecoder(file).Decode(&countries)

	// Vérifier que le joueur n'attaque pas déjà un autre pays
	for name, info := range countries {
		if info.ConqueredBy != nil && *info.ConqueredBy == playerID {
			http.Error(w, fmt.Sprintf("You're already conquering a country : %s", name), http.StatusBadRequest)
			return
		}
	}

	countryInfo, ok := countries[country]
	if !ok {
		http.Error(w, "Country not found", http.StatusNotFound)
		return
	}

	if countryInfo.LeaderID != nil {
		http.Error(w, "Country is already occupied", http.StatusBadRequest)
		return
	}

	if countryInfo.AttackedBy != nil {
		http.Error(w, "Country already under attack", http.StatusBadRequest)
		return
	}

	if countryInfo.ConqueredBy != nil {
		http.Error(w, "Country already being conquered", http.StatusBadRequest)
		return
	}

	countryInfo.ConqueredBy = &playerID

	countries[country] = countryInfo

	data, _ := json.MarshalIndent(countries, "", "  ")
	os.WriteFile("server/data/countryInfos.json", data, 0644)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":      "conquering",
		"country":     country,
		"conqueredBy": playerID,
	})
}

func RuleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./dist/rules.html")
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func BuyTroopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	troop := r.URL.Query().Get("troop")
	cost, ok := troopCosts[troop]
	if !ok {
		http.Error(w, "Troupe invalide", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Not connected", http.StatusUnauthorized)
		return
	}
	playerID, ok := sessions[cookie.Value]
	if !ok {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	playerFile, err := os.Open("server/data/playerInfos.json")
	if err != nil {
		http.Error(w, "Cannot open file", http.StatusInternalServerError)
		return
	}
	defer playerFile.Close()

	var players map[string]PlayerInfo
	json.NewDecoder(playerFile).Decode(&players)

	player, ok := players[playerID]
	if !ok {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	if player.Gold < cost {
		http.Error(w, "Or insuffisant", http.StatusPaymentRequired)
		return
	}

	player.Gold -= cost
	if player.Troops == nil {
		player.Troops = map[string]int{}
	}
	player.Troops[troop] += 1
	players[playerID] = player

	data, _ := json.MarshalIndent(players, "", "  ")
	os.WriteFile("server/data/playerInfos.json", data, 0644)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":         "ok",
		"troop":          troop,
		"gold_remaining": player.Gold,
		"troops":         player.Troops,
	})
}

func DeployTroopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if GetCurrentPhase() != Attack {
		http.Error(w, "Deploy only allowed during Attack phase", http.StatusBadRequest)
		return
	}

	country := r.URL.Query().Get("country")
	troop := r.URL.Query().Get("troop")
	mode := r.URL.Query().Get("mode") // "attacking" ou "defending"
	count := r.URL.Query().Get("count")

	if _, ok := troopCosts[troop]; !ok {
		http.Error(w, "Troupe invalide", http.StatusBadRequest)
		return
	}

	if mode != "attacking" && mode != "defending" {
		http.Error(w, "Mode invalide (attacking ou defending)", http.StatusBadRequest)
		return
	}

	var troopCount int
	if _, err := fmt.Sscanf(count, "%d", &troopCount); err != nil || troopCount <= 0 {
		http.Error(w, "Count invalide", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Not connected", http.StatusUnauthorized)
		return
	}
	playerID, ok := sessions[cookie.Value]
	if !ok {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	countryFile, err := os.Open("server/data/countryInfos.json")
	if err != nil {
		http.Error(w, "Cannot open file", http.StatusInternalServerError)
		return
	}
	defer countryFile.Close()

	var countries map[string]CountryInfo
	json.NewDecoder(countryFile).Decode(&countries)

	countryInfo, ok := countries[country]
	if !ok {
		http.Error(w, "Country not found", http.StatusNotFound)
		return
	}

	playerFile, err := os.Open("server/data/playerInfos.json")
	if err != nil {
		http.Error(w, "Cannot open file", http.StatusInternalServerError)
		return
	}
	defer playerFile.Close()

	var players map[string]PlayerInfo
	json.NewDecoder(playerFile).Decode(&players)

	player, ok := players[playerID]
	if !ok {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	switch mode {
	case "attacking":
		if countryInfo.AttackedBy == nil || *countryInfo.AttackedBy != playerID {
			http.Error(w, "Vous n'attaquez pas ce pays", http.StatusForbidden)
			return
		}
		if countryInfo.TroopsAttacking.Count > 0 {
			http.Error(w, "Vous avez déjà déployé des troupes en attaque", http.StatusBadRequest)
			return
		}

	case "defending":
		if countryInfo.LeaderID == nil || *countryInfo.LeaderID != playerID {
			http.Error(w, "Vous n'êtes pas le leader de ce pays", http.StatusForbidden)
			return
		}
		if countryInfo.TroopsDefending.Count > 0 {
			http.Error(w, "Vous avez déjà déployé des troupes en défense", http.StatusBadRequest)
			return
		}
	}

	if player.Troops == nil || player.Troops[troop] < troopCount {
		http.Error(w, "Troupes insuffisantes", http.StatusPaymentRequired)
		return
	}

	player.Troops[troop] -= troopCount
	players[playerID] = player

	troopData := TroopData{Type: troop, Count: troopCount}
	if mode == "attacking" {
		countryInfo.TroopsAttacking = troopData
	} else {
		countryInfo.TroopsDefending = troopData
	}
	countries[country] = countryInfo

	countryData, _ := json.MarshalIndent(countries, "", "  ")
	os.WriteFile("server/data/countryInfos.json", countryData, 0644)

	playerData, _ := json.MarshalIndent(players, "", "  ")
	os.WriteFile("server/data/playerInfos.json", playerData, 0644)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "deployed",
		"country": country,
		"mode":    mode,
		"troop":   troop,
		"count":   troopCount,
	})
}
