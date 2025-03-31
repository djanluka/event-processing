package subscriber

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	rds "github.com/Bitstarz-eng/event-processing-challenge/internal/redis"
	"github.com/go-redis/redis/v8"
)

type Subscriber interface {
	Subscribe(ctx context.Context, channel string, stopSignal string)
	Unsubscribe(ctx context.Context, channel string)
	HandleEvent(*casino.Event)
	GetStats() interface{}
	ShowStat() // Test purpose
}

type BaseSubscriber struct {
	Name         string
	RedisClient  *redis.Client
	PubSub       *redis.PubSub
	EventHandler func(*casino.Event)
}

func NewBaseSubscriber(name string) *BaseSubscriber {
	return &BaseSubscriber{
		Name:        name,
		RedisClient: rds.GetRedisClient(),
	}
}

func (bs *BaseSubscriber) Subscribe(ctx context.Context, channel, stopSignal string) {
	log.Printf("%s subscribed to %s\n", bs.Name, channel)
	bs.PubSub = bs.RedisClient.Subscribe(ctx, channel)
	defer bs.PubSub.Close()

	ch := bs.PubSub.Channel()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				// Channel is closed, exit the loop
				log.Printf("%s: Pub/Sub channel closed", bs.Name)
				return
			}

			// Stop reading the channel
			if msg.Payload == stopSignal {
				bs.Unsubscribe(ctx, channel)
				return
			}

			// Deserialize the message into an Event struct
			var event casino.Event
			err := json.Unmarshal([]byte(msg.Payload), &event)
			if err != nil {
				log.Printf("%s: Failed to unmarshal event: %v", bs.Name, err)
				continue
			}
			// Handle the event
			bs.EventHandler(&event)

		case <-ctx.Done():
			// Context is canceled, exit the loop
			log.Printf("%s: Context timeout", bs.Name)
			return
		}
	}
}

func (bs *BaseSubscriber) Unsubscribe(ctx context.Context, channel string) {
	err := bs.PubSub.Unsubscribe(ctx, channel)
	if err != nil {
		log.Printf("%s: Unsubscribe error: %v", bs.Name, err)
	}
	log.Printf("%s: Unsubscribed", bs.Name)
}

const (
	PLAYER_SUB = "PlayerSubscriber"
	GAME_SUB   = "GameSubscriber"
	TIME_SUB   = "TimeSubscriber"
)

func GetSubscribers() map[string]Subscriber {
	return map[string]Subscriber{
		PLAYER_SUB: NewPlayerSubscriber(PLAYER_SUB),
		GAME_SUB:   NewGameSubscriber(GAME_SUB),
		TIME_SUB:   NewTimeSubscriber(TIME_SUB),
	}
}
