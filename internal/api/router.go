package api

import (
	"github.com/irensaltali/object-storage-gateway/internal/handlers"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter()

	// User routes
	router.HandleFunc("/object/{id}", handlers.PutObject).Methods("PUT")
	router.HandleFunc("/object/{id}", handlers.GetObject).Methods("GET")

	return router
}
