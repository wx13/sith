package file

import (
	"container/list"
	"sync"
	"time"

	"github.com/wx13/sith/file/buffer"
	"github.com/wx13/sith/file/cursor"
)

type BufferState struct {
	buff  buffer.Buffer
	mc    cursor.MultiCursor
	saved bool
}

func NewBufferState(buff buffer.Buffer, mc cursor.MultiCursor) *BufferState {
	return &BufferState{
		buff: buff,
		mc:   mc,
	}
}

type BufferHist struct {
	list      *list.List
	element   *list.Element
	elemMutex *sync.Mutex

	snapChan chan struct{}
	snapReq  SnapshotRequest
	reqMutex *sync.Mutex
}

type SnapshotRequest struct {
	Buffer buffer.Buffer
	Cursor cursor.MultiCursor
}

func NewBufferHist(buffer buffer.Buffer, cursor cursor.MultiCursor) *BufferHist {
	bh := BufferHist{}
	state := NewBufferState(buffer, cursor)
	bh.list = list.New()
	bh.element = bh.list.PushBack(state)
	bh.elemMutex = &sync.Mutex{}
	bh.snapChan = make(chan struct{}, 1)
	bh.reqMutex = &sync.Mutex{}
	bh.handleSnapshots()
	return &bh
}

func (bh *BufferHist) ForceSnapshot(buff buffer.Buffer, mc cursor.MultiCursor) {
	bh.snapshot(buff.Dup(), mc.Dup())
}

func (bh *BufferHist) Snapshot(buff buffer.Buffer, mc cursor.MultiCursor) {
	request := SnapshotRequest{
		Buffer: buff.Dup(),
		Cursor: mc.Dup(),
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
				bh.snapshot(bh.snapReq.Buffer, bh.snapReq.Cursor)
				bh.reqMutex.Unlock()
			}
		}
	}()
}

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

func (bh *BufferHist) Current() (buffer.Buffer, cursor.MultiCursor) {
	state := bh.element.Value.(*BufferState)
	return state.buff, state.mc.Dup()
}

func (bh *BufferHist) Next() (buffer.Buffer, cursor.MultiCursor) {
	next := bh.element.Next()
	if next != nil {
		bh.element = next
	}
	return bh.Current()
}

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
