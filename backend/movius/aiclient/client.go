package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	apiBaseURL = "https://api.anthropic.com/v1/messages"

	// apiVersion is the required anthropic-version request header value.
	apiVersion = "2023-06-01"

	// EnvAPIKey is the environment variable holding the Anthropic API key.
	EnvAPIKey = "ANTHROPIC_API_KEY"

	// EnvModel optionally overrides DefaultModel.
	EnvModel = "ANTHROPIC_MODEL"

	// DefaultModel is the Claude model used unless ANTHROPIC_MODEL is set.
	DefaultModel = "claude-sonnet-5"

	// MaxGuesses caps how many title guesses we ask for and accept.
	MaxGuesses = 5

	maxTokens = 1024
)

// systemPrompt demands strict JSON so the response can be parsed without
// any prose/markdown stripping heuristics (parseGuesses still strips
// defensively in case the model wraps the array anyway).
const systemPrompt = `You identify movies from vague descriptions. ` +
	`Respond with ONLY a JSON array of at most 5 objects, each shaped as ` +
	`{"title": string, "year": number}, ordered from most to least likely. ` +
	`Use the original release title and year. No prose, no markdown, no code fences.`

// Client is a small Anthropic Messages API client (raw net/http, no SDK).
type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a client for the given API key. An empty apiKey puts the
// client in MOCK mode (see mock.go) - canned guesses instead of live Claude
// calls, so the feature works without provisioning a key.
func NewClient(apiKey string) *Client {
	model := os.Getenv(EnvModel)
	if model == "" {
		model = DefaultModel
	}
	return &Client{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    apiBaseURL,
	}
}

// NewClientFromEnv reads ANTHROPIC_API_KEY from the environment.
func NewClientFromEnv() *Client {
	return NewClient(os.Getenv(EnvAPIKey))
}

// IsMock reports whether this client has no real API key configured and is
// therefore serving MOCK data.
func (c *Client) IsMock() bool {
	return strings.TrimSpace(c.apiKey) == ""
}

// --- Messages API request/response wire types --------------------------

type messagesRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// GuessMovies asks Claude for up to MaxGuesses movie title+year guesses
// matching the given vague description.
func (c *Client) GuessMovies(ctx context.Context, description string) ([]MovieGuess, error) {
	if c.IsMock() {
		return mockGuessMovies(description), nil
	}
	body, err := json.Marshal(messagesRequest{
		Model:     c.model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  []message{{Role: "user", Content: description}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal anthropic request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic messages request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic messages: unexpected status code: %d", resp.StatusCode)
	}
	var response messagesResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode anthropic response: %w", err)
	}
	var text strings.Builder
	for _, block := range response.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	return parseGuesses(text.String())
}

// parseGuesses defensively extracts a JSON array of guesses from the model
// output: strips code fences / surrounding prose, drops entries without a
// title and caps the result at MaxGuesses.
func parseGuesses(text string) ([]MovieGuess, error) {
	start := strings.Index(text, "[")
	end := strings.LastIndex(text, "]")
	if start == -1 || end == -1 || end < start {
		return nil, fmt.Errorf("anthropic response contains no JSON array: %q", text)
	}
	var guesses []MovieGuess
	if err := json.Unmarshal([]byte(text[start:end+1]), &guesses); err != nil {
		return nil, fmt.Errorf("failed to parse guesses JSON: %w", err)
	}
	out := make([]MovieGuess, 0, len(guesses))
	for _, guess := range guesses {
		if strings.TrimSpace(guess.Title) == "" {
			continue
		}
		out = append(out, guess)
		if len(out) >= MaxGuesses {
			break
		}
	}
	return out, nil
}
