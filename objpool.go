// objpool.go - generic, fixed size object pool
//
// Copyright 2024- Sudhi Herle <sw-at-herle-dot-net>
// License: BSD-2-Clause

// objpool implements a concurrency-safe fixed size object pool.
// The objects in the pool are allocated once to reduce memory fragmentation
// and GC pressure. Every object in the pool has a live reference till the pool
// is deleted.
package objpool

import (
	"fmt"
	"sync"
)

// Pool represents a fixed pool of objects for type 'T'. Callers can allocate/free
// individual objects from the pool.
type Pool[T any] struct {
	mu sync.Mutex

	rd, wr int
	avail  int

	q   []*T
	arr []T
}

// New creates a new pool of 'sz' objects of type 'T'
func New[T any](sz int) *Pool[T] {
	arr := make([]T, sz)
	q := make([]*T, sz)

	// now enq pointers to each elem
	for i := range arr {
		q[i] = &arr[i]
	}

	// the pool starts off as "full"; it is full of
	// unconsumed objects
	o := &Pool[T]{
		rd:    0,
		wr:    0,
		avail: sz,
		q:     q,
		arr:   arr,
	}
	return o
}

// Reset resets the pool to its initial state; all extant allocations
// are reclaimed for reuse.
func (p *Pool[T]) Reset() {
	p.mu.Lock()
	p.rd = 0
	p.wr = 0
	p.avail = len(p.q)
	for i := 0; i < len(p.q); i++ {
		p.q[i] = &p.arr[i]
	}
	p.mu.Unlock()
}

// Get returns a single object from the pool. It returns nil if the pool
// has exhausted its capacity.
func (p *Pool[T]) Get() *T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.avail == 0 {
		return nil
	}

	var rd int
	rd, p.rd = p.rd, p.inc(p.rd)
	p.avail -= 1
	return p.q[rd]
}

// Put returns the object back to the pool
func (p *Pool[T]) Put(x *T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// in a well behaved system, we should never have a queue full
	// condition. It can only happen if we have a double free somewhere!
	if p.avail == len(p.q) {
		msg := fmt.Sprintf("%T: unexpected q-full", p)
		panic(msg)
	}

	var wr int
	wr, p.wr = p.wr, p.inc(p.wr)
	p.avail += 1
	p.q[wr] = x
}

// Avail returns number of free objects in the pool
func (p *Pool[T]) Avail() int {
	p.mu.Lock()
	n := p.avail
	p.mu.Unlock()
	return n
}

// String returns a string description of the pool
func (p *Pool[T]) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	var s string
	if p.avail == len(p.q) {
		s = "[FULL] "
	} else if p.avail == 0 {
		s = "[EMPTY] "
	}

	return fmt.Sprintf("<%T %scap=%d, free=%d wr=%d rd=%d",
		p, s, len(p.q), p.avail, p.wr, p.rd)
}

func (p *Pool[T]) inc(i int) int {
	if i = i + 1; i >= len(p.q) {
		i = 0
	}
	return i
}
