// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package server

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	fcm "github.com/appleboy/go-fcm"
	"github.com/kyokomi/emoji"
)

type AndroidNotificationServer struct {
	AndroidPushSettings AndroidPushSettings
}

func NewAndroideNotificationServer(settings AndroidPushSettings) NotificationServer {
	return &AndroidNotificationServer{AndroidPushSettings: settings}
}

func (me *AndroidNotificationServer) Initialize() bool {
	LogInfo(fmt.Sprintf("Initializing Android notification server for tag=%s", me.AndroidPushSettings.Tag))

	if len(me.AndroidPushSettings.AndroidApiKey) == 0 {
		LogError("Android push notifications not configured.  Missing AndroidApiKey.")
		return false
	}

	return true
}

func (me *AndroidNotificationServer) SendNotification(msg *PushNotification) PushResponse {
	pushType := msg.Type
	data := map[string]interface{}{
		"ack_id":     msg.AckId,
		"type":       pushType,
		"badge":      msg.Badge,
		"version":    msg.Version,
		"channel_id": msg.ChannelId,
	}
	n := fcm.Notification{}

	if msg.Badge > 0 {
		n.Badge = strconv.Itoa(msg.Badge)
	}

	hasContent := false
	if msg.IsIdLoaded || pushType == PUSH_TYPE_MESSAGE {
		hasContent = true
		m := emoji.Sprint(msg.Message)

		data["post_id"] = msg.PostId
		data["sender_id"] = msg.SenderId
		data["message"] = m
		data["click_action"] = "FLUTTER_NOTIFICATION_CLICK"
		if len(msg.OverrideUsername) > 0 {
			data["sender_name"] = msg.OverrideUsername
		}

		if msg.IsIdLoaded {
			data["id_loaded"] = true
			data["sender_name"] = "Someone"
		} else if pushType == PUSH_TYPE_MESSAGE {
			data["sender_name"] = msg.SenderName
			data["channel_name"] = msg.ChannelName
			data["root_id"] = msg.RootId
			data["override_username"] = msg.OverrideUsername
			data["override_icon_url"] = msg.OverrideIconUrl
			data["from_webhook"] = msg.FromWebhook
		}

		if msg.ChannelType == "D" {
			n.Title = "New direct message"
			n.Body = fmt.Sprintf("%s: %s", data["sender_name"], m)
		} else {
			n.Title = fmt.Sprintf("New message in %s", msg.ChannelName)
			n.Body = m
		}
	}

	incrementNotificationTotal(PUSH_NOTIFY_ANDROID, pushType)

	fcmMsg := &fcm.Message{
		To:               msg.DeviceId,
		Data:             data,
		Priority:         "high",
		Notification:     &n,
		ContentAvailable: hasContent,
	}

	if pushType == PUSH_TYPE_MESSAGE {
		logInvalidMessageFields(msg, &n)
	}

	if len(me.AndroidPushSettings.AndroidApiKey) > 0 {
		sender, err := fcm.NewClient(me.AndroidPushSettings.AndroidApiKey)
		if err != nil {
			incrementFailure(PUSH_NOTIFY_ANDROID, pushType, "invalid ApiKey")
			return NewErrorPushResponse(err.Error())
		}

		LogInfo(fmt.Sprintf("Sending push notification for type=%v, server=%v, device=%v", msg.Type, msg.ServerTag, msg.DeviceId))

		start := time.Now()

		test, testErr := json.Marshal(fcmMsg)
		if testErr == nil {
			fmt.Println(fmt.Sprintf("SendNotification: %s", string(test)))
		}

		resp, err := sender.SendWithRetry(fcmMsg, 2)
		observerNotificationResponse(PUSH_NOTIFY_ANDROID, time.Since(start).Seconds())

		if err != nil {
			LogError(fmt.Sprintf("Failed to send FCM push sid=%v did=%v err=%v type=%v", msg.ServerId, msg.DeviceId, err, me.AndroidPushSettings.Type))
			incrementFailure(PUSH_NOTIFY_ANDROID, pushType, "unknown transport error")
			return NewErrorPushResponse("unknown transport error")
		}

		if resp.Failure > 0 {
			fcmError := resp.Results[0].Error

			if fcmError == fcm.ErrInvalidRegistration || fcmError == fcm.ErrNotRegistered || fcmError == fcm.ErrMissingRegistration {
				LogInfo(fmt.Sprintf("Android response failure sending remove code: %v type=%v", resp, me.AndroidPushSettings.Type))
				incrementRemoval(PUSH_NOTIFY_ANDROID, pushType, fcmError.Error())
				return NewRemovePushResponse()
			}

			LogError(fmt.Sprintf("Android response failure: %v type=%v", resp, me.AndroidPushSettings.Type))
			incrementFailure(PUSH_NOTIFY_ANDROID, pushType, fcmError.Error())
			return NewErrorPushResponse(fcmError.Error())
		}
	}

	if len(msg.AckId) > 0 {
		incrementSuccessWithAck(PUSH_NOTIFY_ANDROID, pushType)
	} else {
		incrementSuccess(PUSH_NOTIFY_ANDROID, pushType)
	}
	return NewOkPushResponse()
}

func logInvalidMessageFields(msg *PushNotification, n *fcm.Notification) {
	if len(msg.ChannelName) == 0 {
		LogError(fmt.Sprintf("Channel name is empty: %+v, %+v, %+v, %+v", msg.ChannelId, msg.SenderId, msg.PostId, msg.IsIdLoaded))
	}
	if len(msg.SenderName) == 0 {
		LogError(fmt.Sprintf("Sender name is empty: %+v, %+v, %+v, %+v", msg.ChannelId, msg.SenderId, msg.PostId, msg.IsIdLoaded))
	}
	if len(msg.SenderId) == 0 {
		LogError(fmt.Sprintf("Sender id is empty: %+v, %+v, %+v, %+v", msg.ChannelId, msg.SenderId, msg.PostId, msg.IsIdLoaded))
	}
	if len(n.Body) == 0 {
		LogError(fmt.Sprintf("Notification body is empty: %+v, %+v, %+v, %+v", msg.ChannelId, msg.SenderId, msg.PostId, msg.IsIdLoaded))
	}
}
