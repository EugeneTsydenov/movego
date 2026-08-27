package application

type GameCreatedEvent struct {
	GameID     string   `json:"game_id"`
	PlayersIDs []string `json:"players_ids"`
}
