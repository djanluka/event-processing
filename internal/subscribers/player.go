package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	rds "github.com/Bitstarz-eng/event-processing-challenge/internal/redis"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/statistics"
	"github.com/go-redis/redis/v8"
)

type PlayerSubscriber struct {
	RedisClient *redis.Client
	PubSub      *redis.PubSub
	Statistics  map[int]*statistics.PlayerData
}

func NewPlayerSubscriber() Subscriber {
	return &PlayerSubscriber{
		RedisClient: rds.GetRedisClient(),
		Statistics:  make(map[int]*statistics.PlayerData),
	}
}

func (ps *PlayerSubscriber) Subscribe(ctx context.Context, channel, stopSignal string) {

	log.Printf("Player Subscriber subscribed to %s\n", channel)
	ps.PubSub = ps.RedisClient.Subscribe(ctx, channel)
	defer ps.PubSub.Close()

	ch := ps.PubSub.Channel()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				// Channel is closed, exit the loop
				log.Println("Player: Pub/Sub channel closed")
				return
			}

			// Stop reading the channel
			if msg.Payload == stopSignal {
				ps.Unsubscribe(ctx, channel)
				return
			}

			// Deserialize the message into an Event struct
			var event casino.Event
			err := json.Unmarshal([]byte(msg.Payload), &event)
			if err != nil {
				log.Printf("Player: Failed to unmarshal event: %v", err)
				continue
			}
			// Handle the event
			ps.HandleEvent(&event)

		case <-ctx.Done():
			// Context is canceled, exit the loop
			log.Println("Player: Context timeout")
			return
		}
	}
}

func (ps *PlayerSubscriber) Unsubscribe(ctx context.Context, event string) {
	err := ps.PubSub.Unsubscribe(ctx, event)
	if err != nil {
		log.Printf("Player: Unsubscribe error: %v", err)
	}
	log.Println("Player: Unsubscribed")
}

func (ps *PlayerSubscriber) HandleEvent(event *casino.Event) {
	id := event.PlayerID
	spd, ok := ps.Statistics[id]
	if !ok {
		spd = statistics.NewPlayerData()
		ps.Statistics[id] = spd
	}

	switch event.Type {
	case casino.BET:
		spd.CalculateBetValues(id, event.AmountEUR)
	case casino.DEPOSIT:
		spd.CalculateDepositValues(id, event.AmountEUR)
	default:
		break
	}

	if event.HasWon {
		spd.CalculateWonValues(id)
	}

}

func (ps *PlayerSubscriber) GetStats() interface{} {
	return statistics.GetStats()
}

func (ps *PlayerSubscriber) ShowStat() {
	fmt.Println("Player Statistics:")
	for id, pd := range ps.Statistics {
		fmt.Printf("Player %d: %v\n", id, pd)
	}
}
