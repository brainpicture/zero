package zero

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/NaySoftware/go-fcm"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

var androidServerKey = ""
var iosProduction = apns2.Client{}
var iosSandbox = apns2.Client{}
var iosBundleProd = ""
var iosBundleSandbox = ""

// Push describes object for push notifications
type Push struct {
	Title            string
	Body             string
	Type             string
	Bundle           string
	Grouping         string
	Data             H
	NoSound          bool
	Badge            int  // ios flag for badge count
	ContentAvailable bool // ios flag to trigger code execution
	Mutable          bool
	Category         string
	Collapse         string
	// Sender is prefix which whould be placed before Text field, like Sender: Text on ios, and as separate field on android
	Sender string
	Silent bool
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

// Send send push to the device
func (push *Push) Send(platform, deviceToken string, sandbox bool) error {
	if platform == "android" {
		client := fcm.NewFcmClient(androidServerKey)
		client.SetPriority(fcm.Priority_HIGH)
		pushData := H{
			"type":     push.Type,
			"title":    push.Title,
			"text":     push.Body,
			"body":     push.Data,
			"no_sound": push.NoSound,
		}
		if push.Sender != "" {
			pushData["sender"] = push.Sender
		}
		client.NewFcmMsgTo(deviceToken, pushData)
		_, err := client.Send()
		if err != nil {
			fmt.Println("ANDROID! push send error", err)
			return err
		} else {
			//fmt.Println("ANDROID! resp", resp, J(pushData))
		}
	} else if platform == "ios" {
		notification := &apns2.Notification{}

		notification.DeviceToken = deviceToken
		if push.Bundle != "" {
			notification.Topic = push.Bundle
		} else {
			if sandbox {
				notification.Topic = iosBundleSandbox
			} else {
				notification.Topic = iosBundleProd
			}
		}
		sound := "default"
		if push.NoSound {
			sound = "" // no sound should be here
		}
		body := push.Body
		if push.Sender != "" {
			body = push.Sender + ": " + body
		}
		aps := H{}
		if !push.Silent {
			aps["alert"] = apnsAlert{
				Title: push.Title,
				Body:  body,
			}
			if sound != "" {
				aps["sound"] = sound
			}
		}
		if push.Grouping != "" {
			aps["thread-id"] = push.Grouping
		}
		if push.ContentAvailable {
			aps["content-available"] = 1
		}
		if push.Badge != 0 {
			aps["badge"] = push.Badge
		}
		if push.Mutable {
			aps["mutable-content"] = 1
		}
		if push.Category != "" {
			aps["category"] = push.Category
		}
		if push.Collapse != "" {
			notification.CollapseID = push.Collapse
		}
		apsMsg := H{
			"aps": aps,
			"data": H{
				"type": push.Type,
				"body": push.Data,
			},
		}

		notification.Payload, _ = json.Marshal(apsMsg) // See Payload section below

		var client apns2.Client
		if sandbox {
			client = iosSandbox
		} else {
			client = iosProduction
		}

		resp, err := client.Push(notification)
		if err != nil {
			return err
		} else {
			if resp.StatusCode != 200 {
				return errors.New(J(resp.StatusCode, ": ", resp.Reason))
			}
		}
	}

	return nil
}

// PushInitIOS init push notifications for ios
func PushInitIOS(certProduction, passProduction, bundleProd, certSandbox, passSandbox, bundleSandbox string) {
	cert, err := certificate.FromP12File(certProduction, passProduction)
	if err != nil {
		fmt.Println("Production Cert Error:", err)
	} else {
		iosProduction = *apns2.NewClient(cert).Production()
	}
	iosBundleProd = bundleProd

	cert, err = certificate.FromP12File(certSandbox, passSandbox)
	if err != nil {
		fmt.Println("Sandbox Cert Error:", err)
	} else {
		iosSandbox = *apns2.NewClient(cert).Development()
	}
	iosBundleSandbox = bundleSandbox
}

// PushInitAndroid init push notifications for android
func PushInitAndroid(serverKey string) {
	androidServerKey = serverKey
}
