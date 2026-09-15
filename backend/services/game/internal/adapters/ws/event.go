package ws

import (
	"encoding/json"
)

type clientEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type serverEvent[T any] struct {
	Type    string `json:"type"`
	Payload T      `json:"payload"`
}

type playerDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Rating           int    `json:"rating"`
	Color            string `json:"color"`
	IsOnline         bool   `json:"is_online"`
	TimeMs           int64  `json:"time_ms"`
	TimeoutActive    bool   `json:"timeout_active"`
	TimeoutExpiresAt int64  `json:"timeout_expires_at"`
}

type gameState struct {
	FEN      string   `json:"fen"`
	Turn     string   `json:"turn"`
	SANMoves []string `json:"san_moves"`
	LANMoves []string `json:"lan_moves"`
}

type roomStatePayload struct {
	RoomID    string      `json:"room_id"`
	Status    string      `json:"status"`
	Players   []playerDTO `json:"players"`
	GameState *gameState  `json:"game_state"`
}

type roomStateEvent = serverEvent[roomStatePayload]

type disconnectPayload struct {
	PlayerID         string `json:"player_id"`
	TimeoutExpiresAt int64  `json:"timeout_expires_at"`
}

type disconnectEvent = serverEvent[disconnectPayload]

type connectPayload struct {
	PlayerID string `json:"player_id"`
}

type connectEvent = serverEvent[connectPayload]

type startPayload struct{}

type startEvent = serverEvent[startPayload]

// type MovePayload struct {
// 	From      string `json:"from"`
// 	To        string `json:"to"`
// 	Promotion string `json:"promotion"`
// }

// type MoveMadeEvent struct {
// 	Type        string `json:"type"`
// 	GameID      string `json:"game_id"`
// 	Move        string `json:"move"`
// 	Fen         string `json:"fen"`
// 	Turn        string `json:"turn"`
// 	WhiteTimeMs int    `json:"white_time_ms"`
// 	BlackTimeMs int    `json:"black_time_ms"`
// 	Status      string `json:"status"`
// 	Reason      string `json:"reason"`
// }
