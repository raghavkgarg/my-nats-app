package main

import (
	"context"
	"fmt"
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

// processMessage parses a raw message, validates it, and transforms it into a MessageDocument.
// It returns a non-nil document if it's valid and should be stored.
// It returns (nil, nil) if the message should be skipped (e.g., filtered ledger code).
// It returns (nil, error) if the message is malformed.
func processMessage(data []byte) (*models.MessageDocument, error) {
	alldata := string(data)

	if len(alldata) < 7 {
		return nil, fmt.Errorf("data '%s' is too short", alldata)
	}

	accStr := alldata[4:7]
	accMtr := alldata[0:4]

	acc, err := strconv.Atoi(accStr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ledger code '%s' to int: %w", accStr, err)
	}

	// Filter out messages with ledger code 123
	if acc == 123 {
		log.Printf("NATS | Filtering out message with ledger code 123: %s", alldata)
		return nil, nil // Not an error, just skipping
	}

	return &models.MessageDocument{
		MessageID:  primitive.NewObjectID(),
		LedgerCode: acc,
		LedgerMtrs: accMtr,
		RawMessage: alldata,
		ReceivedAt: time.Now(),
	}, nil
}

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
		log.Printf("NATS | Received data: '%s'", string(msg.Data))

		messageDocument, err := processMessage(msg.Data)
		if err != nil {
			log.Printf("NATS | Invalid message format: %v. Skipping.", err)
			return
		}

		// If document is nil, it means the message was valid but filtered out.
		if messageDocument == nil {
			return
		}

		// Insert the document into MongoDB
		insertCtx, insertCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer insertCancel()

		insertResult, err := collection.InsertOne(insertCtx, messageDocument)
		if err != nil {
			log.Printf("MongoDB | Error inserting document: %v", err)
		} else if insertResult.InsertedID != nil {
			log.Printf("MongoDB | Inserted document with _id: %v", insertResult.InsertedID)
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
