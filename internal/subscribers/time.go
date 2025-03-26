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

type TimeSubscriber struct {
	RedisClient *redis.Client
	Statistics  *statistics.TimeStats
}

func NewTimeSubscriber() Subscriber {
	return &TimeSubscriber{
		RedisClient: rds.GetRedisClient(),
		Statistics:  statistics.NewTimeStats(),
	}
}

func (ts *TimeSubscriber) Subscribe(ctx context.Context, event string) {

	fmt.Printf("Time Subscriber subscribed to %s\n", event)
	pubsub := ts.RedisClient.Subscribe(ctx, event)
	defer pubsub.Close()

	ts.ResetRedisKeys(ctx)

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
			ts.HandleEvent(&event)

		case <-ctx.Done():
			// Context is canceled, exit the loop
			log.Println("Time unsubscribing, Context timeout")
			return
		}
	}
}

func (ts *TimeSubscriber) HandleEvent(event *casino.Event) {
	ctx := context.Background()
	timestamp := float64(event.CreatedAt.Unix())

	// Increment total events
	ts.RedisClient.Incr(ctx, statistics.TOTAL_EVENTS)

	// Add timestamp to sorted set
	ts.RedisClient.ZAdd(ctx, statistics.EVENTS_PER_MINUTE, &redis.Z{
		Score:  timestamp,
		Member: event.ID,
	})

	// Add timestamp to list and trim to last 60 seconds
	ts.RedisClient.LPush(ctx, statistics.MOVING_AVG_PER_SECOND, timestamp)
	ts.RedisClient.LTrim(ctx, statistics.MOVING_AVG_PER_SECOND, 0, 59)
}

func (ts *TimeSubscriber) GetStats() interface{} {
	return statistics.CalculateTimeStats(ts.RedisClient)
}

func (ts *TimeSubscriber) ShowStat() {

	tds := statistics.CalculateTimeStats(ts.RedisClient)
	ts.Statistics = tds

	fmt.Printf("Time Statistics:\n%v\n", tds)
}

func (ts *TimeSubscriber) ResetRedisKeys(ctx context.Context) {

	// Reset INCR TOTAL_EVENTS key to 0
	if err := ts.RedisClient.Set(ctx, statistics.TOTAL_EVENTS, 0, 0).Err(); err != nil {
		log.Fatalf("Error resetting total events: %v", err)
	}

	// Delete ZADD EVENTS_PER_MINUTE sorted set
	if err := ts.RedisClient.Del(ctx, statistics.EVENTS_PER_MINUTE).Err(); err != nil {
		log.Fatalf("Error deleting events per minute: %v", err)
	}

	// Delete LPUSH MOVING_AVG_PER_SECOND list
	if err := ts.RedisClient.Del(ctx, statistics.MOVING_AVG_PER_SECOND).Err(); err != nil {
		log.Fatalf("Error deleting event list: %v", err)
	}

	log.Println("Redis keys reset successfully")

}
