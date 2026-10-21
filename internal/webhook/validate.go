//go:build !fixed

package webhook

import (
	"fmt"
	"regexp"
)

func validateEventID(id string) error {
	// ISSUE: 	The regex is compiled on every handler call, increasing allocations.
	// DEMO: 	Profiling Skill
	// FIX: 	Move eventIDPattern to package level so it's compiled once. See validate_fixed.go.
	var eventIDPattern = regexp.MustCompile("^evt_[A-Za-z0-9]{24}$")

	if !eventIDPattern.MatchString(id) {
		return fmt.Errorf("invalid event id %q", id)
	}

	return nil
}
