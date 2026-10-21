package pool

import "sync"

type Resetter interface {
	Reset()
}

type Pool[T any, PT interface {
	*T
	Resetter
}] struct {
	pool sync.Pool
}

func New[T any, PT interface {
	*T
	Resetter
}](allocator func() PT) *Pool[T, PT] {
	return &Pool[T, PT]{
		pool: sync.Pool{
			New: func() any {
				return allocator()
			},
		},
	}
}

func (p *Pool[T, PT]) Get() PT {
	return p.pool.Get().(PT)
}

func (p *Pool[T, PT]) Put(item PT) {
	if item == nil {
		return
	}
	item.Reset()
	p.pool.Put(item)
}
