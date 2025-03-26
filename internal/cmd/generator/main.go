package main

import (
	"log"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/generator"
	"golang.org/x/net/context"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventCh := generator.Generate(ctx)

	for event := range eventCh {
		log.Printf("%#v\n", event)
	}

	log.Println("finished")
}
