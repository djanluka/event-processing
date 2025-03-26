package listener

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/publisher"
)

type Materialized struct {
	Publisher *publisher.Publisher
}

func NewMaterializedListener(p *publisher.Publisher) *Materialized {
	return &Materialized{
		Publisher: p,
	}
}

func (m *Materialized) materializedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		stats := m.Publisher.GetStats()

		response, err := json.Marshal(stats)
		if err != nil {
			http.Error(w, "Failed to marshal combined JSON", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func (m *Materialized) ListenAndServe(wg *sync.WaitGroup) {
	defer wg.Done()

	http.HandleFunc("/materialized", m.materializedHandler)

	// Create an HTTP server
	server := &http.Server{
		Addr:    ":8080", // Listen on port 8080
		Handler: nil,     // Use the default ServeMux
	}

	// Start the server in a goroutine
	go func() {
		log.Println("Starting HTTP server on port 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v\n", err)
		}
	}()

	// Listen for SIGTERM and SIGINT signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Wait for a shutdown signal
	sig := <-sigChan
	log.Printf("Listener received SIGTERM/SIGINT signal: %v\n", sig)

	// Create a context with a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	log.Println("Shutting down HTTP server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v\n", err)
	} else {
		log.Println("HTTP server shut down gracefully")
	}
}
