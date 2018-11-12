package zero

import (
	"sync"
)

// EventChannel is base type for event channel
type EventChannel chan interface{}

// EventRouter router object for routing management
type EventRouter struct {
	Channel EventChannel
}

// EventManager is an manager object for channels
type EventManager struct {
	Map        map[string][]EventChannel
	Mux        sync.RWMutex
	Routers    map[string]*EventRouter
	RoutersMux sync.RWMutex
}

var manager EventManager

// EventListener creates new event channel
func EventListener(listenerKey string) *EventRouter {
	er := EventRouter{
		Channel: make(EventChannel),
	}
	manager.RoutersMux.Lock()
	manager.Routers[listenerKey] = &er
	manager.RoutersMux.Unlock()
	return &er
}

// EventPush event to channels
func EventPush(key string, data interface{}) int {
	manager.Mux.RLock()
	channels, ok := manager.Map[key]
	if !ok {
		manager.Mux.RUnlock()
		return 0
	}
	sent := 0
	unactive := []int{}
	for i, c := range channels {
		select {
		case c <- data:
			sent++
		default:
			unactive = append(unactive, i)
		}
	}
	manager.Mux.RUnlock()
	if len(unactive) > 0 {
		manager.Mux.Lock()
		l := len(channels)
		for _, v := range unactive {
			channels[v] = channels[l-1]
			l--
		}
		channels = channels[:l]
		manager.Mux.Unlock()
	}
	return sent
}

// Subscribe to a channel
func (er *EventRouter) Subscribe(key string) {
	manager.Mux.Lock()
	channels, ok := manager.Map[key]
	if !ok {
		manager.Map[key] = []EventChannel{}
	}
	manager.Map[key] = append(channels, er.Channel)
	manager.Mux.Unlock()
}

// Unsubscribe to a channel
func (er *EventRouter) Unsubscribe(key string) {
	manager.Mux.Lock()
	channels, ok := manager.Map[key]
	if !ok {
		for i, v := range channels {
			if v == er.Channel {
				//channels = append(channels[:i], channels[i+1:]...)

				channels[i] = channels[len(channels)-1]
				channels[len(channels)-1] = nil
				channels = channels[:len(channels)-1]

				// memoryleak avoiding implementation
				//copy(channels[i:], channels[i+1:])
				//channels[len(channels)-1] = nil // or the zero value of T
				//channels = channels[:len(channels)-1]
			}
		}
		//manager.Map[objectID] = []EventChannel{}
	}
	manager.Mux.Unlock()
}

// EventSubscribe will subscribe specific router to the specific listener key
func EventSubscribe(routerKey, subscribingKey string) {
	manager.RoutersMux.RLock()
	router, ok := manager.Routers[routerKey]
	manager.RoutersMux.RUnlock() // unlocking map as soon as possible
	if ok {
		router.Subscribe(subscribingKey)
	}
}

func init() {
	manager = EventManager{
		Map:     map[string][]EventChannel{},
		Routers: map[string]*EventRouter{},
	}
}
