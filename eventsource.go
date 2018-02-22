package zero

import (
	"bufio"
	"fmt"
)

// ServerEvents wrapper for sending server side events (SSE)
type ServerEvents struct {
	Writer    *bufio.Writer
	EventID   int64
	SessionID int64
	Die       chan bool
}

func (se *ServerEvents) Write(id int64, eventType string, dataBytes []byte) error {
	if eventType == "" {
		eventType = "message"
	}
	eventData := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, dataBytes)
	if id != 0 {
		eventData = fmt.Sprintf("id: %d\n", id) + eventData
	}
	_, err := se.Writer.Write([]byte(eventData))
	if err != nil {
		return err
	}
	se.Writer.Flush()
	return nil
}

// Push send event with autoincrement id
func (se *ServerEvents) Push(eventType string, dataBytes []byte) error {
	se.EventID++
	return se.Write(se.EventID, eventType, dataBytes)
}
