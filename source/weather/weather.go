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

var cachedConfig *Config

const (
	apiURL   = "https://api.openweathermap.org/data/2.5/weather"
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
	if cfg.Location == "" || cfg.APIKey == "" {
		return ""
	}

	data, err := fetchWeather(cfg)
	if err != nil {
		return ""
	}

	return formatOutput(data)
}

func loadConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}
	}
	cfgFile := filepath.Join(homeDir, ".config", "weather.cfg")
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return Config{}
	}

	cfg := Config{Units: "imperial"}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "location="); ok {
			cfg.Location = after
		} else if after, ok := strings.CutPrefix(line, "units="); ok {
			cfg.Units = after
		} else if after, ok := strings.CutPrefix(line, "api="); ok {
			cfg.APIKey = after
		}
	}
	return cfg
}

func getCacheFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	cacheDir := filepath.Join(homeDir, ".cache", "openriot")
	newFile := filepath.Join(cacheDir, "weather.json")
	oldFile := filepath.Join(homeDir, ".cache", "openriot-weather.json")
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		if data, err := os.ReadFile(oldFile); err == nil {
			_ = os.MkdirAll(cacheDir, 0o700)
			_ = os.WriteFile(newFile, data, 0o600)
			_ = os.Remove(oldFile)
		}
	}
	return newFile
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
	client := &http.Client{Timeout: 10 * time.Second}
	fetchURL := fmt.Sprintf("%s?q=%s&units=%s&appid=%s", apiURL, url.QueryEscape(cfg.Location), cfg.Units, cfg.APIKey)
	resp, err := client.Get(fetchURL)
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
		os.MkdirAll(filepath.Dir(cacheFile), 0700)
		os.WriteFile(cacheFile, cacheData, 0600)
	}

	return &data, nil
}

func formatOutput(data *APIResponse) string {
	if len(data.Weather) == 0 {
		return ""
	}

	icon := getWeatherIcon(data.Weather[0].Icon)
	temp := formatTemp(data.Main.Temp, getUnits())
	return fmt.Sprintf("%s %s", temp, icon)
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
	if cachedConfig == nil {
		cfg := loadConfig()
		cachedConfig = &cfg
	}
	return cachedConfig.Units
}
