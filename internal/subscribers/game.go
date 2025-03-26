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

type GameSubscriber struct {
	RedisClient *redis.Client
	Statistics  map[int]*statistics.GameData
}

func NewGameSubscriber() Subscriber {
	return &GameSubscriber{
		RedisClient: rds.GetRedisClient(),
		Statistics:  make(map[int]*statistics.GameData),
	}
}

func (gs *GameSubscriber) Subscribe(ctx context.Context, event string) {

	fmt.Printf("Game Subscriber subscribed to %s\n", event)
	pubsub := gs.RedisClient.Subscribe(ctx, event)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				// Channel is closed, exit the loop
				fmt.Println("Pub/Sub channel closed")
				return
			}

			// Deserialize the message into an Event struct
			var event casino.Event
			err := json.Unmarshal([]byte(msg.Payload), &event)
			if err != nil {
				log.Printf("Failed to unmarshal event: %v", err)
				continue
			}
			// Handle the event
			gs.HandleEvent(&event)

		case <-ctx.Done():
			// Context is canceled, exit the loop
			log.Println("Game unsubscribing, Context timeout")
			return
		}
	}
}

func (gs *GameSubscriber) HandleEvent(event *casino.Event) {
	gameId := event.GameID
	gd, ok := gs.Statistics[gameId]
	if !ok {
		gd = statistics.NewGameData(gameId)
		gs.Statistics[gameId] = gd
	}

	switch event.Type {
	case casino.GAME_STOP:
		gd.GamePlayedCounter++
		statistics.CalculateMostPlayedGame(gameId, gd.GamePlayedCounter)
	case casino.BET:
		gd.BetPerCurrency[event.Currency] += float64(event.Amount) * casino.SmallestUnit[event.Currency]
		statistics.CalculateMostBettedGame(gameId, event.AmountEUR)
	default:
		break
	}
}

func (gs *GameSubscriber) GetStats() interface{} {
	return nil
}

func (gs *GameSubscriber) ShowStat() {
	fmt.Println("Game Statistics:")
	for _, gd := range gs.Statistics {
		fmt.Printf("%v\n", gd)
	}
}
