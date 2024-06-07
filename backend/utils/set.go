package utils

type Identifiable[T comparable] interface {
	ID() T
}

type ImmutableSet[T comparable, I Identifiable[T]] interface {
	Size() int
	Next() (I, bool)
	Reset()
	Slice() []I
}

type linkedHashSet[T comparable, I Identifiable[T]] struct {
	order   []I
	iterIdx int
}

// Next implements ImmutableSet.
func (h *linkedHashSet[T, I]) Next() (I, bool) {
	ptrI := new(I)
	zeroI := *ptrI

	if h.iterIdx >= len(h.order) {
		return zeroI, false
	}

	defer func() {
		h.iterIdx++
	}()

	return h.order[h.iterIdx], true
}

// Reset implements ImmutableSet.
func (h *linkedHashSet[T, I]) Reset() {
	h.iterIdx = 0
}

// Size implements ImmutableSet.
func (h *linkedHashSet[T, I]) Size() int {
	return len(h.order)
}

// Slice implements ImmutableSet.
func (h *linkedHashSet[T, I]) Slice() []I {
	clone := make([]I, h.Size())
	copy(clone, h.order)

	return clone
}

func NewLinkedHashSet[T comparable, I Identifiable[T]](data []I) ImmutableSet[T, I] {
	h := &linkedHashSet[T, I]{
		order: make([]I, 0, len(data)),
	}

	hdata := make(map[T]bool, len(data))

	for _, d := range data {
		if !hdata[d.ID()] {
			hdata[d.ID()] = true
			h.order = append(h.order, d)
		}
	}

	return h
}
