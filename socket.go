package zero

import (
	"encoding/json"
	"fmt"
)

// RespService service message type
type RespService struct {
	Type string `json:"type"`
}

// Socket is an
type Socket struct {
	Write  chan []byte
	Read   chan []byte
	Die    chan bool
	Finish bool
}

// Fatal push an fatal error to the socket
func (soc *Socket) Fatal(code string, text interface{}) {
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

// Kill end end conection
func (soc *Socket) Kill() {
	soc.Finish = true
	soc.Die <- true
}

// Send an service event to the user
func (soc *Socket) Service(serviceType string) {
	service := RespService{
		Type: serviceType,
	}
	dataWraped := struct {
		Service *RespService `json:"service"`
	}{&service}
	json, _ := json.Marshal(dataWraped)
	soc.Write <- json
}
