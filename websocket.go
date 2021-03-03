package zero

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/fasthttp-contrib/websocket"
)

// UpgradeWS upgrades http request to websocket
func (req *Request) UpgradeWS(cb func(req *Request)) *Socket {
	soc := Socket{
		Write: make(chan []byte),
		Read:  make(chan []byte),
		Die:   make(chan bool),
	}
	upgrader := websocket.New(func(c *websocket.Conn) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("UNCATCHED PANIC", r)
				debug.PrintStack()
			}
		}()

		go func() {
			for {
				select {
				case message := <-soc.Write:
					if c == nil {
						return
					}

					c.SetWriteDeadline(time.Now().Add(30 * time.Second))
					err := c.WriteMessage(websocket.TextMessage, message)
					if err != nil {
						log.Println("write:", err)
						break
					}
				case <-soc.Die:
					return
				}

			}

		}()

		go func() {
			for {
				_, message, err := c.ReadMessage()

				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
						soc.Kill()
						return
					}
				}
				select {
				case soc.Read <- message:
				default:
				}
			}
		}()
	})

	err := upgrader.Upgrade(req.Ctx) // returns only error, executes the handler you defined on the websocket.New before (the 'chat' function)
	if err != nil {
		fmt.Println("WS upgrade error", err)
	}
	return &soc
}
