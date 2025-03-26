package casino

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	GAME_START = "game_start"
	BET        = "bet"
	DEPOSIT    = "deposit"
	GAME_STOP  = "game_stop"
)

var EventTypes = []string{
	GAME_START,
	BET,
	DEPOSIT,
	GAME_STOP,
}

type Event struct {
	ID       int `json:"id"`
	PlayerID int `json:"player_id"`

	// Except for `deposit`.
	GameID int `json:"game_id,omitempty"`

	Type string `json:"type"`

	// Smallest possible unit for the given currency.
	// Examples: 300 = 3.00 EUR, 1 = 0.00000001 BTC.
	// Only for types `bet` and `deposit`.
	Amount int `json:"amount,omitempty"`

	// Only for types `bet` and `deposit`.
	Currency string `json:"currency,omitempty"`

	// Only for type `bet`.
	HasWon bool `json:"has_won,omitempty"`

	CreatedAt time.Time `json:"created_at"`

	AmountEUR   int    `json:"amount_eur,omitempty"`
	Player      Player `json:"player,omitempty"`
	Description string `json:"description"`
}

func (e Event) String() string {
	jsonData, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling Event to JSON:", err)
	}
	return string(jsonData)
}
