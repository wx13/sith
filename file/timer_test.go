package file_test

import (
	"github.com/wx13/sith/file"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {

	timer := file.MakeTimer()
	time.Sleep(1 * time.Millisecond)
	timer.Tick()
	time.Sleep(1 * time.Millisecond)
	rate := timer.Tick()
	if rate < 800 || rate > 1300 {
		t.Error("Tick gives wrong rate:", rate)
	}

}
