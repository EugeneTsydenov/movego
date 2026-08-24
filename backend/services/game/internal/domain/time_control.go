package domain

import (
	"shared/timecontrol"
	"time"
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
	timecontrol.TIME_CONTROL_1_0: {
		id:        timecontrol.TIME_CONTROL_1_0,
		category:  CategoryBullet,
		initial:   1 * time.Minute,
		increment: 0,
	},
	timecontrol.TIME_CONTROL_2_1: {
		id:        timecontrol.TIME_CONTROL_2_1,
		category:  CategoryBullet,
		initial:   2 * time.Minute,
		increment: 1 * time.Second,
	},

	// Blitz
	timecontrol.TIME_CONTROL_3_0: {
		id:        timecontrol.TIME_CONTROL_3_0,
		category:  CategoryBlitz,
		initial:   3 * time.Minute,
		increment: 0,
	},
	timecontrol.TIME_CONTROL_3_2: {
		id:        timecontrol.TIME_CONTROL_3_2,
		category:  CategoryBlitz,
		initial:   3 * time.Minute,
		increment: 2 * time.Second,
	},
	timecontrol.TIME_CONTROL_5_0: {
		id:        timecontrol.TIME_CONTROL_5_0,
		category:  CategoryBlitz,
		initial:   5 * time.Minute,
		increment: 0,
	},

	// Rapid
	timecontrol.TIME_CONTROL_10_0: {
		id:        timecontrol.TIME_CONTROL_10_0,
		category:  CategoryRapid,
		initial:   10 * time.Minute,
		increment: 0,
	},
	timecontrol.TIME_CONTROL_15_10: {
		id:        timecontrol.TIME_CONTROL_15_10,
		category:  CategoryRapid,
		initial:   15 * time.Minute,
		increment: 10 * time.Second,
	},

	// Classical
	timecontrol.TIME_CONTROL_30_0: {
		id:        timecontrol.TIME_CONTROL_30_0,
		category:  CategoryClassical,
		initial:   30 * time.Minute,
		increment: 0,
	},
}

func NewTimeControl(timeControlID TimeControlID) TimeControl {
	return presets[timeControlID]
}

func (t TimeControl) ID() TimeControlID      { return t.id }
func (t TimeControl) Category() TimeCategory { return t.category }
func (t TimeControl) Duration() (time.Duration, time.Duration) {
	return t.initial, t.increment
}
