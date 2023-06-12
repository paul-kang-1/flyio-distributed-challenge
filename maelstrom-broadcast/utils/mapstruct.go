package utils

import (
	"sync"
)

type mapStruct[K comparable, V any] struct {
	sync.RWMutex
	m map[K]V
}

func (m *mapStruct[K, V]) Get(key K) (value V, ok bool) {
	m.RLock()
	defer m.RUnlock()
	value, ok = m.m[key]
	return
}

func (m *mapStruct[K, V]) Put(key K, value V) {
	m.Lock()
	defer m.Unlock()
	m.m[key] = value
}

func (m *mapStruct[K, V]) Keys() *[]K {
	m.RLock()
	defer m.RUnlock()
	res := make([]K, len(m.m))
	i := 0
	for k := range m.m {
		res[i] = k
		i++
	}
	return &res
}

func (m *mapStruct[K, V]) Length() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.m)
}
