package editor

import "testing"

func TestRepeatCmd(t *testing.T) {
	ed := &Editor{}
	count := 0
	ed.lastCmd = Action{func() { count++ }, "test-cmd"}
	ed.RepeatCmd()
	ed.RepeatCmd()
	if count != 2 {
		t.Errorf("expected command to run twice; ran %d times", count)
	}
}

func TestRepeatCmdRecursionGuard(t *testing.T) {
	// If the repeat command itself ever gets recorded, repeating must
	// not recurse.
	ed := &Editor{}
	count := 0
	ed.lastCmd = Action{func() { count++; ed.RepeatCmd() }, "repeat"}
	ed.RepeatCmd()
	if count != 1 {
		t.Errorf("expected exactly one invocation; got %d", count)
	}
}
