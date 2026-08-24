package domain

import "strconv"

type Rating int

func NewRating(ratingInt int) (Rating, error) {
	if ratingInt < 0 {
		return 0, ErrInvalidRating
	}
	return Rating(ratingInt), nil
}

func (r Rating) Int() int {
	return int(r)
}

func (r Rating) String() string {
	return strconv.Itoa(int(r))
}
