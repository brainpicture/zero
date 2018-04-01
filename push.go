package zero

import (
	"encoding/json"
	"fmt"

	"github.com/NaySoftware/go-fcm"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

var androidServerKey = ""
var iosClient = apns2.Client{}

type Push struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Data  H      `json:"data"`
}

type apnsAlert struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type apnsMessage struct {
	Aps struct {
		Title string    `json:"title"`
		Alert apnsAlert `json:"alert"`
		Sound string    `json:"sound"`
	} `json:"aps"`
}

// PushSend send push to the device
func PushSend(platform, deviceID string, push *Push) {
	if platform == "android" {
		//zero.PushSendAndroid(sett.DeviceID, title, msg)
		client := fcm.NewFcmClient(androidServerKey)
		client.SetPriority(fcm.Priority_HIGH)
		client.NewFcmMsgTo(deviceID, push)
		_, err := client.Send()
		if err != nil {
			fmt.Println("android push send error", err)
		}
	} else if platform == "ios" {
		notification := &apns2.Notification{}
		notification.DeviceToken = deviceID
		apsMsg := H{
			"aps": H{
				"sound": "default",
				"alert": apnsAlert{
					Title: push.Title,
					Body:  push.Body,
				},
			},
			"data": push.Data,
		}

		notification.Payload, _ = json.Marshal(apsMsg) // See Payload section below

		//fmt.Println("push Sending", string(notification.Payload.([]byte)))
		_, err := iosClient.Push(notification)
		//fmt.Println("push Sent", resp, err)

		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}

// PushInitIOS init push notifications for ios
func PushInitIOS(p12File string, password string, development bool) {
	cert, err := certificate.FromP12File(p12File, password)
	if err != nil {
		fmt.Println("Cert Error:", err)
		return
	}
	if development {
		iosClient = *apns2.NewClient(cert).Development()
	} else {
		iosClient = *apns2.NewClient(cert).Production()
	}
}

// PushInitAndroid init push notifications for android
func PushInitAndroid(serverKey string) {
	androidServerKey = serverKey
}
