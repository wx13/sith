package autocomplete_test

import (
	"github.com/wx13/sith/autocomplete"
	"testing"
)

func stringSliceEq(a []string, b ...string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTokenMatch(t *testing.T) {

	tokens := []string{}
	matches := autocomplete.TokenMatch(tokens, "", 3)
	if len(matches) > 0 {
		t.Error("Empty token list:", matches)
	}

	tokens = []string{"tree", "apple tree", "freedom"}
	matches = autocomplete.TokenMatch(tokens, "tree", 3)
	if !stringSliceEq(matches, tokens...) {
		t.Error("Failed match:", tokens, matches)
	}
	matches = autocomplete.TokenMatch(tokens, "tree", 4)
	if !stringSliceEq(matches, "tree", "apple tree") {
		t.Error("Failed match:", tokens, matches)
	}

}

func TestSuffixes(t *testing.T) {

	tokens := []string{}
	suffixes := autocomplete.Suffixes(tokens, "", 3)
	if len(suffixes) > 0 {
		t.Error("Empty token list:", suffixes)
	}

	tokens = []string{"football", "baseball", "baller"}
	suffixes = autocomplete.Suffixes(tokens, "ball", 3)
	if !stringSliceEq(suffixes, "er") {
		t.Error(tokens, suffixes)
	}

	tokens = []string{"football", "baseball", "baller"}
	suffixes = autocomplete.Suffixes(tokens, "ball", 4)
	if !stringSliceEq(suffixes, "er") {
		t.Error(tokens, suffixes)
	}

	tokens = []string{"football", "baseball", "baller"}
	suffixes = autocomplete.Suffixes(tokens, "ball", 4)
	if !stringSliceEq(suffixes, "er") {
		t.Error(tokens, suffixes)
	}

	tokens = []string{"football player", "baseball player", "baller"}
	suffixes = autocomplete.Suffixes(tokens, "ball", 4)
	if !stringSliceEq(suffixes, "er") {
		t.Errorf("%#v %#v", tokens, suffixes)
	}

}

func TestComplete(t *testing.T) {
	ac := autocomplete.New()

	// Basic example.
	ans := ac.Complete("ball", "football baseball baller")
	if !stringSliceEq(ans, "er") {
		t.Error(ans)
	}

	// With punctuation.
	ans = ac.Complete("ball::", "football::player", "baseball::player")
	if !stringSliceEq(ans, "player") {
		t.Error(ans)
	}

	// More punctuation stuff.
	ans = ac.Complete("foot", "football::player", "baseball::player")
	if !stringSliceEq(ans, "ball") {
		t.Error(ans)
	}

}
