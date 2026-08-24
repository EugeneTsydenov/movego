package domain

import "time"

func AssignColors(p1, p2 *Player) (white, black *Player) {
	if time.Now().UnixNano()%2 == 0 {
		return p1, p2
	}
	return p2, p1
}
