package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Message Structures
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
	Status          string    `firestore:"status"`
	Error           string    `firestore:"error,omitempty"`
	Timestamp       time.Time `firestore:"timestamp"`
	DeliveryAttempt int       `firestore:"delivery_attempt"`
	Title           string    `firestore:"title,omitempty"`
	Body            string    `firestore:"body,omitempty"`
}

// Global variables
var (
	firestoreClient *firestore.Client
	projectID       string
)

// Initialization
func init() {
	initializeFirebase()
}
// SendNotificationHandler handles sending notifications
func SendNotificationHandler(c *fiber.Ctx) error {
	type request struct {
		FCMToken       string `json:"fcmToken"`
		Title          string `json:"title"`
		Body           string `json:"body"`
		ConversationID string `json:"conversationId"`
		SenderID       string `json:"senderId"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	go SendPushNotification(req.FCMToken, req.Title, req.Body, req.ConversationID, req.SenderID)

	return c.JSON(fiber.Map{
		"message": "Notification processing started",
	})
}

// TrackNotificationOpenHandler handles tracking notification opens
func TrackNotificationOpenHandler(c *fiber.Ctx) error {
	logId := c.Params("logId")
	if logId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "logId is required",
		})
	}

	if err := TrackNotificationOpen(logId); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to track notification open",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Notification marked as opened",
	})
}

func initializeFirebase() {
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
	projectID = extractProjectID(serviceAccountPath)
}

func extractProjectID(serviceAccountPath string) string {
	jsonKey, err := os.ReadFile(serviceAccountPath)
	if err != nil {
		log.Fatalf("Failed to read service account file: %v\n", err)
	}

	var serviceAccount map[string]interface{}
	if err := json.Unmarshal(jsonKey, &serviceAccount); err != nil {
		log.Fatalf("Failed to parse service account: %v\n", err)
	}
	return serviceAccount["project_id"].(string)
}


// Core Functions
func SendPushNotification(fcmToken, title, body, conversationId, senderId string) {
	ctx := context.Background()

	// Always send the push notification
	go sendPushOnly(fcmToken, title, body, conversationId, senderId)

	// Log only once per conversation
	go func() {
		if conversationId != "general" {
	hasExisting, err := hasExistingNotification(ctx, conversationId, senderId)
	if err != nil {
		log.Printf("Error checking existing log: %v", err)
		return
	}
	if hasExisting {
		log.Printf("Skipping log creation for conversation %s", conversationId)
		return
	}
}

		receiverID, _ := extractReceiverIdFromToken(fcmToken)
		logRef := firestoreClient.Collection("notification_logs").NewDoc()
		logEntry := NotificationLog{
			ReceiverID:      receiverID,
			SenderID:        senderId,
			ConversationID:  conversationId,
			Status:          "sent",
			Timestamp:       time.Now(),
			DeliveryAttempt: 1,
			Title:           title,
			Body:            body,
		}

		if _, err := logRef.Set(ctx, logEntry); err != nil {
			log.Printf("Failed to save log: %v", err)
		}
	}()
}

func sendPushOnly(fcmToken, title, body, conversationId, senderId string) {
	ctx := context.Background()

	// Load service account key
	saKey, err := os.ReadFile("config/rentxpert-a987d-firebase-adminsdk-fbsvc-682d5e8e0b.json")
	if err != nil {
		log.Printf("Error reading service key: %v", err)
		return
	}

	conf, err := google.JWTConfigFromJSON(saKey, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		log.Printf("Error parsing JWT config: %v", err)
		return
	}

	client := conf.Client(ctx)
	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectID)

	message := map[string]interface{}{
		"message": map[string]interface{}{
			"token": fcmToken,
			"notification": map[string]string{
				"title": title,
				"body":  body,
			},
			"data": map[string]string{
				"conversationId": conversationId,
				"senderId":       senderId,
				"click_action":   "FLUTTER_NOTIFICATION_CLICK",
			},
		},
	}

	payload, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending push notification: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Push notification failed: %s", string(bodyBytes))
		return
	}

	log.Printf("Push notification sent to %s", fcmToken)
}

// Helper Functions
func hasExistingNotification(ctx context.Context, conversationId, senderId string) (bool, error) {
	iter := firestoreClient.Collection("notification_logs").
		Where("conversation_id", "==", conversationId).
		Where("sender_id", "==", senderId).
		Limit(1).
		Documents(ctx)

	_, err := iter.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func extractReceiverIdFromToken(fcmToken string) (string, error) {
	if fcmToken == "" {
		return "unknown", nil
	}

	ctx := context.Background()
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

	return doc.Ref.ID, nil
}

func TrackNotificationOpen(logId string) error {
	ctx := context.Background()
	_, err := firestoreClient.Collection("notification_logs").Doc(logId).Update(ctx, []firestore.Update{
		{Path: "status", Value: "opened"},
		{Path: "opened_at", Value: time.Now()},
	})
	return err
}

// GetNotificationsHandler returns notification logs for a user
func GetNotificationsHandler(c *fiber.Ctx) error {
	uid := c.Params("uid")
	if uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "uid is required",
		})
	}

	ctx := context.Background()

	var notifications []NotificationLog

	iter := firestoreClient.Collection("notification_logs").
		Where("receiver_id", "==", uid).
		OrderBy("timestamp", firestore.Desc).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error fetching notifications",
			})
		}

		var notif NotificationLog
		if err := doc.DataTo(&notif); err != nil {
			log.Printf("Failed to parse document: %v", err)
			continue
		}
		notifications = append(notifications, notif)
	}

	return c.JSON(fiber.Map{
		"notifications": notifications,
	})
}
