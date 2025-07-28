package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"my-nats-app/internal/config"
	"my-nats-app/internal/db"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	cfg := config.Load()

	// --- MongoDB Setup ---
	mongoClient, err := db.ConnectMongo(cfg.MongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	ctx := context.Background()
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	collection := mongoClient.Database(cfg.MongoDatabase).Collection(cfg.MongoCollection)

	// --- Ledger Code Input ---
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

	deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer deleteCancel()

	deleteResult, err := collection.DeleteMany(deleteCtx, filter)
	if err != nil {
		log.Fatalf("Error deleting documents: %v", err)
	}

	log.Printf("Deletion successful. Number of documents deleted: %d\n", deleteResult.DeletedCount)
}