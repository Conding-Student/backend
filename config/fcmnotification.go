package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type FCMMessage struct {
	Message struct {
		Token        string            `json:"token"`
		Notification map[string]string `json:"notification,omitempty"`
		Data         map[string]string `json:"data,omitempty"`
	} `json:"message"`
}

type NotificationLog struct {
	ReceiverID      string    `firestore:"receiver_id"`
	SenderID        string    `firestore:"sender_id,omitempty"`
	ConversationID  string    `firestore:"conversation_id"`
	FCMMessageID    string    `firestore:"fcm_message_id,omitempty"`
	Status          string    `firestore:"status"` // "sent", "delivered", "opened", "failed"
	Error           string    `firestore:"error,omitempty"`
	Timestamp       time.Time `firestore:"timestamp"`
	DeliveryAttempt int       `firestore:"delivery_attempt"`
}

var (
	firestoreClient *firestore.Client
	projectID       string
)

func init() {
	ctx := context.Background()
	serviceAccountPath := "config/rentxpert-a987d-firebase-adminsdk-fbsvc-682d5e8e0b.json"
	
	// Initialize Firestore client
	sa := option.WithCredentialsFile(serviceAccountPath)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v\n", err)
	}

	firestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error initializing Firestore: %v\n", err)
	}

	// Extract project ID
	jsonKey, err := os.ReadFile(serviceAccountPath)
	if err != nil {
		log.Fatalf("Failed to read service account file: %v\n", err)
	}

	var serviceAccount map[string]interface{}
	if err := json.Unmarshal(jsonKey, &serviceAccount); err != nil {
		log.Fatalf("Failed to parse service account: %v\n", err)
	}
	projectID = serviceAccount["project_id"].(string)
}

func SendPushNotification(fcmToken, title, body, conversationId, senderId string) error {
	ctx := context.Background()
	startTime := time.Now()
	
	// Create initial log entry
	logRef := firestoreClient.Collection("notification_logs").NewDoc()
	initialLog := NotificationLog{
		ReceiverID: func() string {
			receiverID, err := extractReceiverIdFromToken(fcmToken)
			if err != nil {
				log.Printf("Error extracting receiver ID: %v\n", err)
				return "unknown"
			}
			return receiverID
		}(),
		SenderID:        senderId,
		ConversationID:  conversationId,
		Status:          "attempting",
		Timestamp:       startTime,
		DeliveryAttempt: 1,
	}

	if _, err := logRef.Set(ctx, initialLog); err != nil {
		log.Printf("Failed to create initial log entry: %v\n", err)
	}

	// Prepare FCM message
	msg := FCMMessage{}
	msg.Message.Token = fcmToken
	msg.Message.Notification = map[string]string{
		"title": title,
		"body":  body,
	}
	msg.Message.Data = map[string]string{
		"conversationId": conversationId,
		"senderId":       senderId,
		"click_action":   "FLUTTER_NOTIFICATION_CLICK",
		"logId":          logRef.ID, // Reference back to our log
	}

	bodyBytes, _ := json.Marshal(msg)

	// Get authenticated client
	jsonKey, err := os.ReadFile("config/rentxpert-a987d-firebase-adminsdk-fbsvc-682d5e8e0b.json")
	if err != nil {
		return logAndUpdate(ctx, logRef, "failed", fmt.Sprintf("failed to read service account file: %v", err))
	}

	conf, err := google.JWTConfigFromJSON(jsonKey, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return logAndUpdate(ctx, logRef, "failed", fmt.Sprintf("failed to parse JWT config: %v", err))
	}

	client := conf.Client(ctx)
	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectID)

	// Send with retry logic
	maxAttempts := 3
	var lastError error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Printf("Attempt %d to send notification to %s\n", attempt, initialLog.ReceiverID)
		
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
		if err != nil {
			lastError = fmt.Errorf("push failed: %v", err)
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
			continue
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		
		if resp.StatusCode != 200 {
			lastError = fmt.Errorf("push failed: %s", string(respBody))
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		// Parse FCM response
		var result map[string]interface{}
		if err := json.Unmarshal(respBody, &result); err != nil {
			return logAndUpdate(ctx, logRef, "failed", fmt.Sprintf("failed to parse FCM response: %v", err))
		}

		fcmMessageID := result["name"].(string)
		log.Printf("Successfully sent notification to %s (FCM ID: %s)\n", initialLog.ReceiverID, fcmMessageID)

		// Update log with success
		_, err = logRef.Update(ctx, []firestore.Update{
			{Path: "status", Value: "sent"},
			{Path: "fcm_message_id", Value: fcmMessageID},
			{Path: "delivery_attempt", Value: attempt},
			{Path: "timestamp", Value: time.Now()},
		})
		if err != nil {
			log.Printf("Failed to update log entry: %v\n", err)
		}

		return nil
	}

	return logAndUpdate(ctx, logRef, "failed", lastError.Error())
}

func logAndUpdate(ctx context.Context, logRef *firestore.DocumentRef, status, errorMsg string) error {
	_, err := logRef.Update(ctx, []firestore.Update{
		{Path: "status", Value: status},
		{Path: "error", Value: errorMsg},
		{Path: "timestamp", Value: time.Now()},
	})
	if err != nil {
		log.Printf("Failed to update failed log entry: %v\n", err)
	}
	return errors.New(errorMsg)
}


func extractReceiverIdFromToken(fcmToken string) (string, error) {
    if fcmToken == "" {
        return "unknown", nil
    }

    ctx := context.Background()
    
    // Query all user_tokens documents where tokens array contains our fcmToken
    iter := firestoreClient.Collection("user_tokens").
        Where("tokens", "array-contains", fcmToken).
        Limit(1).
        Documents(ctx)

    doc, err := iter.Next()
    if err == iterator.Done {
        log.Printf("No user found with FCM token: %s", fcmToken)
        return "unknown", nil
    }
    if err != nil {
        log.Printf("Error querying Firestore: %v", err)
        return "", err
    }

    // The document ID is the user ID in this structure
    userId := doc.Ref.ID
    log.Printf("Found user %s for FCM token %s", userId, fcmToken)
    return userId, nil
}

// Call this when a notification is opened in Flutter
func TrackNotificationOpen(logId string) error {
	ctx := context.Background()
	_, err := firestoreClient.Collection("notification_logs").Doc(logId).Update(ctx, []firestore.Update{
		{Path: "status", Value: "opened"},
		{Path: "opened_at", Value: time.Now()},
	})
	return err
}