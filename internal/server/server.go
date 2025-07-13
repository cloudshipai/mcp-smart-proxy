// Package server provides HTTP server implementation for the MCP Smart Proxy
package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"mcp-smart-proxy/pkg/types"

	"github.com/gorilla/mux"
)

// Server wraps the smart proxy with HTTP endpoints
type Server struct {
	proxy ProxyInterface
}

// ProxyInterface defines the interface for the smart proxy
type ProxyInterface interface {
	ListTools(ctx context.Context) ([]types.Tool, error)
	DiscoverTools(ctx context.Context, query string) ([]types.Tool, error)
	UseTool(ctx context.Context, toolName string, arguments map[string]interface{}) (map[string]interface{}, error)
	RefreshTools(ctx context.Context) error
	Close() error
}

// New creates a new HTTP server
func New(proxy ProxyInterface) *Server {
	return &Server{proxy: proxy}
}

// handleList returns all available tools
func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	tools, err := s.proxy.ListTools(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.ProxyResponse{RecommendedTools: tools}
	s.writeJSONResponse(w, response)
}

// handleDiscover uses LLM to recommend tools based on a query
func (s *Server) handleDiscover(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var req types.ProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	tools, err := s.proxy.DiscoverTools(ctx, req.Query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.ProxyResponse{RecommendedTools: tools}
	s.writeJSONResponse(w, response)
}

// handleUse executes a specific tool
func (s *Server) handleUse(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	toolName := vars["tool"]

	if toolName == "" {
		http.Error(w, "Tool name is required", http.StatusBadRequest)
		return
	}

	var req types.ToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := s.proxy.UseTool(ctx, toolName, req.Arguments)
	if err != nil {
		response := types.ProxyResponse{Error: err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
		s.writeJSONResponse(w, response)
		return
	}

	response := types.ProxyResponse{Result: result}
	s.writeJSONResponse(w, response)
}

// handleRefresh refreshes the tool cache
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	if err := s.proxy.RefreshTools(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Tools refreshed successfully"))
}

// handleHealth provides a health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// writeJSONResponse writes a JSON response with proper headers
func (s *Server) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// corsMiddleware adds CORS headers to all responses
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Start starts the HTTP server on the specified address
func (s *Server) Start(addr string) error {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/tools", s.handleList).Methods("GET")
	api.HandleFunc("/discover", s.handleDiscover).Methods("POST")
	api.HandleFunc("/use/{tool}", s.handleUse).Methods("POST")
	api.HandleFunc("/refresh", s.handleRefresh).Methods("POST")
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Add CORS middleware
	r.Use(s.corsMiddleware)

	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, r)
}