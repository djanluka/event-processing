package casino

import (
	"encoding/json"
	"fmt"
	"log"
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

// Set event description field
func (e *Event) SetDescription() {
	playerDesc := e.getPlayerDesc()
	timeDesc := e.getTimeDesc()
	currDesc := e.getCurrDesc()
	gameDesc := e.getGameDesc()

	switch e.Type {
	case GAME_START:
		e.Description = fmt.Sprintf(`%s started playing a game "%s" on %s`, playerDesc, gameDesc, timeDesc)
	case GAME_STOP:
		e.Description = fmt.Sprintf(`%s stopped playing a game "%s" on %s`, playerDesc, gameDesc, timeDesc)
	case BET:
		e.Description = fmt.Sprintf(`%s placed bet of %s on game "%s" on %s`, playerDesc, currDesc, gameDesc, timeDesc)
	case DEPOSIT:
		e.Description = fmt.Sprintf(`%s placed deposit of %s on game "%s" on %s`, playerDesc, currDesc, gameDesc, timeDesc)
	default:
		log.Printf("Unknown event type %v", e.Type)
	}
}

func (e *Event) getPlayerDesc() string {
	if e.Player.IsZero() {
		return fmt.Sprintf("Player ID %d", e.PlayerID)
	}
	return fmt.Sprintf("Player ID %d (%s)", e.PlayerID, e.Player.Email)
}

func (e *Event) getGameDesc() string {
	return Games[e.GameID].Title
}

func (e *Event) getCurrDesc() string {
	amount := SmallestUnit[e.Currency] * float64(e.Amount)
	amountEUR := SmallestUnit["EUR"] * float64(e.AmountEUR)

	if e.Currency == "BTC" {
		return fmt.Sprintf("%.8f %s (%.2f EUR)", amount, e.Currency, amountEUR)
	}
	return fmt.Sprintf("%.2f %s (%.2f EUR)", amount, e.Currency, amountEUR)
}
func (e *Event) getTimeDesc() string {
	t := e.CreatedAt

	day := t.Day()
	var suffix string
	switch day {
	case 1, 21, 31:
		suffix = "st"
	case 2, 22:
		suffix = "nd"
	case 3, 23:
		suffix = "rd"
	default:
		suffix = "th"
	}

	return fmt.Sprintf("%s %d%s, %d at %02d:%02d UTC", t.Month(), t.Day(), suffix, t.Year(), t.Hour(), t.Minute())
}

func (e Event) String() string {
	jsonData, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling Event to JSON:", err)
	}
	return string(jsonData)
}
