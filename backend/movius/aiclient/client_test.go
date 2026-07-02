package aiclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient_IsMock(t *testing.T) {
	if !NewClient("").IsMock() {
		t.Error("expected empty api key to be mock mode")
	}
	if NewClient("real-key").IsMock() {
		t.Error("expected non-empty api key to not be mock mode")
	}
	if !NewClientFromEnv().IsMock() {
		t.Error("expected NewClientFromEnv() to be mock mode when ANTHROPIC_API_KEY is unset in test env")
	}
}

func TestNewClient_ModelOverride(t *testing.T) {
	if got := NewClient("").model; got != DefaultModel {
		t.Errorf("model = %q, want default %q", got, DefaultModel)
	}
	t.Setenv(EnvModel, "claude-test-model")
	if got := NewClient("").model; got != "claude-test-model" {
		t.Errorf("model = %q, want env override %q", got, "claude-test-model")
	}
}

func TestClient_GuessMovies_Mock(t *testing.T) {
	c := NewClient("")
	guesses, err := c.GuessMovies(context.Background(), "a dream heist movie")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(guesses) == 0 {
		t.Fatal("expected canned guesses for a vague description")
	}

	guesses, err = c.GuessMovies(context.Background(), "matrix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(guesses) != 1 || guesses[0].Title != "The Matrix" {
		t.Fatalf("expected The Matrix for 'matrix', got: %+v", guesses)
	}
}

func TestParseGuesses(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    []string
		wantErr bool
	}{
		{
			name: "plain JSON array",
			text: `[{"title":"Inception","year":2010},{"title":"The Matrix","year":1999}]`,
			want: []string{"Inception", "The Matrix"},
		},
		{
			name: "wrapped in code fences and prose",
			text: "Here you go:\n```json\n[{\"title\":\"Interstellar\",\"year\":2014}]\n```\nHope that helps!",
			want: []string{"Interstellar"},
		},
		{
			name: "drops entries without a title",
			text: `[{"title":"","year":2000},{"title":"Pulp Fiction","year":1994}]`,
			want: []string{"Pulp Fiction"},
		},
		{
			name: "caps at MaxGuesses",
			text: `[{"title":"a"},{"title":"b"},{"title":"c"},{"title":"d"},{"title":"e"},{"title":"f"}]`,
			want: []string{"a", "b", "c", "d", "e"},
		},
		{
			name:    "no JSON array",
			text:    "Sorry, I can't help with that.",
			wantErr: true,
		},
		{
			name:    "malformed JSON",
			text:    `[{"title": "Broken"`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGuesses(tt.text)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got: %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d guesses, want %d: %+v", len(got), len(tt.want), got)
			}
			for i, title := range tt.want {
				if got[i].Title != title {
					t.Errorf("guess[%d].Title = %q, want %q", i, got[i].Title, title)
				}
			}
		})
	}
}

// TestClient_GuessMovies_HTTP exercises the live-request path against a local
// httptest server (no external network) - verifying headers, request body and
// response parsing.
func TestClient_GuessMovies_HTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-api-key"); got != "test-key" {
			t.Errorf("x-api-key = %q, want %q", got, "test-key")
		}
		if got := r.Header.Get("anthropic-version"); got != apiVersion {
			t.Errorf("anthropic-version = %q, want %q", got, apiVersion)
		}
		var req messagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if req.Model != DefaultModel {
			t.Errorf("model = %q, want %q", req.Model, DefaultModel)
		}
		if len(req.Messages) != 1 || req.Messages[0].Content != "submarine movie" {
			t.Errorf("unexpected messages: %+v", req.Messages)
		}
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"[{\"title\":\"Titanic\",\"year\":1997}]"}]}`))
	}))
	defer server.Close()

	c := NewClient("test-key")
	c.baseURL = server.URL
	guesses, err := c.GuessMovies(context.Background(), "submarine movie")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(guesses) != 1 || guesses[0].Title != "Titanic" || guesses[0].Year != 1997 {
		t.Fatalf("unexpected guesses: %+v", guesses)
	}
}

func TestClient_GuessMovies_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	c := NewClient("test-key")
	c.baseURL = server.URL
	if _, err := c.GuessMovies(context.Background(), "anything"); err == nil {
		t.Fatal("expected error for non-200 response")
	}
}
