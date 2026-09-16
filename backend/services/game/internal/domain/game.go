package domain

import (
	"time"

	"github.com/corentings/chess/v2"
	"github.com/google/uuid"
)

type GameStatus string

const (
	StatusCreated    GameStatus = "CREATED"
	StatusInProgress GameStatus = "IN_PROGRESS"
	StatusCompleted  GameStatus = "COMPLETED"
	StatusAborted    GameStatus = "ABORTED"
)

func (s GameStatus) String() string {
	return string(s)
}

type Game struct {
	*chess.Game
	id            GameID
	players       []*Player
	status        GameStatus
	turnStartedAt time.Time
	timeControl   TimeControl
	createdAt     time.Time
	updatedAt     time.Time
	finishedAt    time.Time
}

func NewGame(players []*Player, timeControl TimeControl) *Game {
	now := time.Now().UTC()
	return &Game{
		id:            GameID(uuid.Must(uuid.NewV7())),
		Game:          chess.NewGame(),
		players:       players,
		status:        StatusCreated,
		turnStartedAt: now,
		timeControl:   timeControl,
		createdAt:     now,
		updatedAt:     now,
	}
}

func RestoreGame(
	id GameID,
	players []*Player,
	status GameStatus,
	timeControl TimeControl,
	fen string,
	moves []string,
	createdAt time.Time,
	updatedAt time.Time,
	finishedAt time.Time,
) (*Game, error) {
	var chessGame *chess.Game
	if fen != "" {
		fn, err := chess.FEN(fen)
		if err != nil {
			return nil, err
		}
		chessGame = chess.NewGame(fn)
	} else {
		chessGame = chess.NewGame()
		for _, m := range moves {
			if err := chessGame.PushNotationMove(m, chess.AlgebraicNotation{}, nil); err != nil {
				return nil, err
			}
		}
	}

	return &Game{
		Game:        chessGame,
		id:          id,
		players:     players,
		status:      status,
		timeControl: timeControl,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		finishedAt:  finishedAt,
	}, nil
}

func (g *Game) ID() GameID {
	return g.id
}

func (g *Game) Players() []*Player {
	return g.players
}

func (g *Game) WhitePlayer() *Player {
	for _, p := range g.players {
		if p.color == "w" {
			return p
		}
	}

	return nil
}

func (g *Game) BlackPlayer() *Player {
	for _, p := range g.players {
		if p.color == "b" {
			return p
		}
	}

	return nil
}

func (g *Game) Status() GameStatus {
	return g.status
}

func (g *Game) TimeControl() TimeControl {
	return g.timeControl
}

func (g *Game) CreatedAt() time.Time {
	return g.createdAt
}

func (g *Game) UpdatedAt() time.Time {
	return g.updatedAt
}

func (g *Game) FinishedAt() time.Time {
	return g.finishedAt
}

func (g *Game) IsPlayer(playerID PlayerID) bool {
	for _, r := range g.players {
		if r.id == playerID {
			return true
		}
	}

	return false
}

func (g *Game) EnsurePlayer(id PlayerID) error {
	if !g.IsPlayer(id) {
		return ErrPlayerNotInGame
	}
	return nil
}

func (g *Game) Start(now time.Time) {
	g.status = StatusInProgress
	g.turnStartedAt = now
	g.updatedAt = now
}

func (g *Game) IsClockRunning() bool {
	return g.status == StatusInProgress && !g.turnStartedAt.IsZero()
}

func (g *Game) MakeMove(playerID PlayerID, move Move) error {
	if err := g.EnsurePlayer(playerID); err != nil {
		return err
	}

	if err := g.PushNotationMove(move.String(), chess.LongAlgebraicNotation{}, nil); err != nil {
		return ErrIllegalMove
	}

	return nil
}

func (g *Game) MoveLANHistoryStrings() []string {
	moves := g.Moves()
	history := make([]string, len(moves))

	for i, move := range moves {
		history[i] = move.String()
	}

	return history
}

func (g *Game) MoveSANHistoryStrings() []string {
	moves := g.Moves()
	positions := g.Positions()
	encoder := chess.AlgebraicNotation{}

	history := make([]string, len(moves))
	for i, move := range moves {
		history[i] = encoder.Encode(positions[i], move)
	}

	return history
}
func (g *Game) PlayerIDStrings() []string {
	ids := make([]string, len(g.players))
	for i, p := range g.players {
		ids[i] = p.id.String()
	}
	return ids
}
