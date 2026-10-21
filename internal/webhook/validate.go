package webhook

import (
	"fmt"
	"regexp"
)

func validateEventID(id string) error {
	re := regexp.MustCompile("^evt_[A-Za-z0-9]{24}$")
	if !re.MatchString(id) {
		return fmt.Errorf("invalid event id %q", id)
	}
	return nil
}
