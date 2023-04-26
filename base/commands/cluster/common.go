package cluster

import (
	"fmt"
	"strings"
)

var validStates = []string{"ACTIVE", "NO_MIGRATION", "FROZEN", "PASSIVE", "IN_TRANSITION"}

func stateToString(state byte) string {
	if int(state) < len(validStates) {
		return validStates[state]
	}
	return "UNKNOWN"
}

func stringToState(s string) (int32, error) {
	s = strings.ToUpper(s)
	for i, v := range validStates {
		if v == s {
			return int32(i), nil
		}
	}
	return 255, fmt.Errorf("unknown state: %s", s)
}
