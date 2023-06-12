package utils

import (
	"sync"
)

type MapStruct[K comparable, V any] struct {
	sync.RWMutex
	M map[K]V
}

func (m *MapStruct[K, V]) Get(key K) (value V, ok bool) {
	m.RLock()
	defer m.RUnlock()
	value, ok = m.M[key]
	return
}

func (m *MapStruct[K, V]) Put(key K, value V) {
	m.Lock()
	defer m.Unlock()
	m.M[key] = value
}

func (m *MapStruct[K, V]) Keys() *[]K {
	m.RLock()
	defer m.RUnlock()
	res := make([]K, len(m.M))
	i := 0
	for k := range m.M {
		res[i] = k
		i++
	}
	return &res
}

func (m *MapStruct[K, V]) Length() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.M)
}
