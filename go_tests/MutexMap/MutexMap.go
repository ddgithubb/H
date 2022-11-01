package MutexMap

import "sync"

type MutexStringChannelMap struct {
	bare_map map[string](chan string)
	mutex sync.RWMutex
}

func NewMutexStringChannelMap() *MutexStringChannelMap {
	return &MutexStringChannelMap{
		bare_map: make(map[string](chan string)),
	}
}

func (m *MutexStringChannelMap) Set(key string, channel (chan string)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.bare_map[key] = channel
}

func (m *MutexStringChannelMap) Get(key string) (channel (chan string), ok bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	res, ok := m.bare_map[key]
	return res, ok
}

func (m *MutexStringChannelMap) Delete(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.bare_map, key)
}