package file

import (
	"container/list"
	"time"
)

type BufferState struct {
	buff  Buffer
	mc    MultiCursor
	saved bool
}

func NewBufferState(buff Buffer, mc MultiCursor) *BufferState {
	return &BufferState{
		buff: buff,
		mc:   mc,
	}
}

type BufferHist struct {
	list     *list.List
	element  *list.Element
	snapChan chan struct{}
	snapReq  SnapshotRequest
}

type SnapshotRequest struct {
	Buffer Buffer
	Cursor MultiCursor
}

func NewBufferHist(buffer Buffer, cursor MultiCursor) *BufferHist {
	bh := BufferHist{}
	state := NewBufferState(buffer, cursor)
	bh.list = list.New()
	bh.element = bh.list.PushBack(state)
	bh.snapChan = make(chan struct{}, 1)
	bh.handleSnapshots()
	return &bh
}

func (bh *BufferHist) ForceSnapshot(buffer Buffer, mc MultiCursor) {
	bh.snapshot(buffer, mc)
}

func (bh *BufferHist) Snapshot(buffer Buffer, mc MultiCursor) {
	request := SnapshotRequest{
		Buffer: buffer.Dup(),
		Cursor: mc.Dup(),
	}
	bh.snapReq = request
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
				bh.snapshot(bh.snapReq.Buffer, bh.snapReq.Cursor)
			}
		}
	}()
}

func (bh *BufferHist) SnapshotSaved() {
	bh.element.Value.(*BufferState).saved = true
}

func (bh *BufferHist) snapshot(buffer Buffer, mc MultiCursor) {

	state := NewBufferState(buffer, mc)
	bh.element = bh.list.InsertAfter(state, bh.element)
	bh.Trim()
}

func (bh *BufferHist) Trim() {

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

func (bh *BufferHist) Current() (Buffer, MultiCursor) {
	state := bh.element.Value.(*BufferState)
	return state.buff, state.mc.Dup()
}

func (bh *BufferHist) Next() (Buffer, MultiCursor) {
	next := bh.element.Next()
	if next != nil {
		bh.element = next
	}
	return bh.Current()
}

func (bh *BufferHist) Prev() (Buffer, MultiCursor) {
	prev := bh.element.Prev()
	if prev != nil {
		bh.element = prev
	}
	return bh.Current()
}

func (bh *BufferHist) NextSaved() (Buffer, MultiCursor) {
	for el := bh.element.Next(); el != nil; el = el.Next() {
		if el.Value.(*BufferState).saved {
			bh.SnapshotSaved()
			bh.element = el
			break
		}
	}
	return bh.Current()
}

func (bh *BufferHist) PrevSaved() (Buffer, MultiCursor) {
	for el := bh.element.Prev(); el != nil; el = el.Prev() {
		if el.Value.(*BufferState).saved {
			bh.SnapshotSaved()
			bh.element = el
			break
		}
	}
	return bh.Current()
}
