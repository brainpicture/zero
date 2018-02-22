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
	Map map[string][]EventChannel
	Mux sync.RWMutex
}

var manager EventManager

// EventListener creates new event channel
func EventListener() EventRouter {
	return EventRouter{
		Channel: make(EventChannel),
	}
}

// EventPush event to channels
func EventPush(key string, data interface{}) {
	manager.Mux.RLock()
	channels, ok := manager.Map[key]
	if ok {
		for _, c := range channels {
			select {
			case c <- data:
			default:
			}
		}
	}
	manager.Mux.RUnlock()
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

func init() {
	manager = EventManager{
		Map: map[string][]EventChannel{},
	}
}
