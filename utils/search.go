package utils

import (
	"fmt"
	"net/url"
)

func GetFavicon(baseURL string) string {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s://%s/favicon.ico", parsedURL.Scheme, parsedURL.Host)
}
