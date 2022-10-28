package atomic

import "testing"

func TestTypedStoreReadReplace(t *testing.T) {
	s := NewValue[int](1)
	if 1 != s.Load() {
		t.Fatal("initial value is wrong")
		return
	}
	s.Store(2)
	if 2 != s.Load() {
		t.Fatal("replaced value is not stored")
		return
	}
	return
}
