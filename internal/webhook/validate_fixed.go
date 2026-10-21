//go:build fixed

package webhook

import (
	"fmt"
	"regexp"
)

var eventIDPattern = regexp.MustCompile("^evt_[A-Za-z0-9]{24}$")

func validateEventID(id string) error {
	if !eventIDPattern.MatchString(id) {
		return fmt.Errorf("invalid event id %q", id)
	}

	return nil
}
