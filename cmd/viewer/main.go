package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"my-nats-app/internal/config"
	"my-nats-app/internal/db"
	"my-nats-app/internal/models"

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
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	collection := mongoClient.Database(cfg.MongoDatabase).Collection(cfg.MongoCollection)

	filter := bson.M{}

	log.Printf("Looking up all documents in collection '%s'", cfg.MongoCollection)

	findCtx, findCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer findCancel()

	cursor, err := collection.Find(findCtx, filter)
	if err != nil {
		log.Fatalf("Error finding documents: %v", err)
	}
	defer cursor.Close(findCtx)

	var results []models.MessageDocument
	if err = cursor.All(findCtx, &results); err != nil {
		log.Fatalf("Error decoding documents: %v", err)
	}

	if len(results) == 0 {
		log.Println("No documents found.")
		return
	}

	log.Printf("Found %d documents:", len(results))
	for _, doc := range results {
		fmt.Printf("----------------------------------------\n")
		fmt.Printf("  Message ID: %s\n", doc.MessageID.Hex()) // Changed to print ObjectID
		fmt.Printf("  Ledger Code: %d\n", doc.LedgerCode)
		fmt.Printf("  Ledger Mtrs: %s\n", doc.LedgerMtrs)
		fmt.Printf("  Raw Message: %s\n", doc.RawMessage)
		fmt.Printf("  Received At: %s\n", doc.ReceivedAt.Format(time.RFC3339))
	}
	fmt.Printf("----------------------------------------\n")
}
