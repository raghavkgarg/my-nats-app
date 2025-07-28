package handlers

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"time"

	"my-nats-app/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:embed static
var embedFS embed.FS

// staticFS is a sub-filesystem containing the content of the 'static' directory.
// This makes file paths relative to "static", so "static/index.html" becomes "index.html".
var staticFS fs.FS

func init() {
	var err error
	staticFS, err = fs.Sub(embedFS, "static")
	if err != nil {
		log.Fatalf("failed to create static sub-filesystem: %v", err)
	}
}

type WebHandler struct {
	collection *mongo.Collection
}

func NewWebHandler(collection *mongo.Collection) *WebHandler {
	return &WebHandler{collection: collection}
}

func (h *WebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.handleIndex(w, r)
	case "/store":
		h.handleStore(w, r)
	case "/inquiry-page":
		h.handleInquiryPage(w, r)
	case "/inquiry":
		h.handleInquiry(w, r)
	case "/delete-page":
		h.handleDeletePage(w, r)
	case "/delete":
		h.handleDelete(w, r)
	default:
		// For static assets, we serve from the sub-filesystem.
		// A request for "/style.css" will be served from the root of staticFS.
		http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
	}
}

func (h *WebHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Now we can just look for "index.html" at the root of our sub-filesystem.
	tmpl, err := template.ParseFS(staticFS, "index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// handleInquiryPage serves the inquiry.html page.
func (h *WebHandler) handleInquiryPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(staticFS, "inquiry.html")
	if err != nil {
		http.Error(w, "Could not load inquiry page", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// handleDeletePage serves the delete.html page.
func (h *WebHandler) handleDeletePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(staticFS, "delete.html")
	if err != nil {
		http.Error(w, "Could not load delete page", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func (h *WebHandler) handleStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Robustly parse and validate form values
	ledgerCode, err := strconv.Atoi(r.FormValue("ledger_code"))
	if err != nil {
		http.Error(w, "Invalid 'ledger_code', must be an integer.", http.StatusBadRequest)
		return
	}

	ledgerMtrs := r.FormValue("ledger_mtrs")
	if len(ledgerMtrs) == 0 {
		http.Error(w, "'ledger_mtrs' cannot be empty.", http.StatusBadRequest)
		return
	}

	data := models.MessageDocument{
		MessageID:  primitive.NewObjectID(),
		LedgerCode: ledgerCode,
		LedgerMtrs: ledgerMtrs,
		RawMessage: fmt.Sprintf("%s%d", ledgerMtrs, ledgerCode), // Auto-generate RawMessage
		ReceivedAt: time.Now(),
	}

	// Insert into MongoDB
	_, err = h.collection.InsertOne(r.Context(), data)
	if err != nil {
		log.Printf("Error inserting data: %v", err)
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // Use 201 Created for successful resource creation
	// Encode and send the newly created document back to the client.
	// The `json` tags in the struct will ensure it's formatted correctly.
	json.NewEncoder(w).Encode(data)
}

// handleInquiry searches for records by ledger_code and returns them.
func (h *WebHandler) handleInquiry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get ledger_code from query parameters
	ledgerCodeStr := r.URL.Query().Get("ledger_code")
	if ledgerCodeStr == "" {
		http.Error(w, "Query parameter 'ledger_code' is required.", http.StatusBadRequest)
		return
	}

	ledgerCode, err := strconv.Atoi(ledgerCodeStr)
	if err != nil {
		http.Error(w, "Invalid 'ledger_code', must be an integer.", http.StatusBadRequest)
		return
	}

	// Find documents in MongoDB
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// The BSON field name is `ledger_code`. This is likely defined by a `bson:"ledger_code"`
	// tag on the LedgerCode field in the `models.MessageDocument` struct.
	// Using the wrong field name (e.g., the default 'ledgercode') will result in no documents being found.
	filter := bson.M{"ledger_code": ledgerCode}

	// We sort by ReceivedAt descending to get the newest entries first.
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "receivedat", Value: -1}})

	cursor, err := h.collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("Error finding documents: %v", err)
		http.Error(w, "Failed to query data", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var results []models.MessageDocument
	if err = cursor.All(ctx, &results); err != nil {
		log.Printf("Error decoding documents: %v", err)
		http.Error(w, "Failed to process query results", http.StatusInternalServerError)
		return
	}

	// If no results are found, `results` will be an empty slice `[]`.
	// The frontend JavaScript will handle this case gracefully.

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// handleDelete deletes records by ledger_code and returns the count of deleted records.
func (h *WebHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Parse and validate ledger_code
	ledgerCodeStr := r.FormValue("ledger_code")
	if ledgerCodeStr == "" {
		http.Error(w, "Form field 'ledger_code' is required.", http.StatusBadRequest)
		return
	}

	ledgerCode, err := strconv.Atoi(ledgerCodeStr)
	if err != nil {
		http.Error(w, "Invalid 'ledger_code', must be an integer.", http.StatusBadRequest)
		return
	}

	// Delete documents from MongoDB
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	filter := bson.M{"ledger_code": ledgerCode}

	deleteResult, err := h.collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Printf("Error deleting documents: %v", err)
		http.Error(w, "Failed to delete records", http.StatusInternalServerError)
		return
	}

	// Return success response with deletion count
	response := map[string]interface{}{
		"deleted_count": deleteResult.DeletedCount,
		"ledger_code":   ledgerCode,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
