package config

import (
	"os"
	"strconv"
)

// Config holds all runtime configuration for the clock publisher.
type Config struct {
	// MQTT broker address, e.g. "tcp://192.168.1.100:1883"
	BrokerURL string

	// AWTRIX device prefix — the topic prefix your device uses.
	// Default is "awtrix" but can be changed on the device itself.
	DevicePrefix string

	// How often to push the clock payload (seconds). 1 is fine for a live clock.
	// AWTRIX also has a built-in clock app; this gives you full control over
	// the format, colour, and layout.
	TickSeconds int

	// MQTT client ID
	ClientID string

	// Optional MQTT auth
	Username string
	Password string
}

func Load() Config {
	return Config{
		BrokerURL:    getEnv("MQTT_BROKER_URL", "tcp://localhost:1883"),
		DevicePrefix: getEnv("AWTRIX_PREFIX", "awtrix"),
		TickSeconds:  getEnvInt("TICK_SECONDS", 1),
		ClientID:     getEnv("MQTT_CLIENT_ID", "pixelclock-publisher"),
		Username:     getEnv("MQTT_USERNAME", ""),
		Password:     getEnv("MQTT_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
