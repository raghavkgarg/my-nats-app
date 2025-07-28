package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"my-nats-app/internal/config"
	"my-nats-app/internal/db"
	"my-nats-app/internal/models"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	cfg := config.Load()

	// --- NATS Setup ---
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatalf("Error connecting to NATS at %s: %v", cfg.NatsURL, err)
	}
	defer nc.Close()

	log.Println("Connected to NATS server at", nc.ConnectedUrl())

	// --- MongoDB Setup ---
	mongoClient, err := db.ConnectMongo(cfg.MongoURI)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	ctx := context.Background()
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	collection := mongoClient.Database(cfg.MongoDatabase).Collection(cfg.MongoCollection)

	sub, err := nc.Subscribe(cfg.NatsSubject, func(msg *nats.Msg) {
		// Message data processing
		alldata := string(msg.Data)
		log.Printf("NATS | Received data: '%s'", alldata)

		var acc int
		var accMtr string
		if len(alldata) >= 7 {
			accStr := alldata[4:7] // e.g., "abc" from "xxxxabc"
			accMtr = alldata[0:4]  // e.g., "xxxx" from "xxxxabc"

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

		// Prepare data for MongoDB
		if acc != 123 {
			messageDocument := models.MessageDocument{
				MessageID:  primitive.NewObjectID(), // Changed to use ObjectID
				LedgerCode: acc,
				LedgerMtrs: accMtr,
				RawMessage: alldata,
				ReceivedAt: time.Now(),
			}

			// Insert the document into MongoDB using the struct
			insertCtx, insertCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer insertCancel()

			insertResult, err := collection.InsertOne(insertCtx, messageDocument)
			if err != nil {
				log.Printf("MongoDB | Error inserting document: %v", err)
				return
			}
			if insertResult.InsertedID != nil {
				log.Printf("MongoDB | Inserted document with _id: %v", insertResult.InsertedID)
			} else {
				log.Printf("MongoDB | Inserted document, but InsertedID is nil")
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
