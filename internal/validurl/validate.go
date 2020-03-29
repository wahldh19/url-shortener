package validurl

import (
	"fmt"
	"net/url"
)

// Validate returns the validity of a string URL for future redirection.
func Validate(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	if u.Host == "" {
		return fmt.Errorf("missing host")
	}
	if u.Scheme == "" {
		return fmt.Errorf("missing scheme")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid scheme")
	}

	return nil
}
