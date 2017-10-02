// Package autocomplete provides basic autocompletion functionality.
// It takes in a set of files and creates a completer based on all the
// file contents.
package autocomplete

import (
	"regexp"
	"sync"
)

// Completer provides autocompletion suggestions.
type Completer struct {
	words       []string    // Master list of possible suggestions
	mutex       *sync.Mutex // Protect the master word list.
	requestChan chan string // So we only run one update at a time.
	minLen      int         // Words shorter than this are "insignificant".
}

// NewCompleter creates a new completer object.
func NewCompleter(text string, minLen int) *Completer {
	cmplt := Completer{}
	cmplt.mutex = &sync.Mutex{}
	cmplt.minLen = minLen

	// Set up the updater to wait for requests.
	cmplt.requestChan = make(chan string, 1)
	go cmplt.keepUpdating()

	// Request an update.
	cmplt.Update(text)

	return &cmplt
}

// Update requests an update.
func (cmplt *Completer) Update(text string) {
	select {
	case cmplt.requestChan <- text:
	default:
	}
}

// Split defines the word boundaries for autocompletion.
func Split(str string) []string {
	re := regexp.MustCompile("[^a-zA-Z0-9_]+")
	return re.Split(str, -1)
}

func (cmplt *Completer) Split(str string) []string {
	return Split(str)
}

// keepUpdating processes update requests.
func (cmplt *Completer) keepUpdating() {
	for {
		select {
		case text := <-cmplt.requestChan:
			// Construct the words list.
			words_list := Split(text)
			word_frequencies := map[string]int{}
			for _, word := range words_list {
				if len(word) > cmplt.minLen {
					word_frequencies[word]++
				}
			}
			words := []string{}
			for word, _ := range word_frequencies {
				words = append(words, word)
			}

			// Carefully swap out the words list.
			cmplt.mutex.Lock()
			cmplt.words = words
			cmplt.mutex.Unlock()
		}
	}
}

// getCommonPrefix looks for the longest common prefix of a set of strings.
// The parameter 'prefix' provides a starting point, so we don't have to
// search as hard.
func getCommonPrefix(prefix string, matches []string) string {
	if len(matches) == 0 {
		return prefix
	}
loop:
	for i := len(prefix); i < len(matches[0]); i++ {
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

// Complete returns completion results for a prefix string.
func (cmplt *Completer) Complete(prefix string) (string, []string) {
	words := Split(prefix)
	prefix = words[len(words)-1]
	matches := []string{}
	n := len(prefix)
	for _, word := range cmplt.words {
		if len(word) <= n {
			continue
		}
		if word[:n] == prefix {
			matches = append(matches, word)
		}
	}
	// Check if matches share additional prefix.
	commonPrefix := getCommonPrefix(prefix, matches)
	return commonPrefix, matches
}
