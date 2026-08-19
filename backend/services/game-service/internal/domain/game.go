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

type Game struct {
	*chess.Game
	id                 uuid.UUID
	whitePlayer        *Player
	blackPlayer        *Player
	status             GameStatus
	whiteTimeRemaining time.Duration
	blackTimeRemaining time.Duration
	timeControl        TimeControl
	createdAt          time.Time
	updatedAt          time.Time
	finishedAt         time.Time
}

func NewGame(whitePlayer, blackPlayer *Player, timeControl TimeControl) *Game {
	now := time.Now().UTC()
	initialTime, _ := timeControl.Duration()
	return &Game{
		id:                 uuid.Must(uuid.NewV7()),
		Game:               chess.NewGame(),
		whitePlayer:        whitePlayer,
		blackPlayer:        blackPlayer,
		status:             StatusInProgress,
		whiteTimeRemaining: initialTime,
		blackTimeRemaining: initialTime,
		timeControl:        timeControl,
		createdAt:          now,
		updatedAt:          now,
	}
}

func RestoreGame(
	id uuid.UUID,
	whitePlayer *Player,
	blackPlayer *Player,
	status GameStatus,
	whiteTimeRemaining time.Duration,
	blackTimeRemaining time.Duration,
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
		Game:               chessGame,
		id:                 id,
		whitePlayer:        whitePlayer,
		blackPlayer:        blackPlayer,
		status:             status,
		whiteTimeRemaining: whiteTimeRemaining,
		blackTimeRemaining: blackTimeRemaining,
		timeControl:        timeControl,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
		finishedAt:         finishedAt,
	}, nil
}

func (g *Game) ID() uuid.UUID {
	return g.id
}

func (g *Game) WhitePlayer() *Player {
	return g.whitePlayer
}

func (g *Game) BlackPlayer() *Player {
	return g.blackPlayer
}

func (g *Game) Status() GameStatus {
	return g.status
}

func (g *Game) WhiteTimeRemaining() time.Duration {
	return g.whiteTimeRemaining
}

func (g *Game) BlackTimeRemaining() time.Duration {
	return g.blackTimeRemaining
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

func (g *Game) IsPlayer(playerID uuid.UUID) bool {
	return g.whitePlayer.ID() == playerID || g.blackPlayer.ID() == playerID
}
