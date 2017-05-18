package potential

import (
	"sync"
)

/*
cellStore is a data structure that wraps an array to make it thread safe
when we want it. It also handles expanding and contracting, and issuing
ids that iterate but where the indexes are stored on the cells so they
can be referenced outside the array.
*/
type cellStore struct {
	_store []*Cell
	cursor int
	mux    sync.Mutex
}

func newCellStore() *cellStore {
	s := cellStore{
		_store: make([]*Cell, 5000),
		cursor: 0,
	}
	return &s
}

func (cs *cellStore) add(cell *Cell) {

}

func (cs *cellStore) compact() {

}
