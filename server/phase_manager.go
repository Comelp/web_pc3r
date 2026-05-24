package main

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

type GoldEarnedEvent struct {
	Player string `json:"player"`
	Amount int    `json:"amount"`
}

type ConqueredEvent struct {
	Country string `json:"country"`
	By      string `json:"by"`
}

type GainedAttackEvent struct {
	Country       string `json:"country"`
	Attacker      string `json:"attacker"`
	PreviousOwner string `json:"previous_owner"`
}

type LostAttackEvent struct {
	Country  string `json:"country"`
	Attacker string `json:"attacker"`
	Owner    string `json:"owner"`
}

type LostWeatherEvent struct {
	Country       string `json:"country"`
	PreviousOwner string `json:"previous_owner"`
	Reason        string `json:"reason"`
}

type ImprovedEvent struct {
	Country  string `json:"country"`
	Owner    string `json:"owner"`
	OldLevel int    `json:"old_level"`
	NewLevel int    `json:"new_level"`
}

var lastPhase string

// Multiplicateur pierre-feuille-ciseaux :
// Soldats > Tanks > Avions > Soldats
// Retourne le multiplicateur pour troopA contre troopB
func combatMultiplier(troopA, troopB string) float64 {
	advantages := map[string]string{
		"soldiers": "tanks",
		"tanks":    "planes",
		"planes":   "soldiers",
	}
	if advantages[troopA] == troopB {
		return 2.0 // avantage
	}
	if advantages[troopB] == troopA {
		return 0.5 // désavantage
	}
	return 1.0 // neutre
}

// Calcule le score de combat d'une troupe
func combatScore(troop TroopData, againstTroop string) float64 {
	cost := float64(troopCosts[troop.Type])
	mult := combatMultiplier(troop.Type, againstTroop)
	return float64(troop.Count) * cost * mult
}

// Résout un combat entre attaquant et défenseur.
// Retourne true si l'attaquant gagne, ainsi que les pertes de chaque camp (en nombre de troupes).
func resolveCombat(attacking, defending TroopData, players map[string]PlayerInfo, attackerID, defenderID string) (attackerWins bool, attackerLosses, defenderLosses int) {

	scoreAttacker := combatScore(attacking, defending.Type)
	scoreDefender := combatScore(defending, attacking.Type)

	// si aucune troupe déployée d'un côté, l'autre gagne automatiquement
	if attacking.Count == 0 {
		return false, 0, 0
	}
	if defending.Count == 0 {
		return true, 0, 0
	}

	total := scoreAttacker + scoreDefender
	chanceAttacker := scoreAttacker / total

	// tirage aléatoire : l'attaquant gagne avec chanceAttacker% de probabilité
	attackerWins = rand.Float64() < chanceAttacker

	// le perdant perd 100% de ses troupes
	// le gagnant perd entre 0% et chanceAdversaire% de ses troupes (aléatoire)
	chanceLoser := 1.0 - chanceAttacker
	if attackerWins {
		defenderLosses = defending.Count
		attackerLosses = int(math.Round(float64(attacking.Count) * rand.Float64() * chanceLoser))
	} else {
		attackerLosses = attacking.Count
		defenderLosses = int(math.Round(float64(defending.Count) * rand.Float64() * chanceAttacker))
	}

	return
}

func StartPhaseWatcher() {
	lastPhase = GetCurrentPhase()
	savePhaseSnapshot()
	for {
		sleepUntilNextPhase()

		current := GetCurrentPhase()

		if current != lastPhase {
			onPhaseEnd(lastPhase)
			lastPhase = current
			savePhaseSnapshot()
		}
	}
}

func savePhaseSnapshot() {
	// Crée le snapshot depuis l'état en mémoire (gameState)
	countries := gameState.GetCountries()
	data, err := json.MarshalIndent(countries, "", "  ")
	if err != nil {
		println("Erreur snapshot marshal:", err.Error())
		return
	}
	os.WriteFile("server/data/phaseSnapshot.json", data, 0644)
}

func onPhaseEnd(lastPhase string) {
	// Lire l'état courant des pays et des joueurs (protégé par un RLock)
	countries := gameState.GetCountries()
	players := gameState.GetPlayers()

	// Charger le snapshot du début de phase pour détecter les changements
	// (pas besoin de mutex car créé au début de la loope et pas modifié pendant)
	snapshotFile, err := os.Open("server/data/phaseSnapshot.json")
	if err != nil {
		println("Erreur open snapshot:", err.Error())
		return
	}
	defer snapshotFile.Close()

	var oldCountries map[string]CountryInfo
	if err := json.NewDecoder(snapshotFile).Decode(&oldCountries); err != nil {
		println("Erreur decode snapshot:", err.Error())
		return
	}

	// listes d'événements à envoyer au client
	var conquered []ConqueredEvent
	var gainedByAttack []GainedAttackEvent
	var lostAttack []LostAttackEvent
	var lostToWeather []LostWeatherEvent
	var improved []ImprovedEvent
	var goldEarned []GoldEarnedEvent

	// Phase-specific processing
	switch lastPhase {
	case Attack:
		for name, country := range countries {
			if country.AttackedBy == nil {
				continue
			}

			attacker := *country.AttackedBy

			var previousOwner string
			if country.LeaderID != nil {
				previousOwner = *country.LeaderID
			}

			// résoudre le combat
			attackerWins, attackerLosses, defenderLosses := resolveCombat(
				country.TroopsAttacking,
				country.TroopsDefending,
				players,
				attacker,
				previousOwner,
			)

			// appliquer les pertes de troupes aux joueurs
			if p, ok := players[attacker]; ok && country.TroopsAttacking.Type != "" {
				p.Troops[country.TroopsAttacking.Type] -= attackerLosses
				if p.Troops[country.TroopsAttacking.Type] < 0 {
					p.Troops[country.TroopsAttacking.Type] = 0
				}
				players[attacker] = p
			}
			if previousOwner != "" && country.TroopsDefending.Type != "" {
				if p, ok := players[previousOwner]; ok {
					p.Troops[country.TroopsDefending.Type] -= defenderLosses
					if p.Troops[country.TroopsDefending.Type] < 0 {
						p.Troops[country.TroopsDefending.Type] = 0
					}
					players[previousOwner] = p
				}
			}

			if attackerWins {
				// event : attaque réussie
				gainedByAttack = append(gainedByAttack, GainedAttackEvent{
					Country:       name,
					Attacker:      attacker,
					PreviousOwner: previousOwner,
				})

				newLeader := attacker
				country.LeaderID = &newLeader
				country.Level = 0

			} else {
				// event : attaque ratée
				lostAttack = append(lostAttack, LostAttackEvent{
					Country:  name,
					Attacker: attacker,
					Owner:    previousOwner,
				})
			}

			// cleanup
			country.ConqueredBy = nil
			country.AttackedBy = nil
			country.TroopsAttacking = TroopData{}
			country.TroopsDefending = TroopData{}

			countries[name] = country
		}

		// cas d'une défense qui n'a pas servie
		for name, country := range countries {
			if country.AttackedBy != nil {
				continue
			}
			if country.TroopsDefending.Count == 0 {
				continue
			}
			if country.LeaderID == nil {
				continue
			}

			leaderID := *country.LeaderID
			if p, ok := players[leaderID]; ok {
				if p.Troops == nil {
					p.Troops = map[string]int{}
				}
				p.Troops[country.TroopsDefending.Type] += country.TroopsDefending.Count
				players[leaderID] = p
			}

			country.TroopsDefending = TroopData{}
			countries[name] = country
		}

	case Distribution:
		// appliquer les conquêtes en période de paix
		for name, country := range countries {
			if country.ConqueredBy != nil {

				newOwner := *country.ConqueredBy

				country.LeaderID = &newOwner

				// event : conquête
				conquered = append(conquered, ConqueredEvent{
					Country: name,
					By:      newOwner,
				})

				// cleanup
				country.ConqueredBy = nil
				country.AttackedBy = nil

				countries[name] = country
			}
		}

		// distribuer l'or : chaque pays rapporte (level+1) * produced_gold * 1000 à son leader
		// Récupérer les joueurs depuis l'état en mémoire

		goldPerPlayer := map[string]int{}

		for _, country := range countries {
			if country.LeaderID == nil {
				continue
			}
			player, ok := players[*country.LeaderID]
			if !ok {
				continue
			}
			earned := (country.Level + 1) * country.ProducedGold * 1000
			player.Gold += earned
			players[*country.LeaderID] = player
			goldPerPlayer[*country.LeaderID] += earned
		}

		for player, amount := range goldPerPlayer {
			goldEarned = append(goldEarned, GoldEarnedEvent{
				Player: player,
				Amount: amount,
			})
		}

	default:
		println("Incohérence dans onPhaseEnd : phase reçue est \"" + lastPhase + "\"")
	}

	// Appliquer la logique météo (perte de leader et lvl0 si condition contient rain/thunder)
	for name, country := range countries {
		if country.Meteo != nil {
			cond := strings.ToLower(country.Meteo.Condition)

			if strings.Contains(cond, "rain") || strings.Contains(cond, "thunder") {
				if country.LeaderID != nil {

					prevOwner := *country.LeaderID

					lostToWeather = append(lostToWeather, LostWeatherEvent{
						Country:       name,
						PreviousOwner: prevOwner,
						Reason:        country.Meteo.Condition,
					})

					country.LeaderID = nil
					country.Level = 0
					countries[name] = country
				}
			}
		}
	}

	// détecter les pays améliorés (level augmenté)
	for name, newInfo := range countries {
		if old, ok := oldCountries[name]; ok {
			if newInfo.Level > old.Level {

				owner := ""
				if newInfo.LeaderID != nil {
					owner = *newInfo.LeaderID
				}

				improved = append(improved, ImprovedEvent{
					Country:  name,
					Owner:    owner,
					OldLevel: old.Level,
					NewLevel: newInfo.Level,
				})
			}
		}
	}

	// sauvegarder les pays et les joueurs modifiés (protégé par un Lock)
	gameState.SetCountries(countries)
	gameState.SetPlayers(players)

	data, _ := json.MarshalIndent(countries, "", "  ")
	os.WriteFile("server/data/countryInfos.json", data, 0644)

	playerData, _ := json.MarshalIndent(players, "", "  ")
	os.WriteFile("server/data/playerInfos.json", playerData, 0644)

	// construire le popup à envoyer au client
	popup := map[string]any{
		"phase":           lastPhase,
		"conquered":       conquered,
		"gained_attack":   gainedByAttack,
		"lost_attack":     lostAttack,
		"lost_to_weather": lostToWeather,
		"improved":        improved,
		"gold_earned":     goldEarned,
	}

	popupData, err := json.MarshalIndent(popup, "", "  ")
	if err != nil {
		println("Erreur marshal popup:", err.Error())
		return
	}
	if err = os.WriteFile("server/data/phasePopup.json", popupData, 0644); err != nil {
		println("Erreur write popup:", err.Error())
		return
	}
}

func sleepUntilNextPhase() {
	now := time.Now()

	next := now.Truncate(10 * time.Minute).Add(10 * time.Minute)

	time.Sleep(time.Until(next))
}
