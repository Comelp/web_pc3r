package main

import (
	"encoding/json"
	"os"
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
	// ce qui se passe quand une phase finit
	switch lastPhase {

	case Attack:
		{
			// on vient de finir la phase d'attaque, il faut :
			// - regarder parmi tous les pays ceux qui ont un "attacked_by"
			// - appliquer la logique de guerre
			// - mettre à jour les json

			// Pour l'instant on donne juste le pays attaqué au joueur directement

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

			for name, country := range countries {

				if country.AttackedBy != nil {
					attacker := *country.AttackedBy

					// victoire automatique
					country.LeaderID = &attacker

					// cleanup
					country.ConqueredBy = nil
					country.AttackedBy = nil

					countries[name] = country
				}
			}

			data, err := json.MarshalIndent(countries, "", "  ")
			if err != nil {
				println("Erreur marshal countries:", err.Error())
				return
			}

			err = os.WriteFile("server/data/countryInfos.json", data, 0644)
			if err != nil {
				println("Erreur write countries:", err.Error())
				return
			}
		}

	case Distribution:
		{
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

			for name, country := range countries {

				if country.ConqueredBy != nil {
					newOwner := *country.ConqueredBy

					// on réussit toujours une conquête pendant la paix
					country.LeaderID = &newOwner

					// cleanup
					country.ConqueredBy = nil
					country.AttackedBy = nil

					countries[name] = country
				}
			}

			data, err := json.MarshalIndent(countries, "", "  ")
			if err != nil {
				println("Erreur marshal countries:", err.Error())
				return
			}

			err = os.WriteFile("server/data/countryInfos.json", data, 0644)
			if err != nil {
				println("Erreur write countries:", err.Error())
				return
			}
		}

	default:
		println("Incohérence dans onPhaseEnd : phase reçue est \"" + lastPhase + "\"")

	}
}

func sleepUntilNextPhase() {
	now := time.Now()

	next := now.Truncate(10 * time.Minute).Add(10 * time.Minute)

	time.Sleep(time.Until(next))
}
