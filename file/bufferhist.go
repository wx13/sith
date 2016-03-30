package file

import "time"

type BufferHist struct {
	buffers  []Buffer
	cursors  []MultiCursor
	idx      int
	snapChan chan struct{}
	snapReq  SnapshotRequest
}

type SnapshotRequest struct {
	Buffer Buffer
	Cursor MultiCursor
}

func NewBufferHist(buffer Buffer, cursor MultiCursor) *BufferHist {
	bh := BufferHist{}
	bh.buffers = append(bh.buffers, buffer)
	bh.cursors = append(bh.cursors, cursor.Dup())
	bh.snapChan = make(chan struct{}, 1)
	bh.handleSnapshots()
	return &bh
}

func (bh *BufferHist) ForceSnapshot(buffer Buffer, mc MultiCursor) {
	bh.snapshot(buffer, mc)
}

func (bh *BufferHist) Snapshot(buffer Buffer, mc MultiCursor) {
	request := SnapshotRequest{
		Buffer: buffer,
		Cursor: mc,
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

func (bh *BufferHist) snapshot(buffer Buffer, mc MultiCursor) {

	var buffers []Buffer
	var cursors []MultiCursor

	dist := bh.cursors[bh.idx][0].row - mc[0].row
	if bh.idx < len(bh.buffers) && (dist < -1 || dist > 1) {
		bh.idx = bh.idx + 1

		buffers = append(bh.buffers[:bh.idx], bh.buffers[bh.idx-1].Dup())
		bh.buffers = append(buffers, bh.buffers[bh.idx:]...)

		cursors = append(bh.cursors[:bh.idx], mc.Dup())
		bh.cursors = append(cursors, bh.cursors[bh.idx:]...)
	}

	bh.idx = bh.idx + 1

	buffers = append(bh.buffers[:bh.idx], buffer.Dup())
	bh.buffers = append(buffers, bh.buffers[bh.idx:]...)

	cursors = append(bh.cursors[:bh.idx], mc.Dup())
	bh.cursors = append(cursors, bh.cursors[bh.idx:]...)

	bh.Trim(200)

}

func (bh *BufferHist) Trim(n int) {
	if bh.idx+n < len(bh.buffers) {
		bh.buffers = bh.buffers[:(bh.idx + n)]
		bh.cursors = bh.cursors[:(bh.idx + n)]
	}
	if bh.idx >= n {
		bh.buffers = bh.buffers[(bh.idx - n):]
		bh.cursors = bh.cursors[(bh.idx - n):]
		bh.idx -= bh.idx - n
	}
}

func (bh *BufferHist) Current() (Buffer, MultiCursor) {
	return bh.buffers[bh.idx], bh.cursors[bh.idx].Dup()
}

func (bh *BufferHist) Next() (Buffer, MultiCursor) {
	return bh.Increment(1)
}

func (bh *BufferHist) Prev() (Buffer, MultiCursor) {
	return bh.Increment(-1)
}

func (bh *BufferHist) Increment(n int) (Buffer, MultiCursor) {
	bh.idx += n
	if bh.idx >= len(bh.buffers) {
		bh.idx = len(bh.buffers) - 1
	}
	if bh.idx < 0 {
		bh.idx = 0
	}
	return bh.Current()
}
