package statistics

import (
	"encoding/json"
	"log"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
)

var (
	mostPlayedGame StatisticCount
	mostBettedGame StatisticAmount
)

type GameData struct {
	Id                int                `json:"id"`
	Name              string             `json:"name"`
	GamePlayedCounter int                `json:"game_played_count"`
	BetPerCurrency    map[string]float64 `json:"bet_per_currency"`
}

func NewGameData(id int) *GameData {
	return &GameData{
		Id:                id,
		Name:              casino.Games[id].Title,
		GamePlayedCounter: 0,
		BetPerCurrency:    make(map[string]float64),
	}
}

func CalculateMostPlayedGame(gameId, counter int) {
	if counter > mostPlayedGame.Count {
		mostPlayedGame = StatisticCount{
			Id:    gameId,
			Count: counter,
		}
	}
}

func CalculateMostBettedGame(gameId, amount int) {
	if amount > mostBettedGame.Amount {
		mostBettedGame = StatisticAmount{
			Id:     gameId,
			Amount: amount,
		}
	}
}

func GetMostPlayedGame() StatisticCount {
	return mostPlayedGame
}

func GetMostBettedGame() StatisticAmount {
	return mostBettedGame
}

func (gd *GameData) String() string {
	gameData, err := json.MarshalIndent(gd, "", "  ")
	if err != nil {
		log.Println("Error marshaling GameData to JSON:", err)
	}
	return string(gameData)
}
