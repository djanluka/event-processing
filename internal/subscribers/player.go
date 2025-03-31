package subscriber

import (
	"context"
	"fmt"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/statistics"
)

type PlayerSubscriber struct {
	BaseSubscriber *BaseSubscriber
	Statistics     map[int]*statistics.PlayerData
}

func NewPlayerSubscriber(name string) Subscriber {
	baseSubscriber := NewBaseSubscriber(name)
	ps := &PlayerSubscriber{
		BaseSubscriber: baseSubscriber,
		Statistics:     make(map[int]*statistics.PlayerData),
	}

	ps.BaseSubscriber.EventHandler = ps.HandleEvent
	return ps
}

func (ps *PlayerSubscriber) Subscribe(ctx context.Context, channel, stopSignal string) {
	ps.BaseSubscriber.Subscribe(ctx, channel, stopSignal)
}

func (ps *PlayerSubscriber) Unsubscribe(ctx context.Context, channel string) {
	ps.BaseSubscriber.Unsubscribe(ctx, channel)
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
