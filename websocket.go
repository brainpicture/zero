package server

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/fasthttp-contrib/websocket"
)

type Socket struct {
	Write       chan []byte
	Read        chan []byte
	Die         chan bool
	SessionID   int64
	LastEventID int
	Finish      bool
	Token       string
}

type Request struct {
	ID     int         `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type Service struct {
	Type string `json:"type"`
}

func (soc *Socket) Resp(data interface{}) {
	dataWraped := struct {
		Resp interface{} `json:"resp"`
	}{data}
	json, _ := json.Marshal(dataWraped)
	soc.Write <- json
}

func (soc *Socket) Err(code string, text interface{}) {

	data := struct {
		Code string `json:"code"`
		Desc string `json:"desc"`
	}{
		Code: code,
		Desc: fmt.Sprintf("%s", text),
	}

	dataWraped := struct {
		Fatal interface{} `json:"fatal"`
	}{data}
	json, _ := json.Marshal(dataWraped)
	soc.Write <- json
}

func (soc *Socket) Kill() {
	soc.Finish = true
	soc.Die <- true
}

func (soc *Socket) Service(serviceType string) {
	service := Service{
		Type: serviceType,
	}
	dataWraped := struct {
		Service *Service `json:"service"`
	}{&service}
	json, _ := json.Marshal(dataWraped)
	soc.Write <- json
}

func (srv *Server) UpgradeWS(soc *Socket) {
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

					if soc.Finish { //unsafe
						return
					}
					if message != nil {
						//fmt.Print("WRITE:", string(message))
						c.SetWriteDeadline(time.Now().Add(30 * time.Second))
						err := c.WriteMessage(websocket.TextMessage, message)
						if err != nil {
							log.Println("write:", err)
							break
						}
					}
				case <-soc.Die:
					return
				}

			}

		}()

		go func() {
			for {
				_, message, err := c.ReadMessage()
				//fmt.Print("READ:", msgType, string(message), err)

				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
						soc.Kill()
						return
					}
					log.Println("read:", err)
				}
				soc.Read <- message
			}
		}()

		handlerSocket(soc)

		fmt.Println("Updating....")
	})

	err := upgrader.Upgrade(srv.Ctx) // returns only error, executes the handler you defined on the websocket.New before (the 'chat' function)
	if err != nil {
		fmt.Println("WS upgrade error", err)
	}
}
