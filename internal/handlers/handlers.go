package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/irensaltali/object-storage-gateway/internal/storage"
)

// PutObject handles the PUT /object/{id} endpoint
func PutObject(w http.ResponseWriter, r *http.Request, gateway *storage.Gateway) {
	objectID := mux.Vars(r)["id"]

	// Get content length from request header
	contentLength := r.ContentLength
	if contentLength < 0 {
		http.Error(w, "content-length header is required", http.StatusBadRequest)
		return
	}

	// Store object in gateway
	if err := gateway.PutObject(r.Context(), objectID, r.Body, contentLength); err != nil {
		statusCode := http.StatusInternalServerError
		http.Error(w, err.Error(), statusCode)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"id":      objectID,
		"status":  "stored",
		"message": "object stored successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// GetObject handles the GET /object/{id} endpoint
func GetObject(w http.ResponseWriter, r *http.Request, gateway *storage.Gateway) {
	objectID := mux.Vars(r)["id"]

	// Retrieve object from gateway
	object, err := gateway.GetObject(r.Context(), objectID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		http.Error(w, err.Error(), statusCode)
		return
	}

	defer object.Close()

	// Set response headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	// Stream object data to response
	if _, err := io.Copy(w, object); err != nil {
		// Error already partially sent, log and return
		// In production, this would be logged to a proper logger
		return
	}
	w.WriteHeader(http.StatusOK)
}
