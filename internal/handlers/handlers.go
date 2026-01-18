package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/irensaltali/object-storage-gateway/internal/storage"
)

// PutObject handles the PUT /object/{id} endpoint
func PutObject(w http.ResponseWriter, r *http.Request, gateway *storage.Gateway) {
	objectID := mux.Vars(r)["id"]
	log.Printf("PUT /object/%s - received request, content-length: %d", objectID, r.ContentLength)

	// Get content length from request header
	contentLength := r.ContentLength
	if contentLength < 0 {
		log.Printf("PUT /object/%s - error: content-length header is required", objectID)
		http.Error(w, "content-length header is required", http.StatusBadRequest)
		return
	}

	// Store object in gateway
	if err := gateway.PutObject(r.Context(), objectID, r.Body, contentLength); err != nil {
		log.Printf("PUT /object/%s - error storing object: %v", objectID, err)
		statusCode := http.StatusInternalServerError
		http.Error(w, err.Error(), statusCode)
		return
	}

	log.Printf("PUT /object/%s - object stored successfully", objectID)

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
	log.Printf("GET /object/%s - received request", objectID)

	// Retrieve object from gateway
	object, err := gateway.GetObject(r.Context(), objectID)
	if err != nil {
		// Check if it's a "not found" error by checking the error message
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "object not found") {
			log.Printf("GET /object/%s - object not found", objectID)
			http.Error(w, "object not found", http.StatusNotFound)
			return
		}
		log.Printf("GET /object/%s - error retrieving object: %v", objectID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer object.Close()

	log.Printf("GET /object/%s - object found, streaming to client", objectID)

	// Set response headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	// Stream object data to response
	if _, err := io.Copy(w, object); err != nil {
		log.Printf("GET /object/%s - error streaming object: %v", objectID, err)
		return
	}

	log.Printf("GET /object/%s - object streamed successfully", objectID)
}
