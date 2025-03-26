package statistics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	TOTAL_EVENTS          = "total_events"
	EVENTS_PER_MINUTE     = "event_per_minute"
	MOVING_AVG_PER_SECOND = "moving_avg_per_second"
)

type TimeStats struct {
	TotalEvents        int     `json:"total_events"`
	EventsPerMinute    int64   `json:"events_per_minute"`
	MovingAvgPerSecond float64 `json:"moving_avg_per_second"`
}

func NewTimeStats() *TimeStats {
	return &TimeStats{}
}

func CalculateTimeStats(r *redis.Client) *TimeStats {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	currentTimestamp := float64(time.Now().Unix())

	// Total events
	totalEvents, err := r.Get(ctx, TOTAL_EVENTS).Int()
	if err != nil && err != redis.Nil {
		log.Fatalf("Error getting total events: %v", err)
	}

	// Events per minute
	eventsPerMinute, err := r.ZCount(ctx, EVENTS_PER_MINUTE, fmt.Sprintf("%f", currentTimestamp-60), fmt.Sprintf("%f", currentTimestamp)).Result()
	if err != nil {
		log.Fatalf("Error getting events per minute: %v", err)
	}

	// Moving average events per second
	eventsInLastMinute, err := r.LLen(ctx, MOVING_AVG_PER_SECOND).Result()
	if err != nil {
		log.Fatalf("Error getting events in last minute: %v", err)
	}
	movingAvgPerSecond := float64(eventsInLastMinute) / 60.0

	return &TimeStats{
		TotalEvents:        totalEvents,
		EventsPerMinute:    eventsPerMinute,
		MovingAvgPerSecond: movingAvgPerSecond,
	}

}

func (td *TimeStats) String() string {
	timeData, err := json.MarshalIndent(td, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling TimeData to JSON:", err)
	}
	return string(timeData)
}
