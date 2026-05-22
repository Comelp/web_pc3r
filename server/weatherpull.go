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

func getWeatherForCountry(countryName string) (*MeteoInfo, error) {
	queryName := countryName
	if mappedName, exists := countryNameMap[countryName]; exists {
		queryName = mappedName
	}

	url := fmt.Sprintf(
		"http://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no",
		apiKey,
		queryName,
	)

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

	return &MeteoInfo{
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

	var countries map[string]CountryInfo
	if err := json.Unmarshal(data, &countries); err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return
	}

	for name, country := range countries {
		weather, err := getWeatherForCountry(name)
		if err != nil {
			log.Printf("Error fetching weather for %s: %v", name, err)
			continue
		}
		country.Meteo = weather
		countries[name] = country
		log.Printf("Updated weather for %s", name)
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
