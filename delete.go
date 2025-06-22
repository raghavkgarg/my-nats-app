package main

import (
	"context"
	"log"
	"os" // Import for accessing command-line arguments
	"strconv" // Import for converting string to int
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := LoadConfig()

	// --- MongoDB Setup ---
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)

	// Use a context with a timeout for the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Error pinging MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB!")

	collection := mongoClient.Database(cfg.MongoDatabase).Collection(cfg.MongoCollection)

	// --- Ledger Code Input ---
	// For this example, we'll hardcode the ledger code.
	// We will now take this from command-line arguments.
	// ledgerCodeToDelete := 123 // <<-- This will be replaced

	// Example for command-line input (uncomment and adapt if needed):
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run delete.go <ledger_code_to_delete>")
	}
	ledgerCodeStr := os.Args[1]
	ledgerCodeToDelete, err := strconv.Atoi(ledgerCodeStr)
	if err != nil {
		log.Fatalf("Invalid ledger code provided: '%s'. Must be an integer. Error: %v", ledgerCodeStr, err)
	}

	log.Printf("Attempting to delete documents with ledger_code: %d from collection '%s'", ledgerCodeToDelete, cfg.MongoCollection)

	// --- Deletion Logic ---
	filter := bson.M{"ledger_code": ledgerCodeToDelete}

	deleteResult, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Fatalf("Error deleting documents: %v", err)
	}

	log.Printf("Deletion successful. Number of documents deleted: %d\n", deleteResult.DeletedCount)
}