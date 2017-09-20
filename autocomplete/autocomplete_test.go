package autocomplete_test

import (
	"github.com/wx13/sith/autocomplete"
	"testing"
	"time"
)

func TestAutoComplete(t *testing.T) {

	text := "elephant foo.bar::telephone(sports_car%abc->experiment)"
	ac := autocomplete.NewCompleter(text, 5)
	time.Sleep(time.Millisecond * 100)

	prefix, results := ac.Complete("soup")
	if len(results) != 0 {
		t.Error("Expected '', but got", prefix)
	}

	prefix, results = ac.Complete("eleph")
	if len(results) != 1 || results[0] != "elephant" {
		t.Error("Expected elephant, but got", results)
	}

	prefix, results = ac.Complete("sport")
	if len(results) != 1 || results[0] != "sports_car" {
		t.Error("Expected sports_car, but got", results)
	}

	prefix, results = ac.Complete("e")
	if len(results) != 2 {
		t.Error("Expected two matches, but got", len(results))
	}

}
