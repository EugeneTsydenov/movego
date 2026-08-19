package domain

import "time"

type TimeControlID string

const (
	// Bullet
	TimeControl1Min     TimeControlID = "1+0"
	TimeControl2Min1Sec TimeControlID = "2+1"

	// Blitz
	TimeControl3Min     TimeControlID = "3+0"
	TimeControl3Min2Sec TimeControlID = "3+2"
	TimeControl5Min     TimeControlID = "5+0"

	// Rapid
	TimeControl10Min      TimeControlID = "10+0"
	TimeControl15Min10Sec TimeControlID = "15+10"

	// Classical
	TimeControl30Min TimeControlID = "30+0"
)

type TimeCategory string

const (
	CategoryBullet    TimeCategory = "BULLET"
	CategoryBlitz     TimeCategory = "BLITZ"
	CategoryRapid     TimeCategory = "RAPID"
	CategoryClassical TimeCategory = "CLASSICAL"
)

type TimeControl struct {
	id        TimeControlID
	category  TimeCategory
	initial   time.Duration
	increment time.Duration
}

var presets = map[TimeControlID]TimeControl{
	// Bullet
	TimeControl1Min: {
		id:        TimeControl1Min,
		category:  CategoryBullet,
		initial:   1 * time.Minute,
		increment: 0,
	},
	TimeControl2Min1Sec: {
		id:        TimeControl2Min1Sec,
		category:  CategoryBullet,
		initial:   2 * time.Minute,
		increment: 1 * time.Second,
	},

	// Blitz
	TimeControl3Min: {
		id:        TimeControl3Min,
		category:  CategoryBlitz,
		initial:   3 * time.Minute,
		increment: 0,
	},
	TimeControl3Min2Sec: {
		id:        TimeControl3Min2Sec,
		category:  CategoryBlitz,
		initial:   3 * time.Minute,
		increment: 2 * time.Second,
	},
	TimeControl5Min: {
		id:        TimeControl5Min,
		category:  CategoryBlitz,
		initial:   5 * time.Minute,
		increment: 0,
	},

	// Rapid
	TimeControl10Min: {
		id:        TimeControl10Min,
		category:  CategoryRapid,
		initial:   10 * time.Minute,
		increment: 0,
	},
	TimeControl15Min10Sec: {
		id:        TimeControl15Min10Sec,
		category:  CategoryRapid,
		initial:   15 * time.Minute,
		increment: 10 * time.Second,
	},

	// Classical
	TimeControl30Min: {
		id:        TimeControl30Min,
		category:  CategoryClassical,
		initial:   30 * time.Minute,
		increment: 0,
	},
}

func NewTimeControl(id string) (TimeControl, error) {
	preset, exists := presets[TimeControlID(id)]
	if !exists {
		return TimeControl{}, ErrInvalidTimeControl
	}
	return preset, nil
}

func (t TimeControl) ID() TimeControlID      { return t.id }
func (t TimeControl) Category() TimeCategory { return t.category }
func (t TimeControl) Duration() (time.Duration, time.Duration) {
	return t.initial, t.increment
}
