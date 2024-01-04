package objpool_test

import (
	"github.com/opencoff/go-objpool"
	"testing"
)

// Basic sanity tests
func TestBasic(t *testing.T) {
	assert := newAsserter(t)

	size := 3

	o := objpool.New[int](size)

	assert(o.Avail() == size, "pool: exp %d, saw %d", size, o.Avail())

	p := o.Get()
	assert(p != nil, "0: expected obj; got nil")
	assert(o.Avail() == size-1, "pool: exp %d, saw %d", size-1, o.Avail())

	o.Put(p)
	assert(o.Avail() == size, "pool: exp %d, saw %d", size, o.Avail())
}

func TestAll(t *testing.T) {
	assert := newAsserter(t)

	size := 3

	o := objpool.New[int](size)

	arr := make([]*int, size)

	for i := 0; i < size; i++ {
		p := o.Get()
		assert(p != nil, "%d: expected obj; got nil", i)
		arr[i] = p
	}

	assert(o.Avail() == 0, "expected pool to be empty, saw %d", o.Avail())

	p := o.Get()
	assert(p == nil, "%s:\nexp nil ptr", p)

	for i := 0; i < size; i++ {
		o.Put(arr[i])
	}

	assert(o.Avail() == size, "size: exp %d, saw %d", size, o.Avail())
}
