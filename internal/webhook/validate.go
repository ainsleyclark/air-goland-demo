package webhook

import (
	"fmt"
	"regexp"
)

func validateEventID(id string) error {
	// ISSUE: 	The regex is compiled on every handler call, increasing allocations.
	// FIX: 	Benchmark Skill should move this line above validateEventID, to package level, so it's compiled once:
	//
	//	var eventIDPattern = regexp.MustCompile("^evt_[A-Za-z0-9]{24}$")
	var eventIDPattern = regexp.MustCompile("^evt_[A-Za-z0-9]{24}$")

	if !eventIDPattern.MatchString(id) {
		return fmt.Errorf("invalid event id %q", id)
	}

	return nil
}
