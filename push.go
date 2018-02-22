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

// PushSendAndroid sends push to android
func PushSendAndroid(deviceID string, title string, msg string) {
	/*data := map[string]string{
		"msg": msg,
	}*/
	androidClient := fcm.NewFcmClient(androidServerKey)
	androidClient.SetPriority(fcm.Priority_HIGH)
	androidClient.Message.To = deviceID
	//androidClient.NewFcmMsgTo(deviceID, data)
	androidClient.SetNotificationPayload(&fcm.NotificationPayload{
		Title: title,
		Body:  msg,
	})

	status, err := androidClient.Send()

	if err != nil {
		fmt.Println("push send error", err)
	}
	status.PrintResults()
}

// PushSendIPhone send push to iphone
func PushSendIPhone(deviceID string, title string, text string, sound string) {

	notification := &apns2.Notification{}
	notification.DeviceToken = deviceID
	//notification.Topic = "com.sideshow.Apns2"

	apsMsg := apnsMessage{}
	apsMsg.Aps.Alert = apnsAlert{
		Title: title,
		Body:  text,
	}
	if sound == "" {
		sound = "default"
	}
	apsMsg.Aps.Sound = sound

	notification.Payload, _ = json.Marshal(apsMsg) // See Payload section below

	res, err := iosClient.Push(notification)

	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Printf("PUSH SENT: %v %v %v\n", res.StatusCode, res.ApnsID, res.Reason)
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
