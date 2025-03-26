package casino

import (
	"encoding/json"
	"fmt"
	"time"
)

type Player struct {
	Email          string    `json:"email"`
	LastSignedInAt time.Time `json:"last_signed_in_at"`
}

func (p Player) IsZero() bool {
	return p.Email == "" || p.LastSignedInAt.IsZero()
}

func (p Player) String() string {
	jsonData, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling Player to JSON:", err)
	}

	return string(jsonData)
}
