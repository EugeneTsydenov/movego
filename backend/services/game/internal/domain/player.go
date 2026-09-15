package domain

import "time"

type Player struct {
	id            PlayerID
	name          string
	rating        Rating
	color         string
	timeRemaining time.Duration
}

func NewPlayer(id PlayerID, name string, rating Rating, color string, timeRemaining time.Duration) *Player {
	return &Player{
		id:            id,
		name:          name,
		rating:        rating,
		color:         color,
		timeRemaining: timeRemaining,
	}
}

func (p *Player) ID() PlayerID {
	return p.id
}

func (p *Player) Name() string {
	return p.name
}

func (p *Player) Rating() Rating {
	return p.rating
}

func (p *Player) Color() string {
	return p.color
}

func (p *Player) TimeRemaining() time.Duration {
	return p.timeRemaining
}
