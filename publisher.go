package main

import (
	"bufio"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	cfg := LoadConfig()

	// Connect to a NATS server.
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatalf("Error connecting to NATS at %s: %v", cfg.NatsURL, err)
	}
	defer nc.Close() // Close the connection when main exits just testing git

	log.Println("Connected to NATS server at", nc.ConnectedUrl())

	inputFile := "input.txt" // The file to read messages from.

	// Open the input file
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Error opening input file '%s': %v", inputFile, err)
	}
	defer file.Close()

	log.Printf("Reading messages from '%s' and publishing to subject [%s]", inputFile, cfg.NatsSubject)

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	// Read file line by line
	for scanner.Scan() {
		lineNumber++
		message := scanner.Text()
		if message == "" { // Skip empty lines
			continue
		}

		// Publish the message.
		if err := nc.Publish(cfg.NatsSubject, []byte(message)); err != nil {
			log.Printf("Error publishing message (line %d) '%s': %v", lineNumber, message, err)
		} else {
			log.Printf("Published (line %d) to [%s]: %s", lineNumber, cfg.NatsSubject, message)
		}
		time.Sleep(200 * time.Millisecond) // Optional: slow down publishing a bit
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from input file '%s': %v", inputFile, err)
	}

	if err := nc.FlushTimeout(5 * time.Second); err != nil {
		log.Printf("Error flushing messages: %v", err)
	}

	log.Println("Publisher finished.")
}
