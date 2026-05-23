package main

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

var lastPhase string

func StartPhaseWatcher() {
	lastPhase = GetCurrentPhase()
	for {
		sleepUntilNextPhase()

		current := GetCurrentPhase()

		if current != lastPhase {
			onPhaseEnd(lastPhase)
			lastPhase = current
		}
	}
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

	// garder une copie pour détecter les changements (améliorations, pertes, ...)
	oldCountries := make(map[string]CountryInfo)
	for k, v := range countries {
		oldCountries[k] = v
	}

	// listes d'événements à envoyer au client
	var lostToWar []string
	var lostToWeather []string
	var conquered []string
	var gainedByAttack []string
	var improved []string

	// Phase-specific processing
	switch lastPhase {
	case Attack:
		// appliquer les attaques : l'attaquant gagne automatiquement
		for name, country := range countries {
			if country.AttackedBy != nil {
				attacker := *country.AttackedBy
				var prevLeader *string
				if country.LeaderID != nil {
					prev := *country.LeaderID
					prevLeader = &prev
				}

				newLeader := attacker
				country.LeaderID = &newLeader

				// cleanup
				country.ConqueredBy = nil
				country.AttackedBy = nil

				countries[name] = country

				gainedByAttack = append(gainedByAttack, name)
				if prevLeader != nil && *prevLeader != newLeader {
					lostToWar = append(lostToWar, name)
				}
			}
		}

	case Distribution:
		// appliquer les conquêtes en période de paix
		for name, country := range countries {
			if country.ConqueredBy != nil {
				newOwner := *country.ConqueredBy
				var prevLeader *string
				if country.LeaderID != nil {
					prev := *country.LeaderID
					prevLeader = &prev
				}

				country.LeaderID = &newOwner

				// cleanup
				country.ConqueredBy = nil
				country.AttackedBy = nil

				countries[name] = country

				conquered = append(conquered, name)
				if prevLeader != nil && *prevLeader != newOwner {
					lostToWar = append(lostToWar, name)
				}
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
					countries[name] = func(c CountryInfo) CountryInfo {
						c.LeaderID = nil
						return c
					}(country)
					lostToWeather = append(lostToWeather, name)
				}
			}
		}
	}

	// détecter les pays améliorés (level augmenté)
	for name, newInfo := range countries {
		if old, ok := oldCountries[name]; ok {
			if newInfo.Level > old.Level {
				improved = append(improved, name)
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
		"phase":            lastPhase,
		"lost_to_war":      lostToWar,
		"lost_to_weather":  lostToWeather,
		"conquered":        conquered,
		"gained_by_attack": gainedByAttack,
		"improved":         improved,
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
