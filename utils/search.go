package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

func GetFavicon(baseURL string) string {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s://%s/favicon.ico", parsedURL.Scheme, parsedURL.Host)
}

func ExtractRedirectURL(html string) (string, error) {
	// Define a regular expression to match the JavaScript line containing the redirect URL
	re := regexp.MustCompile(`var\s+u\s*=\s*"([^"]+)"`)

	// Find the match in the HTML
	matches := re.FindStringSubmatch(html)
	if len(matches) < 2 {
		return "", fmt.Errorf("redirect URL not found in the HTML")
	}

	// Return the extracted URL
	return matches[1], nil
}

func GetRedirectedURL(rawURL string) (string, error) {
	// Parse the raw URL to ensure it's valid
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	// Create a custom HTTP client with a timeout and redirect policy
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// This function is called for each redirect.
			// You can customize the behavior here if needed.
			return nil
		},
	}

	// Make the HTTP GET request
	resp, err := client.Get(parsedURL.String())

	// Check for errors
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %v", err)
	}

	// Close the response body when the function returns
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	bodyString := string(bodyBytes)

	// Extract the redirect URL from the response body
	redirectURL, err := ExtractRedirectURL(bodyString)

	// Check for errors
	if err != nil {
		return "", fmt.Errorf("failed to extract redirect URL: %v", err)
	}

	// Return the extracted URL

	return redirectURL, nil

}
