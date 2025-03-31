package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/listener"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/publisher"
	rds "github.com/Bitstarz-eng/event-processing-challenge/internal/redis"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Connect to Redis and start publishing events
	publisher := publisher.NewPublisher()
	wg.Add(1)
	go publisher.StartPublishing(ctx, &wg)

	// Listen localhost/materialized endpoint for statistics
	materialized := listener.NewMaterializedListener(publisher)
	wg.Add(1)
	go materialized.ListenAndServe(&wg)

	// Listen for OS signals (e.g., SIGTERM, SIGINT)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-ctx.Done():
		log.Println("Stop publishing, Context timeout")
		publisher.ShowStats()
	case <-time.After(time.Second * 10):
		log.Println("Processing completed")
	case sig := <-sigChan:
		log.Printf("Received SIGTERM/SIGINT signal: %v\n", sig)
		// Cancel the context to stop the publisher
		cancel()
	}

	// Wait for all Go routines to finish
	wg.Wait()
	rds.Close()
}
