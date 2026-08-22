package domain

import "github.com/google/uuid"

type Player struct {
	id   uuid.UUID
	name string
}

func NewPlayer(id uuid.UUID, name string) *Player {
	return &Player{
		id:   id,
		name: name,
	}
}

func (p *Player) ID() uuid.UUID {
	return p.id
}

func (p *Player) Name() string {
	return p.name
}
