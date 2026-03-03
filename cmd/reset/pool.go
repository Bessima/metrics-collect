package main

type Item[T any] interface {
	Reset()
}

type Pool struct {
	items []Item[any]
}

func New() *Pool {
	return &Pool{items: make([]Item[any], 0)}
}

func (p *Pool) Get() Item[any] {
	if len(p.items) > 0 {
		newLength := len(p.items) - 1
		elem := p.items[newLength]
		p.items = p.items[:newLength]
		return elem
	}
	return nil
}

func (p *Pool) Put(item Item[any]) {
	if item == nil {
		return
	}
	p.items = append(p.items, item)
}
