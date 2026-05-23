package main

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

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
	data, err := os.ReadFile("server/data/countryInfos.json")
	if err != nil {
		println("Erreur snapshot:", err.Error())
		return
	}
	os.WriteFile("server/data/phaseSnapshot.json", data, 0644)
}

func onPhaseEnd(lastPhase string) {
	// Lire l'état courant des pays
	countryFile, err := os.Open("server/data/countryInfos.json")
	if err != nil {
		println("Erreur open country file:", err.Error())
		return
	}
	defer countryFile.Close()

	var countries map[string]CountryInfo
	if err := json.NewDecoder(countryFile).Decode(&countries); err != nil {
		println("Erreur decode countries:", err.Error())
		return
	}

	// Charger le snapshot du début de phase pour détecter les changements
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

	// Phase-specific processing
	switch lastPhase {
	case Attack:
		// appliquer les attaques : l'attaquant gagne automatiquement
		for name, country := range countries {
			if country.AttackedBy != nil {

				attacker := *country.AttackedBy

				var previousOwner string
				if country.LeaderID != nil {
					previousOwner = *country.LeaderID
				}

				// event : attaque réussie
				gainedByAttack = append(gainedByAttack, GainedAttackEvent{
					Country:       name,
					Attacker:      attacker,
					PreviousOwner: previousOwner,
				})

				newLeader := attacker
				country.LeaderID = &newLeader

				// cleanup
				country.ConqueredBy = nil
				country.AttackedBy = nil
				country.TroopsAttacking = TroopData{}
				country.TroopsDefending = TroopData{}

				countries[name] = country
			}
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

	default:
		println("Incohérence dans onPhaseEnd : phase reçue est \"" + lastPhase + "\"")
	}

	// Appliquer la logique météo (perte de leader si condition contient rain/thunder)
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
					countries[name] = country
				}
			}
		}
	}

	// détecter les pays améliorés via le snapshot (level augmenté)
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

	// sauvegarder les pays modifiés
	data, err := json.MarshalIndent(countries, "", "  ")
	if err != nil {
		println("Erreur marshal countries:", err.Error())
		return
	}
	if err = os.WriteFile("server/data/countryInfos.json", data, 0644); err != nil {
		println("Erreur write countries:", err.Error())
		return
	}

	// construire le popup à envoyer au client
	popup := map[string]any{
		"phase":           lastPhase,
		"conquered":       conquered,
		"gained_attack":   gainedByAttack,
		"lost_attack":     lostAttack,
		"lost_to_weather": lostToWeather,
		"improved":        improved,
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
