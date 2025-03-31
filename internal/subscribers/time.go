package subscriber

import (
	"context"
	"fmt"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/statistics"
	"github.com/go-redis/redis/v8"
)

type TimeSubscriber struct {
	BaseSubscriber *BaseSubscriber
	Statistics     *statistics.TimeStats
}

func NewTimeSubscriber(name string) Subscriber {
	baseSubscriber := NewBaseSubscriber(name)
	ts := &TimeSubscriber{
		BaseSubscriber: baseSubscriber,
		Statistics:     statistics.NewTimeStats(),
	}

	ts.BaseSubscriber.EventHandler = ts.HandleEvent
	return ts
}

func (ts *TimeSubscriber) Subscribe(ctx context.Context, channel, stopSignal string) {
	ts.Statistics.ResetRedisKeys(ctx)
	ts.BaseSubscriber.Subscribe(ctx, channel, stopSignal)
}

func (ts *TimeSubscriber) Unsubscribe(ctx context.Context, channel string) {
	ts.BaseSubscriber.Unsubscribe(ctx, channel)
}

func (ts *TimeSubscriber) HandleEvent(event *casino.Event) {
	ctx := context.Background()
	timestamp := float64(event.CreatedAt.Unix())

	// Increment total events
	ts.BaseSubscriber.RedisClient.Incr(ctx, statistics.TOTAL_EVENTS)

	// Add timestamp to sorted set
	ts.BaseSubscriber.RedisClient.ZAdd(ctx, statistics.EVENTS_PER_MINUTE, &redis.Z{
		Score:  timestamp,
		Member: event.ID,
	})

	// Add timestamp to list and trim to last 60 seconds
	ts.BaseSubscriber.RedisClient.LPush(ctx, statistics.MOVING_AVG_PER_SECOND, timestamp)
	ts.BaseSubscriber.RedisClient.LTrim(ctx, statistics.MOVING_AVG_PER_SECOND, 0, 59)
}

func (ts *TimeSubscriber) GetStats() interface{} {
	return ts.Statistics.CalculateTimeStats()
}

func (ts *TimeSubscriber) ShowStat() {
	tds := ts.Statistics.CalculateTimeStats()
	ts.Statistics = tds

	fmt.Printf("Time Statistics:\n%v\n", tds)
}
