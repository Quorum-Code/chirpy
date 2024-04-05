package internal

import (
	"slices"
	"strings"
)

func StripProfane(text string) string {
	profane := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	words := strings.Split(text, " ")
	for i := range words {
		if slices.Contains(profane, strings.ToLower(words[i])) {
			words[i] = "****"
		}
	}

	cleaned_text := strings.Join(words, " ")

	return cleaned_text
}
