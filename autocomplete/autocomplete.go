// Package autocomplete provides basic autocompletion functionality.
// It takes in a set of files and creates a completer based on all the
// file contents.
package autocomplete

import (
	"regexp"
	"strings"
)

type AutoComplete struct {
	minLen int
}

func New() *AutoComplete {
	return &AutoComplete{minLen: 3}
}

func (ac *AutoComplete) Complete(prefix string, corpora ...string) []string {

	f := strings.Fields(prefix)
	prefix = f[len(f)-1]

	// If prefix is too short, just return.
	if len(prefix) < ac.minLen {
		return []string{}
	}

	// Join corpora into one corpus, split at whitespace, and limit in size.
	corpus := strings.Join(corpora, "\n")
	tokens_map := map[string]bool{}
	for _, token := range strings.Fields(corpus) {
		if len(tokens_map) > 10000 {
			break
		}
		if len(token) <= ac.minLen {
			continue
		}
		tokens_map[token] = true
	}
	if len(tokens_map) == 0 {
		return []string{}
	}
	tokens := []string{}
	for token, _ := range tokens_map {
		// Don't match yourself:
		if token != prefix {
			tokens = append(tokens, token)
		}
	}

	matches := TokenMatch(tokens, prefix, ac.minLen)
	if len(matches) <= 1 {
		return Suffixes(matches, prefix, ac.minLen)
	}

	for n := 1; n < 100; n++ {
		tokens = matches
		matches = TokenMatch(tokens, prefix, ac.minLen+n)
		if len(matches) == 0 {
			return Suffixes(tokens, prefix, ac.minLen+n-1)
		}
		if len(matches) == 1 {
			return Suffixes(matches, prefix, ac.minLen+n)
		}
	}

	return matches
}

func TokenMatch(tokens []string, prefix string, n int) []string {
	if len(prefix) < n {
		return []string{}
	}

	// Create a regex to search the text.
	tail := prefix[len(prefix)-n:]
	re, err := regexp.Compile(regexp.QuoteMeta(tail))
	if err != nil {
		return []string{}
	}

	// Find tail matches.
	matches := []string{}
	for _, token := range tokens {
		if re.MatchString(token) {
			matches = append(matches, token)
		}
	}

	return matches
}

// Suffixes returns the string suffix after the prefix.
func Suffixes(tokens []string, prefix string, n int) []string {
	if len(prefix) < n {
		return []string{}
	}

	// Create a regex to search the text.
	tail := prefix[len(prefix)-n:]
	re, err := regexp.Compile(regexp.QuoteMeta(tail))
	if err != nil {
		return []string{}
	}

	startsWithPunctRe := regexp.MustCompile("^[^a-zA-Z0-9]")
	punctRe := regexp.MustCompile("[^a-zA-Z0-9]")
	charRe := regexp.MustCompile("[a-zA-Z0-9]")

	// Store in a map to dedup.
	suffixes_map := map[string]bool{}
	for _, token := range tokens {
		idx := re.FindStringIndex(token)[1]
		if idx == len(token) {
			continue
		}
		// Start at end of prefix.
		suffix := token[idx:]
		// If first char is not punctuation, go until first punctuation.
		if startsWithPunctRe.MatchString(suffix) {
			suffix = charRe.Split(suffix, 2)[0]
		} else {
			suffix = punctRe.Split(suffix, 2)[0]
		}
		suffixes_map[suffix] = true
	}

	suffixes := []string{}
	for suffix, _ := range suffixes_map {
		if len(suffix) == 0 || suffix[0] == ' ' || suffix[0] == '\t' {
			continue
		}
		suffixes = append(suffixes, strings.Fields(suffix)[0])
	}

	return suffixes
}

// GetCommonPrefix looks for the longest common prefix of a set of strings.
func GetCommonPrefix(matches []string) string {
	if len(matches) == 0 {
		return ""
	}
	prefix := ""
loop:
	for i := 0; i < len(matches[0]); i++ {
		c := matches[0][i]
		for _, match := range matches {
			// If even one word is too short, then we are done.
			if i >= len(match) {
				break loop
			}
			// If even one character is mismatched, then we are done.
			if match[i] != c {
				break loop
			}
		}
		// We didn't break, so append the char.
		prefix += string(c)
	}
	return prefix
}

func (ac *AutoComplete) GetCommonPrefix(matches []string) string {
	return GetCommonPrefix(matches)
}
