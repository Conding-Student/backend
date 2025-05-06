package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"golang.org/x/oauth2/google"
)

type FCMMessage struct {
	Message struct {
		Token        string            `json:"token"`
		Notification map[string]string `json:"notification,omitempty"`
		Data         map[string]string `json:"data,omitempty"`
	} `json:"message"`
}

func SendPushNotification(fcmToken, title, body, conversationId string) error {
	serviceAccountPath := "config/rentxpert-a987d-firebase-adminsdk-fbsvc-682d5e8e0b.json"
	jsonKey, err := os.ReadFile(serviceAccountPath)
	if err != nil {
		return fmt.Errorf("failed to read service account file: %v", err)
	}

	conf, err := google.JWTConfigFromJSON(jsonKey, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return fmt.Errorf("failed to parse JWT config: %v", err)
	}

	client := conf.Client(context.Background())

	var serviceAccount map[string]interface{}
	if err := json.Unmarshal(jsonKey, &serviceAccount); err != nil {
		return fmt.Errorf("failed to parse service account: %v", err)
	}
	projectId := serviceAccount["project_id"]

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectId)

	msg := FCMMessage{}
	msg.Message.Token = fcmToken
	msg.Message.Notification = map[string]string{
		"title": title,
		"body":  body,
	}
	msg.Message.Data = map[string]string{
		"conversationId": conversationId,
		"click_action":   "FLUTTER_NOTIFICATION_CLICK",
	}

	bodyBytes, _ := json.Marshal(msg)

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("push failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("push failed: %s", string(respBody))
	}

	fmt.Println("Push success:", string(respBody))
	return nil
}
