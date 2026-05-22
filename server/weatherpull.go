package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type MeteoData struct {
	Temperature float64 `json:"temperature"`
	Condition   string  `json:"condition"`
}

type ResourcesInfo struct {
	Gold int `json:"gold"`
}

type TroopsInfo struct {
	Planes int `json:"planes"`
}

type CountryWeatherInfo struct {
	LeaderID           *string       `json:"leader_id"`
	Meteo              *MeteoData    `json:"meteo"`
	ProducedRessources ResourcesInfo `json:"produced_ressources"`
	Troops             TroopsInfo    `json:"troops"`
	AttackedBy         *string       `json:"attacked_by"`
}

type WeatherData struct {
	Current struct {
		TempC     float64 `json:"temp_c"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

var countryNameMap = map[string]string{
	"East Germany": "Berlin",
	"West Germany": "Frankfurt",
	"Czechia":      "Prague",
}

const (
	apiKey              = "9b783f757e7142c1b1a90216262205"
	countryInfoFile     = "server/data/countryInfos.json"
	weatherUpdateTicker = 10 * time.Minute
)

func getWeatherForCountry(countryName string) (*MeteoData, error) {
	queryName := countryName
	// Certains pays ont pas le même nom dans l'API (parce qu'on utilise une carte du siècle dernier)
	if mappedName, exists := countryNameMap[countryName]; exists {
		queryName = mappedName
	}

	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", apiKey, queryName)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather api returned status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var weather WeatherData
	if err := json.Unmarshal(body, &weather); err != nil {
		return nil, err
	}

	return &MeteoData{
		Temperature: weather.Current.TempC,
		Condition:   weather.Current.Condition.Text,
	}, nil
}

func updateWeatherData() {
	data, err := os.ReadFile(countryInfoFile)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return
	}

	var countries map[string]CountryWeatherInfo
	if err := json.Unmarshal(data, &countries); err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return
	}

	for countryName := range countries {
		weather, err := getWeatherForCountry(countryName)
		if err != nil {
			log.Printf("Error fetching weather for %s: %v", countryName, err)
			continue
		}

		countryInfo := countries[countryName]
		countryInfo.Meteo = weather
		countries[countryName] = countryInfo
		log.Printf("Updated weather for %s", countryName)
	}

	updatedData, err := json.MarshalIndent(countries, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}

	if err := os.WriteFile(countryInfoFile, updatedData, 0644); err != nil {
		log.Printf("Error writing file: %v", err)
		return
	}

	log.Println("Weather data updated successfully")
}

func StartWeatherPuller() {
	updateWeatherData()
	ticker := time.NewTicker(weatherUpdateTicker)
	defer ticker.Stop()

	for range ticker.C {
		updateWeatherData()
	}
}
