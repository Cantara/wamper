package atomic

import "sync/atomic"

type Value[t any] interface {
	CompareAndSwap(old, new t) (swapped bool)
	Load() (val t)
	Store(val t)
	Swap(new t) (old t)
}

type value[t any] struct {
	v atomic.Value
}

func NewValue[t any](v t) Value[t] {
	val := value[t]{}
	val.v.Store(v)
	return val
}

func (v value[t]) CompareAndSwap(old, new t) (swapped bool) {
	return v.v.CompareAndSwap(old, new)
}

func (v value[t]) Load() (val t) {
	return v.v.Load().(t)
}

func (v value[t]) Store(val t) {
	v.v.Store(val)
}

func (v value[t]) Swap(new t) (old t) {
	return v.v.Swap(new).(t)
}
