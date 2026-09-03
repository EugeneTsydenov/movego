package domain

import (
	"strings"
)

type Move struct {
	from      string
	to        string
	promotion string
}

func NewMove(from, to, promotion string) (Move, error) {
	from = strings.ToLower(strings.TrimSpace(from))
	to = strings.ToLower(strings.TrimSpace(to))
	promotion = strings.ToLower(strings.TrimSpace(promotion))

	if !isValidSquare(from) || !isValidSquare(to) {
		return Move{}, ErrInvalidSquare
	}

	if promotion != "" && !isValidPromotion(promotion) {
		return Move{}, ErrInvalidPromotion
	}

	return Move{
		from:      from,
		to:        to,
		promotion: promotion,
	}, nil
}

func isValidSquare(sq string) bool {
	if len(sq) != 2 {
		return false
	}
	file := sq[0] // letter 'a'-'h'
	rank := sq[1] // digit '1'-'8'
	return file >= 'a' && file <= 'h' && rank >= '1' && rank <= '8'
}

func isValidPromotion(p string) bool {
	return p == "q" || p == "r" || p == "b" || p == "n"
}

func (m Move) From() string      { return m.from }
func (m Move) To() string        { return m.to }
func (m Move) Promotion() string { return m.promotion }
func (m Move) String() string    { return m.from + m.to + m.promotion }
