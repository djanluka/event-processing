package statistics

import "sync"

type StatisticCount struct {
	Mu    sync.Mutex `json:"-"`
	Id    int        `json:"id"`
	Count int        `json:"count"`
}

func NewStatisticCount() *StatisticCount {
	return &StatisticCount{}
}

func (sc *StatisticCount) SetValues(id, count int) {
	sc.Mu.Lock()
	defer sc.Mu.Unlock()
	sc.Id = id
	sc.Count = count
}

type StatisticAmount struct {
	Id     int `json:"id"`
	Amount int `json:"amount"`
}
