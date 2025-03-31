package statistics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	rds "github.com/Bitstarz-eng/event-processing-challenge/internal/redis"
	"github.com/go-redis/redis/v8"
)

const (
	TOTAL_EVENTS          = "total_events"
	EVENTS_PER_MINUTE     = "event_per_minute"
	MOVING_AVG_PER_SECOND = "moving_avg_per_second"
)

type TimeStats struct {
	RedisClient        *redis.Client `json:"-"`
	TotalEvents        int           `json:"total_events"`
	EventsPerMinute    int64         `json:"events_per_minute"`
	MovingAvgPerSecond float64       `json:"moving_avg_per_second"`
}

func NewTimeStats() *TimeStats {
	return &TimeStats{
		RedisClient: rds.GetRedisClient(),
	}
}

func (ts *TimeStats) CalculateTimeStats() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	currentTimestamp := float64(time.Now().Unix())

	// Total events
	totalEvents, err := ts.RedisClient.Get(ctx, TOTAL_EVENTS).Int()
	if err != nil && err != redis.Nil {
		log.Fatalf("Error getting total events: %v", err)
	}

	// Events per minute
	eventsPerMinute, err := ts.RedisClient.ZCount(ctx, EVENTS_PER_MINUTE, fmt.Sprintf("%f", currentTimestamp-60), fmt.Sprintf("%f", currentTimestamp)).Result()
	if err != nil {
		log.Fatalf("Error getting events per minute: %v", err)
	}

	// Moving average events per second
	eventsInLastMinute, err := ts.RedisClient.LLen(ctx, MOVING_AVG_PER_SECOND).Result()
	if err != nil {
		log.Fatalf("Error getting events in last minute: %v", err)
	}
	movingAvgPerSecond := math.Round(float64(eventsInLastMinute)/60.*100) / 100

	ts.TotalEvents = totalEvents
	ts.EventsPerMinute = eventsPerMinute
	ts.MovingAvgPerSecond = movingAvgPerSecond
}

// Increment total events
func (ts *TimeStats) IncrementTotalEvents(ctx context.Context) {
	ts.RedisClient.Incr(ctx, TOTAL_EVENTS)
}

// Add timestamp to sorted set
func (ts *TimeStats) AddEventPerMinute(ctx context.Context, timestamp float64, id int) {
	ts.RedisClient.ZAdd(ctx, EVENTS_PER_MINUTE, &redis.Z{
		Score:  timestamp,
		Member: id,
	})
}

// Add timestamp to list and trim to last 60 seconds
func (ts *TimeStats) AddMovingAvgPerSecond(ctx context.Context, timestamp float64) {
	ts.RedisClient.LPush(ctx, MOVING_AVG_PER_SECOND, timestamp)
	ts.RedisClient.LTrim(ctx, MOVING_AVG_PER_SECOND, 0, 59)
}

func (ts *TimeStats) ResetRedisKeys(ctx context.Context) {

	// Reset INCR TOTAL_EVENTS key to 0
	if err := ts.RedisClient.Set(ctx, TOTAL_EVENTS, 0, 0).Err(); err != nil {
		log.Fatalf("Error resetting total events: %v", err)
	}

	// Delete ZADD EVENTS_PER_MINUTE sorted set
	if err := ts.RedisClient.Del(ctx, EVENTS_PER_MINUTE).Err(); err != nil {
		log.Fatalf("Error deleting events per minute: %v", err)
	}

	// Delete LPUSH MOVING_AVG_PER_SECOND list
	if err := ts.RedisClient.Del(ctx, MOVING_AVG_PER_SECOND).Err(); err != nil {
		log.Fatalf("Error deleting event list: %v", err)
	}

	log.Println("Redis keys reset successfully")

}

func (ts *TimeStats) String() string {
	timeStats, err := json.MarshalIndent(ts, "", "  ")
	if err != nil {
		log.Println("Error marshaling TimeData to JSON:", err)
	}
	return string(timeStats)
}
