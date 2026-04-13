package file

import (
	"container/list"
	"sync"
	"time"

	"github.com/wx13/sith/file/buffer"
	"github.com/wx13/sith/file/cursor"
)

// BufferState contains a snapshot of the buffer state,
// including multicursor positions.
type BufferState struct {
	buff      buffer.Buffer
	mc        cursor.MultiCursor
	saved     bool
	timestamp time.Time
}

// NewBufferState creates a new buffer state snapshot from the
// current state.
func NewBufferState(buff buffer.Buffer, mc cursor.MultiCursor) *BufferState {
	return &BufferState{
		buff:      buff,
		mc:        mc,
		timestamp: time.Now(),
	}
}

// BufferHist manages a history of buffer states.
type BufferHist struct {
	list      *list.List
	element   *list.Element
	elemMutex *sync.Mutex

	snapChan chan struct{}
	snapReq  BufferState
	reqMutex *sync.Mutex
}

// NewBufferHist creates a new BufferHist object initialized with the current state.
func NewBufferHist(buffer buffer.Buffer, cursor cursor.MultiCursor) *BufferHist {
	bh := BufferHist{}
	state := NewBufferState(buffer, cursor)
	bh.list = list.New()
	bh.element = bh.list.PushBack(state)
	bh.elemMutex = &sync.Mutex{}
	bh.snapChan = make(chan struct{}, 1)
	bh.reqMutex = &sync.Mutex{}
	bh.handleSnapshots()
	bh.ForceSnapshot(buffer, cursor)
	bh.SnapshotSaved()
	return &bh
}

// ForceSnapshot forces a snapshot rather than requesting one.
func (bh *BufferHist) ForceSnapshot(buff buffer.Buffer, mc cursor.MultiCursor) {
	bh.snapshot(buff.Dup(), mc.Dup())
}

//  Snapshot places a snapshot request onto the snapshot queue.
func (bh *BufferHist) Snapshot(buff buffer.Buffer, mc cursor.MultiCursor) {
	request := BufferState{
		buff: buff.Dup(),
		mc:   mc.Dup(),
	}

	bh.reqMutex.Lock()
	bh.snapReq = request
	bh.reqMutex.Unlock()

	select {
	case bh.snapChan <- struct{}{}:
	default:
	}
}

func (bh *BufferHist) handleSnapshots() {
	go func() {
		for range time.Tick(time.Millisecond * 100) {
			select {
			case <-bh.snapChan:
				bh.reqMutex.Lock()
				bh.snapshot(bh.snapReq.buff, bh.snapReq.mc)
				bh.reqMutex.Unlock()
			}
		}
	}()
}

// SnapshotSaved toggles on the "saved" attribute for the current state.
func (bh *BufferHist) SnapshotSaved() {
	bh.elemMutex.Lock()
	bh.element.Value.(*BufferState).saved = true
	bh.elemMutex.Unlock()
}

func (bh *BufferHist) snapshot(buff buffer.Buffer, mc cursor.MultiCursor) {

	curBuf, curMC := bh.Current()

	curRow := curMC.GetRow(0)
	newRow := mc.GetRow(0)
	if curRow-newRow > 5 || newRow-curRow > 5 {
		state := NewBufferState(curBuf, mc)
		bh.elemMutex.Lock()
		bh.element = bh.list.InsertAfter(state, bh.element)
		bh.elemMutex.Unlock()
	}

	state := NewBufferState(buff, mc)
	bh.elemMutex.Lock()
	bh.element = bh.list.InsertAfter(state, bh.element)
	bh.elemMutex.Unlock()

	bh.trim()

}

func (bh *BufferHist) trim() {

	if bh.list.Len() < 200 {
		return
	}

	rm := []*list.Element{}

	n := 0
	ns := 0
	for el := bh.element.Next(); el != nil; el = el.Next() {
		n++
		state := el.Value.(*BufferState)
		if state.saved {
			ns++
			if ns > 50 {
				rm = append(rm, el)
			}
			continue
		}
		if n > 50 {
			rm = append(rm, el)
		}
	}

	n = 0
	ns = 0
	for el := bh.element.Prev(); el != nil; el = el.Prev() {
		n++
		state := el.Value.(*BufferState)
		if state.saved {
			ns++
			if ns > 50 {
				rm = append(rm, el)
			}
			continue
		}
		if n > 50 {
			rm = append(rm, el)
		}
	}

	for _, el := range rm {
		bh.list.Remove(el)
	}

}

// Current returns the current buffer snapshot..
func (bh *BufferHist) Current() (buffer.Buffer, cursor.MultiCursor) {
	state := bh.element.Value.(*BufferState)
	return state.buff, state.mc.Dup()
}

// Next bumps the current pointer to the next state (redo).
func (bh *BufferHist) Next() (buffer.Buffer, cursor.MultiCursor) {
	next := bh.element.Next()
	if next != nil {
		bh.element = next
	}
	return bh.Current()
}

// Prev bumps the current pointer to the previous state (undo).
func (bh *BufferHist) Prev() (buffer.Buffer, cursor.MultiCursor) {
	bh.elemMutex.Lock()
	prev := bh.element.Prev()
	bh.elemMutex.Unlock()
	if prev != nil {
		bh.elemMutex.Lock()
		bh.element = prev
		bh.elemMutex.Unlock()
	}
	return bh.Current()
}

// NextSaved bumps the current pointer to the next saved state (macro redo).
func (bh *BufferHist) NextSaved() (buffer.Buffer, cursor.MultiCursor) {
	for el := bh.element.Next(); el != nil; el = el.Next() {
		if el.Value.(*BufferState).saved {
			bh.SnapshotSaved()
			bh.elemMutex.Lock()
			bh.element = el
			bh.elemMutex.Unlock()
			break
		}
	}
	return bh.Current()
}

// PrevSaved bumps the current pointer to the previous saved state (macro undo).
func (bh *BufferHist) PrevSaved() (buffer.Buffer, cursor.MultiCursor) {
	for el := bh.element.Prev(); el != nil; el = el.Prev() {
		if el.Value.(*BufferState).saved {
			bh.SnapshotSaved()
			bh.elemMutex.Lock()
			bh.element = el
			bh.elemMutex.Unlock()
			break
		}
	}
	return bh.Current()
}

// StateInfo contains metadata about a buffer state for display purposes.
type StateInfo struct {
	Timestamp time.Time
	LineDelta int // relative to current state
	IsCurrent bool
	IsSaved   bool
	Index     int // index in the returned slice, for jumping to this state
	element   *list.Element
}

// GetSavedStates returns metadata about all saved states plus the current state.
// States are returned in chronological order (oldest first).
func (bh *BufferHist) GetSavedStates() []StateInfo {
	bh.elemMutex.Lock()
	defer bh.elemMutex.Unlock()

	currentLines := bh.element.Value.(*BufferState).buff.Length()
	currentIsSaved := bh.element.Value.(*BufferState).saved

	states := []StateInfo{}

	// Collect all saved states from the beginning
	for el := bh.list.Front(); el != nil; el = el.Next() {
		state := el.Value.(*BufferState)
		isCurrent := el == bh.element

		// Include if saved OR if it's the current position
		if state.saved || isCurrent {
			info := StateInfo{
				Timestamp: state.timestamp,
				LineDelta: state.buff.Length() - currentLines,
				IsCurrent: isCurrent,
				IsSaved:   state.saved,
				Index:     len(states),
				element:   el,
			}
			states = append(states, info)
		}
	}

	// If current is not saved, we already included it above
	// Mark it appropriately
	if !currentIsSaved {
		for i := range states {
			if states[i].IsCurrent {
				states[i].IsSaved = false
				break
			}
		}
	}

	return states
}

// JumpToState moves to the state at the given element.
func (bh *BufferHist) JumpToState(info StateInfo) (buffer.Buffer, cursor.MultiCursor) {
	bh.elemMutex.Lock()
	bh.element = info.element
	bh.elemMutex.Unlock()
	return bh.Current()
}

// GetStateDiff returns the lines that differ between the current state and the target state.
// Returns two slices: lines only in current (removals) and lines only in target (additions).
func (bh *BufferHist) GetStateDiff(info StateInfo) (removals, additions []string) {
	bh.elemMutex.Lock()
	currentBuff := bh.element.Value.(*BufferState).buff
	targetBuff := info.element.Value.(*BufferState).buff
	bh.elemMutex.Unlock()

	currentLines := make(map[string]int)
	for _, line := range currentBuff.Lines() {
		currentLines[line.ToString()]++
	}

	targetLines := make(map[string]int)
	for _, line := range targetBuff.Lines() {
		targetLines[line.ToString()]++
	}

	// Lines in current but not in target (would be removed)
	for _, line := range currentBuff.Lines() {
		str := line.ToString()
		if targetLines[str] < currentLines[str] {
			removals = append(removals, str)
			targetLines[str]++ // avoid counting same line twice
		}
	}

	// Reset target counts
	targetLines = make(map[string]int)
	for _, line := range targetBuff.Lines() {
		targetLines[line.ToString()]++
	}

	// Lines in target but not in current (would be added)
	for _, line := range targetBuff.Lines() {
		str := line.ToString()
		if currentLines[str] < targetLines[str] {
			additions = append(additions, str)
			currentLines[str]++ // avoid counting same line twice
		}
	}

	return removals, additions
}
