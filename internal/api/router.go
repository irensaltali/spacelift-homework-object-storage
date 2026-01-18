package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/irensaltali/object-storage-gateway/internal/handlers"
	"github.com/irensaltali/object-storage-gateway/internal/storage"
)

func NewRouter(gateway *storage.Gateway) *mux.Router {
	router := mux.NewRouter()

	// Object storage endpoints
	router.HandleFunc("/object/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.PutObject(w, r, gateway)
	}).Methods("PUT")
	router.HandleFunc("/object/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetObject(w, r, gateway)
	}).Methods("GET")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	}).Methods("GET")

	// Readiness check endpoint
	router.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ready",
		})
	}).Methods("GET")

	return router
}
