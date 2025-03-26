package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/db"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/generator"
	rds "github.com/Bitstarz-eng/event-processing-challenge/internal/redis"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/statistics"
	subs "github.com/Bitstarz-eng/event-processing-challenge/internal/subscribers"
	"github.com/go-redis/redis/v8"
)

type Publisher struct {
	RedisClient *redis.Client
	Subscribers map[string]subs.Subscriber
	DB          *db.DB
}

var CASINO_EVENT = "casino_event"

func NewPublisher() *Publisher {
	redisClient := rds.GetRedisClient()
	subscribers := subs.GetSubscribers()
	db := db.GetDB()

	return &Publisher{
		RedisClient: redisClient,
		Subscribers: subscribers,
		DB:          db,
	}
}

func (p *Publisher) StartPublishing(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	eventCh := generator.Generate(ctx)

	go p.startSubscription(ctx)

	for event := range eventCh {
		// Process event data
		p.processEvent(&event)

		// Serialize the event to JSON
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %s", event.String())
		}

		// Publish event
		err = p.RedisClient.Publish(ctx, CASINO_EVENT, eventJSON).Err()
		if err != nil {
			log.Printf("Failed to publish message: %v", err)
		}
		log.Println(event)
	}

	log.Println("Publishing finished")
}

func (p *Publisher) startSubscription(ctx context.Context) {
	var wg sync.WaitGroup
	for _, subscriber := range p.Subscribers {
		wg.Add(1)
		go func(subscriber subs.Subscriber) {
			defer wg.Done()
			subscriber.Subscribe(ctx, CASINO_EVENT)
		}(subscriber)
	}

	wg.Wait()
	log.Println("Succesfully waited for subscribers to finish the work")
}

func (p *Publisher) processEvent(event *casino.Event) {
	// Calculate AmountEUR for BEY and DEPOSIT events
	if event.Type == casino.BET || event.Type == casino.DEPOSIT {
		EUR := casino.Currencies[0]
		if event.Currency == EUR {
			event.AmountEUR = event.Amount
		} else {
			event.AmountEUR = event.Amount
			// event.AmountEUR = int(p.getExchangedValue(event.Currency, EUR, event.Amount))
		}
	}

	// Find the player data
	player, err := p.DB.GetPlayer(event.PlayerID)
	if err != nil {
		log.Printf("Failed to get player data for ID %d: %v", event.PlayerID, err)
		return
	}
	event.Player = *player
}

func (p *Publisher) getExchangedValue(from, to string, amount int) float64 {

	// Check if value is already in cache
	key := from + to
	value, err := p.RedisClient.Get(p.RedisClient.Context(), key).Float64()
	if err == nil {
		fmt.Printf("Found in cache: %f\n", value)
		return value * float64(amount)
	} else if err != redis.Nil {
		log.Fatalf("Error checking Redis cache: %v", err)
	}

	// If not in cache, call the API
	exchangeRateResponse := casino.GetExchangedValueFromApi(from, to, amount)

	if !exchangeRateResponse.Success {
		log.Println("API call was not successful")
	}

	// Store in Redis with TTL of 1 second
	err = p.RedisClient.Set(p.RedisClient.Context(), key, exchangeRateResponse.Info.Quote, 1*time.Second).Err()
	if err != nil {
		log.Fatalf("Error setting Redis key: %v", err)
	}
	fmt.Printf("Stored in cache: %f\n", exchangeRateResponse.Info.Quote)

	return exchangeRateResponse.Result
}

func (p *Publisher) GetStats() interface{} {
	playerStats := p.Subscribers[subs.PLAYER_SUB].GetStats().(*statistics.PlayerStats)
	timeStats := p.Subscribers[subs.TIME_SUB].GetStats().(*statistics.TimeStats)

	// Create combined Stats
	response := make(map[string]interface{})
	response["top_player_bet"] = playerStats.TopPlayerBet
	response["top_player_deposit"] = playerStats.TopPlayerDeposit
	response["top_player_win"] = playerStats.TopPlayerWin
	response["total_events"] = timeStats.TotalEvents
	response["events_per_minute"] = timeStats.EventsPerMinute
	response["moving_avg_per_second"] = timeStats.MovingAvgPerSecond

	return response
}

func (p Publisher) String() string {
	return "Publisher"
}
