package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://www.dictionaryapi.com/api/v3/references"

// Client holds the HTTP client and both MW API keys.
type Client struct {
	http            *http.Client
	apiKeyDict      string
	apiKeyThesaurus string
	baseURL         string // overrideable in tests
}

// NewClient creates a Client with a 10-second timeout.
func NewClient(apiKeyDict, apiKeyThesaurus string) *Client {
	return &Client{
		http:            &http.Client{Timeout: 10 * time.Second},
		apiKeyDict:      apiKeyDict,
		apiKeyThesaurus: apiKeyThesaurus,
		baseURL:         defaultBaseURL,
	}
}

// SetBaseURL overrides the API base URL. Used in tests to point at a local server.
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// get performs a GET to the MW API and decodes JSON into dest (must be a pointer).
// MW returns HTTP 200 with a []string of suggestions when a word is not found,
// so this method probes the raw JSON before decoding into the target type.
func (c *Client) get(ref, word, apiKey string, dest interface{}) error {
	url := fmt.Sprintf("%s/%s/json/%s?key=%s", c.baseURL, ref, word, apiKey)

	resp, err := c.http.Get(url)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// MW returns []string when the word is not found.
	// Some API errors (e.g. invalid key) come back as HTTP 200 with plain text.
	var raw []json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("API error: %s", strings.TrimSpace(string(body)))
	}

	if len(raw) == 0 {
		return &WordNotFoundError{Word: word}
	}

	// If the first element is a JSON string, the whole array is suggestions.
	var probe string
	if err := json.Unmarshal(raw[0], &probe); err == nil {
		suggestions := make([]string, 0, len(raw))
		for _, r := range raw {
			var s string
			if json.Unmarshal(r, &s) == nil {
				suggestions = append(suggestions, s)
			}
		}
		return &WordNotFoundError{Word: word, Suggestions: suggestions}
	}

	return json.Unmarshal(body, dest)
}
