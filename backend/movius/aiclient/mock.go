package aiclient

import "strings"

// mockGuesses are canned guesses used when ANTHROPIC_API_KEY is not
// configured (Client.IsMock() == true). They intentionally match the MOCK
// movie set in movius/tmdbclient, so the identify flow resolves to real
// candidates end-to-end without any keys provisioned.
var mockGuesses = []MovieGuess{
	{Title: "Inception", Year: 2010},
	{Title: "The Matrix", Year: 1999},
	{Title: "Interstellar", Year: 2014},
}

func mockGuessMovies(description string) []MovieGuess {
	q := strings.ToLower(strings.TrimSpace(description))
	var out []MovieGuess
	for _, guess := range mockGuesses {
		if q != "" && strings.Contains(strings.ToLower(guess.Title), q) {
			out = append(out, guess)
		}
	}
	if len(out) == 0 { // vague description - return all canned guesses
		out = mockGuesses
	}
	return out
}
