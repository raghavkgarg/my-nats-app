package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := LoadConfig()

	// --- MongoDB Setup ---
	// Use a context with a timeout for the connection and operations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Ping MongoDB to verify connection
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Error pinging MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB!")

	collection := mongoClient.Database(cfg.MongoDatabase).Collection(cfg.MongoCollection)

	// Example: Find documents with message_id between 1 and 5 (inclusive)
	// You can adjust this filter as needed, or use an empty filter bson.M{} to get all documents.
	filter := bson.M{
		"message_id": bson.M{
			"$gte": 1, // Greater than or equal to 1
			"$lte": 5, // Less than or equal to 5
		},
	}
	// To find all documents, use: filter := bson.M{}

	log.Printf("Looking up documents in collection '%s' with filter: %v", cfg.MongoCollection, filter)

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatalf("Error finding documents: %v", err)
	}
	defer cursor.Close(ctx)

	var results []MessageDocument
	if err = cursor.All(ctx, &results); err != nil {
		log.Fatalf("Error decoding documents: %v", err)
	}

	if len(results) == 0 {
		log.Println("No documents found matching the filter.")
		return
	}

	log.Printf("Found %d documents:", len(results))
	for _, doc := range results {
		fmt.Printf("----------------------------------------\n")
		fmt.Printf("  Message ID: %d\n", doc.MessageID)
		fmt.Printf("  Ledger Code: %d\n", doc.LedgerCode)
		fmt.Printf("  Ledgermtrs : %s\n", doc.LedgerMtrs)
		fmt.Printf("  Raw Message: %s\n", doc.RawMessage)
		fmt.Printf("  Received At: %s\n", doc.ReceivedAt.Format(time.RFC3339))
	}
	fmt.Printf("----------------------------------------\n")
}
