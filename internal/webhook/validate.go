package webhook

import (
	"fmt"
	"regexp"
)

func validateEventID(id string) error {
	var eventIDPattern = regexp.MustCompile("^evt_[A-Za-z0-9]{24}$")

	if !eventIDPattern.MatchString(id) {
		return fmt.Errorf("invalid event id %q", id)
	}

	return nil
}
