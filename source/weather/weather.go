package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"net/url"
)

var homeDir, _ = os.UserHomeDir()

const (
	apiURL   = "https://api.openweathermap.org/data/2.5/weather"
	apiKey   = "85a4e3c55b73909f42c6a23ec35b7147"
	cacheTTL = 10 * time.Minute
)

type Config struct {
	Location string
	Units    string
	APIKey   string
}

type APIResponse struct {
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Weather []struct {
		Icon string `json:"icon"`
	} `json:"weather"`
	Name string `json:"name"`
}

func Get() string {
	cfg := loadConfig()
	if cfg.Location == "" {
		return ""
	}

	data, err := fetchWeather(cfg)
	if err != nil {
		return ""
	}

	return formatOutput(data)
}

func loadConfig() Config {
	cfgFile := filepath.Join(homeDir, ".config/weather.cfg")
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return Config{}
	}

	cfg := Config{Units: "imperial", APIKey: apiKey}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "location=") {
			cfg.Location = strings.TrimPrefix(line, "location=")
		} else if strings.HasPrefix(line, "units=") {
			cfg.Units = strings.TrimPrefix(line, "units=")
		} else if strings.HasPrefix(line, "api=") {
			cfg.APIKey = strings.TrimPrefix(line, "api=")
		}
	}
	return cfg
}

func getCacheFile() string {
	return filepath.Join(homeDir, ".cache", "openriot-weather.json")
}

func fetchWeather(cfg Config) (*APIResponse, error) {
	// Check cache first
	cacheFile := getCacheFile()
	if data, err := os.ReadFile(cacheFile); err == nil {
		var cached struct {
			Data   APIResponse
			Fresh  time.Time
		}
		if json.Unmarshal(data, &cached) == nil && time.Since(cached.Fresh) < cacheTTL {
			return &cached.Data, nil
		}
	}

	// Fetch from API
	url := fmt.Sprintf("%s?q=%s&units=%s&appid=%s", apiURL, url.QueryEscape(cfg.Location), cfg.Units, cfg.APIKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data APIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	// Cache the response
	cached := struct {
		Data  APIResponse
		Fresh time.Time
	}{data, time.Now()}
	if cacheData, err := json.Marshal(cached); err == nil {
		os.MkdirAll(filepath.Dir(cacheFile), 0755)
		os.WriteFile(cacheFile, cacheData, 0644)
	}

	return &data, nil
}

func formatOutput(data *APIResponse) string {
	if len(data.Weather) == 0 {
		return ""
	}

	icon := getWeatherIcon(data.Weather[0].Icon)
	temp := formatTemp(data.Main.Temp, getUnits())
	return fmt.Sprintf("%s %s", icon, temp)
}

func getWeatherIcon(code string) string {
	switch {
	case strings.HasPrefix(code, "01"): // clear
		return "󰖕"
	case strings.HasPrefix(code, "02"): // few clouds
		return ""
	case strings.HasPrefix(code, "03"), strings.HasPrefix(code, "04"): // clouds
		return ""
	case strings.HasPrefix(code, "09"): // drizzle
		return ""
	case strings.HasPrefix(code, "10"): // rain
		return ""
	case strings.HasPrefix(code, "11"): // thunderstorm
		return ""
	case strings.HasPrefix(code, "13"): // snow
		return ""
	case strings.HasPrefix(code, "50"): // mist/fog/haze
		return "󰖑"
	default:
		return "󰨹"
	}
}

func formatTemp(temp float64, units string) string {
	symbol := "°F"
	if units == "metric" {
		symbol = "°C"
	}
	return fmt.Sprintf("%.0f%s", temp, symbol)
}

func getUnits() string {
	cfgFile := filepath.Join(homeDir, ".config/weather.cfg")
	data, _ := os.ReadFile(cfgFile)
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "units=") {
			return strings.TrimPrefix(line, "units=")
		}
	}
	return "imperial"
}
