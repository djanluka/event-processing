package subscriber

import (
	"context"
	"fmt"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/statistics"
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
	ts.Statistics.IncrementTotalEvents(ctx)
	// Add timestamp to sorted set
	ts.Statistics.AddEventPerMinute(ctx, timestamp, event.ID)
	// Add timestamp to list and trim to last 60 seconds
	ts.Statistics.AddMovingAvgPerSecond(ctx, timestamp)
}

func (ts *TimeSubscriber) GetStats() interface{} {
	return ts.Statistics
}

func (ts *TimeSubscriber) ShowStat() {
	ts.Statistics.CalculateTimeStats()
	fmt.Printf("Time Statistics:\n%v\n", ts.Statistics)
}
