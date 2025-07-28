package config

import "os"

// Config holds all configuration values
type Config struct {
	MongoURI        string
	MongoDatabase   string
	MongoCollection string
	WebPort         string
	NatsURL         string
	NatsSubject     string
}

// Load returns a new Config instance
func Load() *Config {
	return &Config{
		MongoURI:        getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:   getEnv("MONGO_DB", "messagedb"),
		MongoCollection: getEnv("MONGO_COLLECTION", "messages"),
		WebPort:         getEnv("WEB_PORT", "8080"),
		NatsURL:         getEnv("NATS_URL", "nats://localhost:4222"),
		NatsSubject:     getEnv("NATS_SUBJECT", "messages"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
