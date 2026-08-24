package timecontrol

const (
	TIME_CONTROL_1_0   = "1+0"
	TIME_CONTROL_2_1   = "2+1"
	TIME_CONTROL_3_0   = "3+0"
	TIME_CONTROL_3_2   = "3+2"
	TIME_CONTROL_5_0   = "5+0"
	TIME_CONTROL_10_0  = "10+0"
	TIME_CONTROL_15_10 = "15+10"
	TIME_CONTROL_30_0  = "30+0"
)

var allowedTimeControlIDs = map[string]struct{}{
	TIME_CONTROL_1_0:   {},
	TIME_CONTROL_2_1:   {},
	TIME_CONTROL_3_0:   {},
	TIME_CONTROL_3_2:   {},
	TIME_CONTROL_5_0:   {},
	TIME_CONTROL_10_0:  {},
	TIME_CONTROL_15_10: {},
	TIME_CONTROL_30_0:  {},
}

func IsValidTimeControlID(tcID string) bool {
	_, ok := allowedTimeControlIDs[tcID]
	return ok
}

func AllTimeControlIDs() []string {
	tcs := make([]string, 0, len(allowedTimeControlIDs))
	for tcID := range allowedTimeControlIDs {
		tcs = append(tcs, tcID)
	}
	return tcs
}
