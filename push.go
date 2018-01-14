package zero

import (
	"encoding/json"
	"fmt"

	"github.com/NaySoftware/go-fcm"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

var AndroidServerKey = ""

type ApnsMessage struct {
	Aps struct {
		Alert string `json:"alert"`
		Sound string `json:"sound"`
	} `json:"aps"`
}

func PushSendAndroid(deviceID string, msg string) {
	data := map[string]string{
		"msg": msg,
	}

	c := fcm.NewFcmClient(AndroidServerKey)
	c.NewFcmMsgTo(deviceID, data)

	_, err := c.Send()

	if err != nil {
		fmt.Println("push send error", err)
	}
}

func PushSendIPhone(deviceID string, msg string) {
	cert, err := certificate.FromP12File("./certificates/cert.p12", "")
	if err != nil {
		fmt.Printf("Cert Error:", err)
	}

	notification := &apns2.Notification{}
	notification.DeviceToken = deviceID
	//notification.Topic = "com.sideshow.Apns2"

	apsMsg := ApnsMessage{}
	apsMsg.Aps.Alert = msg
	apsMsg.Aps.Sound = "default"

	notification.Payload, _ = json.Marshal(apsMsg) // See Payload section below

	client := apns2.NewClient(cert).Development()
	res, err := client.Push(notification)

	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Printf("PUSH SENT: %v %v %v\n", res.StatusCode, res.ApnsID, res.Reason)
}
