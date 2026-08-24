package domain

type Player struct {
	id     PlayerID
	name   string
	rating Rating
}

func NewPlayer(id PlayerID, name string, rating Rating) *Player {
	return &Player{
		id:     id,
		name:   name,
		rating: rating,
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
