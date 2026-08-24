package domain

import (
	"math"
	"time"
)

type Player struct {
	id            PlayerID
	name          string
	timeControlID TimeControlID
	rating        Rating
	joinedAt      time.Time
}

func NewPlayer(id PlayerID, name string, timeControlID TimeControlID, rating Rating) *Player {
	return &Player{
		id:            id,
		name:          name,
		timeControlID: timeControlID,
		rating:        rating,
		joinedAt:      time.Now(),
	}
}

func RestorePlayer(id PlayerID, name string, timeControlID TimeControlID, rating Rating, joinedAt time.Time) *Player {
	return &Player{
		id:            id,
		name:          name,
		timeControlID: timeControlID,
		rating:        rating,
		joinedAt:      joinedAt,
	}
}

func (p *Player) ID() PlayerID {
	return p.id
}

func (p *Player) Name() string {
	return p.name
}

func (p *Player) TimeControlID() TimeControlID {
	return p.timeControlID
}

func (p *Player) Rating() Rating {
	return p.rating
}

func (p *Player) JoinedAt() time.Time {
	return p.joinedAt
}

func (p *Player) WaitDuration() time.Duration {
	return time.Since(p.joinedAt)
}

func (p *Player) CanMatch(opponent *Player) bool {
	if p.timeControlID != opponent.timeControlID {
		return false
	}

	if p.id == opponent.id {
		return false
	}

	effectiveWindow := int(math.Max(float64(p.CalculateRatingWindow()), float64(opponent.CalculateRatingWindow())))

	ratingDiff := int(math.Abs(float64(p.rating - opponent.rating)))

	return ratingDiff <= effectiveWindow
}

func (p *Player) CalculateRatingWindow() int {
	seconds := p.WaitDuration().Seconds()
	if seconds < 50 {
		return 50
	}

	expansion := int(seconds/5) * 50
	return 50 + expansion
}
