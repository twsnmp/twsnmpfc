package notify

import (
	"fmt"
	"net/url"
)

// validateURL parses the raw URL and checks if it uses http or https scheme and has a non-empty host.
func validateURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("only http and https schemes are allowed")
	}
	if u.Host == "" {
		return "", fmt.Errorf("url host is empty")
	}
	return u.String(), nil
}
