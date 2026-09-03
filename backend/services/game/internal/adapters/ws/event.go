package ws

import (
	"encoding/json"
)

type ClientEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type PlayerDTO struct {
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
	Rating   int    `json:"rating"`
}

type GameInitEvent struct {
	Type        string    `json:"type"`
	GameID      string    `json:"game_id"`
	Status      string    `json:"status"`
	WhitePlayer PlayerDTO `json:"white_player"`
	BlackPlayer PlayerDTO `json:"black_player"`
	FEN         string    `json:"fen"`
	Turn        string    `json:"turn"`
	WhiteTimeMs int       `json:"white_time_ms"`
	BlackTimeMs int       `json:"black_time_ms"`
}

type MovePayload struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Promotion string `json:"promotion"`
}

type MoveMadeEvent struct {
	Type        string `json:"type"`
	GameID      string `json:"game_id"`
	Move        string `json:"move"`
	Fen         string `json:"fen"`
	Turn        string `json:"turn"`
	WhiteTimeMs int    `json:"white_time_ms"`
	BlackTimeMs int    `json:"black_time_ms"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
}
