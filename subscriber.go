package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := LoadConfig()

	// --- NATS Setup ---
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatalf("Error connecting to NATS at %s: %v", cfg.NatsURL, err)
	}
	defer nc.Close()

	log.Println("Connected to NATS server at", nc.ConnectedUrl())

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

	var accountIDCounter uint64 = 0 // Simple counter for unique account IDs

	sub, err := nc.Subscribe(cfg.NatsSubject, func(msg *nats.Msg) {
		// Message data processing
		alldata := string(msg.Data)
		log.Printf("NATS | Received data: '%s'", alldata)

		var acc int
		var accMtr string
		if len(alldata) >= 7 {
			accStr := alldata[4:7] // e.g., "abc" from "xxxxabc"
			accMtr = alldata[0:4] // e.g., "xxxx" from "xxxxabc"

			var convErr error
			acc, convErr = strconv.Atoi(accStr)
			if convErr != nil {
				log.Printf("NATS | Failed to convert '%s' to int for acc: %v. Skipping message.", accStr, convErr)
				return // Skip processing this message, don't crash the subscriber
			}
			log.Printf("NATS | Extracted acc: %d", acc)
		} else {
			log.Printf("NATS | Data '%s' too short to extract 'acc'. Skipping message.", alldata)
			return // Skip processing this message
		}

		// Increment account ID for uniqueness in this demo
		currentAccountID := atomic.AddUint64(&accountIDCounter, 1)

		// Prepare data for MongoDB
		if acc != 123 {
			messageDocument := MessageDocument{
				MessageID:  currentAccountID,
				LedgerCode: acc,
				LedgerMtrs: accMtr,
				RawMessage: alldata,
				ReceivedAt: time.Now(),
			}
			// Insert the document into MongoDB using the struct
			insertResult, err := collection.InsertOne(ctx, messageDocument)
			if err != nil {
				log.Printf("MongoDB | Error inserting document for message_id %d: %v", currentAccountID, err)
				return
			}
			if insertResult.InsertedID != nil {
				log.Printf("MongoDB | Inserted document for message_id %d with _id: %v", currentAccountID, insertResult.InsertedID)
			} else {
				log.Printf("MongoDB | Inserted document for message_id %d, but InsertedID is nil", currentAccountID)
			}
		}
	})
	if err != nil {
		log.Fatalf("Error subscribing to NATS subject [%s]: %v", cfg.NatsSubject, err)
	}
	defer sub.Unsubscribe() // Unsubscribe when main exits (though nc.Close() also handles this).

	log.Printf("Subscribed to NATS subject [%s]. Listening for messages...", cfg.NatsSubject)

	// Create a timer for 20 seconds.
	timeout := time.NewTimer(20 * time.Second)
	defer timeout.Stop() // Good practice to stop the timer if we exit for other reasons

	// We'll wait for a SIGINT (Ctrl+C) or SIGTERM to gracefully shut down.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigs:
		log.Println("Shutting down subscriber...")
	case <-timeout.C:
		log.Println("Timeout reached. Shutting down subscriber...")
	}

}
