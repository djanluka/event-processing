package subscriber

import (
	"context"
	"fmt"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/statistics"
)

type GameSubscriber struct {
	BaseSubscriber *BaseSubscriber
	Statistics     map[int]*statistics.GameData
}

func NewGameSubscriber(name string) Subscriber {
	baseSubscriber := NewBaseSubscriber(name)
	gs := &GameSubscriber{
		BaseSubscriber: baseSubscriber,
		Statistics:     make(map[int]*statistics.GameData),
	}

	gs.BaseSubscriber.EventHandler = gs.HandleEvent
	return gs
}

func (gs *GameSubscriber) Subscribe(ctx context.Context, channel, stopSignal string) {
	gs.BaseSubscriber.Subscribe(ctx, channel, stopSignal)
}

func (gs *GameSubscriber) Unsubscribe(ctx context.Context, channel string) {
	gs.BaseSubscriber.Unsubscribe(ctx, channel)
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
