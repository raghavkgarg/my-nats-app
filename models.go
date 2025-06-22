package main

import "time"

// MessageDocument defines the structure for messages stored in MongoDB.
type MessageDocument struct {
	MessageID  uint64    `bson:"message_id"`
	LedgerCode int       `bson:"ledger_code"`
	LedgerMtrs string    `bson:"ledger_mtrs"`
	RawMessage string    `bson:"raw_message"`
	ReceivedAt time.Time `bson:"received_at"`
}