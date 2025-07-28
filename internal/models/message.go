package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageDocument represents a message in the system
type MessageDocument struct {
	MessageID  primitive.ObjectID `bson:"_id,omitempty" json:"message_id"`
	LedgerCode int                `bson:"ledger_code" json:"ledger_code"`
	LedgerMtrs string             `bson:"ledger_mtrs" json:"ledger_mtrs"`
	RawMessage string             `bson:"raw_message" json:"raw_message"`
	ReceivedAt time.Time          `bson:"received_at" json:"received_at"`
}
