package zero

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/go-redis/redis"
)

// Queue is manager for event listeners with hostory of events
type Queue struct {
	redis         *redis.Client
	limit         int64
	channels      map[string]map[chan QueueEvent]bool
	channelsRWMux sync.RWMutex
}

// QueueEvent is an event objects stored in the queue
type QueueEvent struct {
	ID        int64
	Key       string
	Type      string
	UserID    int64
	SessionID int64
	Data      []byte
}

// QueueChan is a struct to controll Queue channels
type QueueChan struct {
	q    *Queue
	Chan chan QueueEvent
	keys map[string]bool
}

// QueueNew is a constructor of queue, limit == 0 means no history used
func QueueNew(redis *redis.Client, limit int64) *Queue {
	return &Queue{
		redis:    redis,
		limit:    limit,
		channels: map[string]map[chan QueueEvent]bool{},
	}
}

// QueueDataParse parse
func QueueDataParse(dataRaw string) *QueueEvent {
	chunks := strings.SplitN(dataRaw, " ", 3)
	if len(chunks) != 3 {
		return nil
	}
	eventID, sessionID, userID := SplitTrippleInt64(chunks[0], ":")
	return &QueueEvent{
		ID:        eventID,
		UserID:    userID,
		SessionID: sessionID,
		Type:      chunks[1],
		Data:      []byte(chunks[2]),
	}
}

// ListenAll will start listen process
func (q *Queue) ListenAll(pattern string) {
	pubsub := q.redis.PSubscribe("q." + pattern)
	fmt.Println("[REDIS] listen", "q."+pattern)
	_, err := pubsub.Receive()
	if err != nil {
		panic(err)
	}
	ch := pubsub.Channel()
	go func() {
		for {
			//fmt.Println("wait new ev")
			msg, ok := <-ch
			//fmt.Println("ev is here")

			if !ok {
				fmt.Println("[REDIS], msg ont ok")
				break
			}
			//fmt.Println("event from redis", msg.Channel, msg)
			if len(msg.Channel) < 3 {
				break
			}
			key := msg.Channel[2:]
			q.channelsRWMux.RLock()
			chArr, ok := q.channels[key]
			q.channelsRWMux.RUnlock()
			if ok { // channel found
				event := QueueDataParse(msg.Payload)
				if event == nil {
					continue
				}
				event.Key = msg.Channel
				//fmt.Println(">>>>>", event)
				for ch := range chArr {
					ch <- *event // all the channels should be nonblocking
				}
				//fmt.Println("<<<<< PUSHED")
			}
		}
	}()
}

// EventID will return new event ID for push method
func (q *Queue) EventID() int64 {
	return Now64()
}

// Push event to local subscribers
// - save param means the event will be stored in queue
func (q *Queue) Push(key string, eventType string, userID, sessionID, eventID int64, eventBytes []byte, save bool) error {
	var eventData string
	if sessionID == 0 && userID == 0 {
		eventData = J(eventID)
	} else {
		if userID == 0 {
			eventData = J(eventID, ":", sessionID)
		} else {
			eventData = J(eventID, ":", sessionID, ":", userID)
		}
	}
	data := eventData + " " + eventType + " " + string(eventBytes)
	fullKey := "q." + key
	if q.limit > 0 && save {
		num, err := q.redis.RPush(fullKey, data).Result()
		if err != nil {
			fmt.Println("PUSHING HISTORY FAIL", err)
		}
		if num > q.limit*2 {
			q.redis.LTrim(fullKey, int64(-q.limit), -1)
		}
	}

	return q.redis.Publish("q."+key, data).Err()
}

// GetHistory return list of events from lastEventID
func (q *Queue) GetHistory(key string, lastEventID int64) ([]*QueueEvent, bool, error) {
	events := []*QueueEvent{}
	if q.limit == 0 {
		return events, false, errors.New("this queue doesnt store any events")
	}
	fullKey := "q." + key
	eventsStr, err := q.redis.LRange(fullKey, int64(-q.limit), -1).Result()
	//fmt.Println("FETCHING History", fullKey, "params", 0, int64(q.limit))
	//eventsStr, err := q.redis.LRange(fullKey, 0, q.limit).Result()
	if err != nil {
		fmt.Println("redis scanslice error", err)
		return events, false, err
	}
	skipped := 0
	for _, dataRaw := range eventsStr {
		event := QueueDataParse(dataRaw)
		if event == nil {
			continue
		}
		event.Key = key

		if event.ID <= lastEventID {
			skipped++
			continue
		}
		events = append(events, event)
	}
	if skipped == 0 && len(eventsStr) >= int(q.limit) {
		return events, true, nil // drop results
	}
	return events, false, nil
}

// Chan will return QueueChan object to controll channel
func (q *Queue) Chan() QueueChan {
	ch := make(chan QueueEvent)
	return QueueChan{
		q:    q,
		Chan: ch,
		keys: map[string]bool{},
	}
}

// Subsribe will subscribe chan to specific key
func (qc *QueueChan) Subsribe(key string) {
	q := qc.q
	q.channelsRWMux.Lock()
	defer q.channelsRWMux.Unlock()
	_, ok := q.channels[key]
	if !ok {
		q.channels[key] = map[chan QueueEvent]bool{}
	}
	q.channels[key][qc.Chan] = true
	qc.keys[key] = true
	//qc.log(key)
}

func (qc *QueueChan) log(key string) {
	q := qc.q
	fmt.Println("whats inside:")
	for k, v := range q.channels {
		fmt.Println("  k: ", k)
		for _ = range v {
			fmt.Println("     - some chan")
		}
	}
}

// Unsubscribe will remove subscription from specific key
func (qc *QueueChan) Unsubscribe(key string) {
	q := qc.q
	q.channelsRWMux.Lock()
	defer q.channelsRWMux.Unlock()
	chArr, ok := q.channels[key]
	//fmt.Println("------ unsubscribing channel", key, ok)
	if ok {
		delete(q.channels[key], qc.Chan)
		cnt := 0
		for _ = range chArr {
			cnt++
		}
		if cnt == 0 {
			delete(q.channels, key)
		}
	}
	delete(qc.keys, key)
	//qc.log(key)
}

// UnsubscribeAll remove all subscribtions
func (qc *QueueChan) UnsubscribeAll() {
	for key := range qc.keys {
		//fmt.Println("~~~~~ unsubscribe", key)
		qc.Unsubscribe(key)
	}
}

// QueueDuplicate is special type to check if event is duplicated
type QueueDuplicate struct {
	items []*QueueEvent
	limit int
}

// QueueDuplcateNew will create new class for duplicate checking
func QueueDuplcateNew(limit int) QueueDuplicate {
	return QueueDuplicate{
		items: []*QueueEvent{},
		limit: limit,
	}
}

// Check will check an event and add this event to the list for future occurance checking
func (qd *QueueDuplicate) Check(event *QueueEvent) bool {
	for _, v := range qd.items {
		//fmt.Println("SESSION", v.SessionID, event.SessionID)
		//fmt.Println("USERID", v.UserID, event.UserID)
		if v.ID == event.ID && v.SessionID == event.SessionID {
			//fmt.Println("DATA", len(v.Data), len(event.Data))
			if bytes.Equal(v.Data, event.Data) {
				//fmt.Println("EQUAL")
				return true
			}
		}
	}
	qd.items = append(qd.items, event)
	if len(qd.items) > qd.limit {
		qd.items = qd.items[1:]
	}
	return false
}
