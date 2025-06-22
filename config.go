package main

import "os"

// Config holds the application configuration.
type Config struct {
	NatsURL         string
	NatsSubject     string
	MongoURI        string
	MongoDatabase   string
	MongoCollection string
}

// LoadConfig loads configuration from environment variables with sane defaults.
func LoadConfig() *Config {
	return &Config{
		NatsURL:         getEnv("NATS_URL", "nats://127.0.0.1:4222"),
		NatsSubject:     getEnv("NATS_SUBJECT", "updates"),
		MongoURI:        getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:   getEnv("MONGO_DATABASE", "nats_data"),
		MongoCollection: getEnv("MONGO_COLLECTION", "messages"),
	}
}

// getEnv reads an environment variable or returns a fallback value.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}