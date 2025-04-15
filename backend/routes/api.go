package routes

import (
	"github.com/gorilla/mux"
	"net/http"
	"yourapp/handlers"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/api/nodes", handlers.HandleNodes).Methods("GET", "POST")
	r.HandleFunc("/api/nodes/{id}", handlers.DeleteNode).Methods("DELETE")
	r.HandleFunc("/api/stats", handlers.ClusterStats).Methods("GET")

	// WebSocket endpoint
	r.HandleFunc("/ws", handlers.HandleWebSocket)

	return r
}
