package domain

import "shared/timecontrol"

type TimeControlID string

func NewTimeControlID(tcStr string) (TimeControlID, error) {
	if !timecontrol.IsValidTimeControlID(tcStr) {
		return "", ErrInvalidTimeControlID
	}
	return TimeControlID(tcStr), nil
}

func (t TimeControlID) String() string {
	return string(t)
}
