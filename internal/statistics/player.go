package statistics

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
)

type PlayerData struct {
	BetCount      atomic.Int64 `json:"bet_count"`
	BetAmount     atomic.Int64 `json:"bet_amount"`
	DepositCount  atomic.Int64 `json:"deposit_count"`
	DepositAmount atomic.Int64 `json:"deposit_amount"`
	WonCount      atomic.Int64 `json:"won_count"`
}

type PlayerStats struct {
	TopPlayerBet     *StatisticCount `json:"top_player_bet"`
	TopPlayerDeposit *StatisticCount `json:"top_player_deposit"`
	TopPlayerWin     *StatisticCount `json:"top_player_win"`
}

var playerStats PlayerStats = PlayerStats{
	TopPlayerBet:     NewStatisticCount(),
	TopPlayerDeposit: NewStatisticCount(),
	TopPlayerWin:     NewStatisticCount(),
}

func NewPlayerData() *PlayerData {
	return &PlayerData{}
}

func (pd *PlayerData) CalculateBetValues(id, amount int) {
	pd.BetCount.Add(1)
	pd.BetAmount.Add(int64(amount))

	if pd.BetCount.Load() > int64(playerStats.TopPlayerBet.Count) {
		playerStats.TopPlayerBet.SetValues(id, int(pd.BetCount.Load()))
	}
}

func (pd *PlayerData) CalculateDepositValues(id, amount int) {
	pd.DepositCount.Add(1)
	pd.DepositAmount.Add(int64(amount))

	if pd.DepositCount.Load() > int64(playerStats.TopPlayerDeposit.Count) {
		playerStats.TopPlayerDeposit.SetValues(id, int(pd.DepositCount.Load()))
	}
}

func (pd *PlayerData) CalculateWonValues(id int) {
	pd.WonCount.Add(1)

	if pd.WonCount.Load() > int64(playerStats.TopPlayerWin.Count) {
		playerStats.TopPlayerWin.SetValues(id, int(pd.WonCount.Load()))
	}
}

func GetStats() interface{} {
	return &playerStats
}

func (pd *PlayerData) String() string {
	response := make(map[string]interface{})
	response["bet_count"] = pd.BetCount.Load()
	response["bet_amount"] = pd.BetAmount.Load()
	response["deposit_count"] = pd.DepositCount.Load()
	response["deposit_amount"] = pd.DepositAmount.Load()
	response["win_count"] = pd.WonCount.Load()

	playerData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling Event to JSON:", err)
	}
	return string(playerData)
}
