package leakybuffer

import (
	"bytes"
)

// LeakyBuffer implements a pool of bytes.Buffers in the form of a bounded
// channel. Buffers are pre-allocated to the requested size.
type LeakyBuffer struct {
	c chan *bytes.Buffer
	a int
}

// NewLeakyBuffer creates a new BufferPool bounded to the given size.
func NewLeakyBuffer(size, alloc int) *LeakyBuffer {
	return &LeakyBuffer{
		c: make(chan *bytes.Buffer, size),
		a: alloc,
	}
}

// Get gets a Buffer from the LeakyBuffer, or creates a new one if none are
// available in the pool. Buffers have a pre-allocated capacity.
func (lb *LeakyBuffer) Get() (b *bytes.Buffer) {
	select {
	case b = <-lb.c:
	// reuse existing buffer
	default:
		// create new buffer
		b = bytes.NewBuffer(make([]byte, 0, lb.a))
	}
	return
}

// Put returns the given Buffer to the LeakyBuffer.
func (lb *LeakyBuffer) Put(b *bytes.Buffer) {
	b.Reset()

	// Release buffers over our maximum capacity and re-create a pre-sized
	// buffer to replace it.
	// Note that the cap(b.Bytes()) provides the capacity from the read off-set
	// only, but as we've called b.Reset() the full capacity of the underlying
	// byte slice is returned.
	if cap(b.Bytes()) > lb.a {
		b = bytes.NewBuffer(make([]byte, 0, lb.a))
	}

	select {
	case lb.c <- b:
	default: // Discard the buffer if the pool is full.
	}
}
