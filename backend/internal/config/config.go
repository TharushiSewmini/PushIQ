package config

import (
    "fmt"
    "os"
)

type Config struct {
    Environment    string
    DatabaseURL    string
    APIKey         string
    FCMServerKey   string
    APNSKeyPath    string
    APNSKeyID      string
    APNSTeamID     string
    APNSTopic      string
}

func Load() (*Config, error) {
    cfg := &Config{
        Environment:  defaultEnv("ENVIRONMENT", "development"),
        DatabaseURL:  os.Getenv("DATABASE_URL"),
        APIKey:       os.Getenv("API_KEY"),
        FCMServerKey: os.Getenv("FCM_SERVER_KEY"),
        APNSKeyPath:  os.Getenv("APNS_KEY_PATH"),
        APNSKeyID:    os.Getenv("APNS_KEY_ID"),
        APNSTeamID:   os.Getenv("APNS_TEAM_ID"),
        APNSTopic:    os.Getenv("APNS_TOPIC"),
    }

    if cfg.DatabaseURL == "" {
        return nil, fmt.Errorf("DATABASE_URL is required")
    }
    if cfg.APIKey == "" {
        return nil, fmt.Errorf("API_KEY is required")
    }
    if cfg.FCMServerKey == "" {
        return nil, fmt.Errorf("FCM_SERVER_KEY is required")
    }
    if cfg.APNSKeyPath == "" || cfg.APNSKeyID == "" || cfg.APNSTeamID == "" || cfg.APNSTopic == "" {
        return nil, fmt.Errorf("APNS_KEY_PATH, APNS_KEY_ID, APNS_TEAM_ID, and APNS_TOPIC are required")
    }

    return cfg, nil
}

func defaultEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
