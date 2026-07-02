// Package aiclient is a small Claude (Anthropic Messages API) client used to
// turn a vague natural-language movie description ("that submarine movie with
// the guy from Titanic") into a short list of likely movie title guesses. The
// guesses are then grounded via TMDB search by facade4listus.IdentifyMovies.
//
// If ANTHROPIC_API_KEY is not configured the client transparently falls back
// to MOCK data (see mock.go) so the AI-identify feature works end-to-end
// without provisioning a key.
package aiclient

// MovieGuess is a single movie title guess proposed by the AI.
type MovieGuess struct {
	Title string `json:"title"`
	Year  int    `json:"year,omitempty"`
}
