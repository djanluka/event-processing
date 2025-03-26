package subscriber

import (
	"context"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
)

type Subscriber interface {
	Subscribe(context.Context, string)
	HandleEvent(*casino.Event)
	GetStats() interface{}
	ShowStat() // Test purpose
}

const (
	PLAYER_SUB = "PlayerSubscriber"
	GAME_SUB   = "GameSubscriber"
	TIME_SUB   = "TimeSubscriber"
)

func GetSubscribers() map[string]Subscriber {
	return map[string]Subscriber{
		PLAYER_SUB: NewPlayerSubscriber(),
		GAME_SUB:   NewGameSubscriber(),
		TIME_SUB:   NewTimeSubscriber(),
	}
}
