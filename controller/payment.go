package controller

import (
	"bytes"

	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"time"

	"github.com/Conding-Student/backend/model"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type PayMongoService struct {
	DB        *gorm.DB
	PublicKey string
	SecretKey string
}

type SourceRequest struct {
	Data struct {
		Attributes struct {
			Amount   int    `json:"amount"`
			Currency string `json:"currency"`
			Type     string `json:"type"`
			Redirect struct {
				Success string `json:"success"`
				Failed  string `json:"failed"`
			} `json:"redirect"`
		} `json:"attributes"`
	} `json:"data"`
}

type SourceResponse struct {
	Data struct {
		ID         string `json:"id"`
		Attributes struct {
			Redirect struct {
				CheckoutURL string `json:"checkout_url"`
			} `json:"redirect"`
			Status string `json:"status"`
		} `json:"attributes"`
	} `json:"data"`
}

type WebhookPayload struct {
	Data struct {
		Attributes struct {
			Type string `json:"type"`
			Data struct {
				ID         string `json:"id"`
				Attributes struct {
					Amount int    `json:"amount"`
					Status string `json:"status"`
				} `json:"attributes"`
			} `json:"data"`
		} `json:"attributes"`
	} `json:"data"`
}

// CreateSource creates a PayMongo GCash source and saves transaction
func (s *PayMongoService) CreateSource(c *fiber.Ctx) error {
	type Request struct {
		UserID     string  `json:"user_id"`
		BaseAmount float64 `json:"base_amount"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Body parsing error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Validate input
	if req.UserID == "" || req.BaseAmount <= 0 {
		log.Printf("Invalid input: user_id=%s, base_amount=%.2f", req.UserID, req.BaseAmount)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "UserID and positive BaseAmount required"})
	}

	// Calculate 0.2% interest
	interest := req.BaseAmount * 0.002
	totalAmount := req.BaseAmount + interest
	amountInCentavos := int(totalAmount * 100) // Convert to centavos

	// Create PayMongo source
	sourceReq := SourceRequest{}
	sourceReq.Data.Attributes.Amount = amountInCentavos
	sourceReq.Data.Attributes.Currency = "PHP"
	sourceReq.Data.Attributes.Type = "gcash"
	sourceReq.Data.Attributes.Redirect.Success = "https://bd1a-103-72-190-240.ngrok-free.app/success"
	sourceReq.Data.Attributes.Redirect.Failed = "https://bd1a-103-72-190-240.ngrok-free.app/failed"

	body, err := json.Marshal(sourceReq)
	if err != nil {
		log.Printf("Failed to marshal source request: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create source request"})
	}

	client := &http.Client{}
	paymongoReq, err := http.NewRequest("POST", "https://api.paymongo.com/v1/sources", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to create PayMongo request: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create PayMongo request"})
	}
	paymongoReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.PublicKey+":")))
	paymongoReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(paymongoReq)
	if err != nil {
		log.Printf("PayMongo API error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "PayMongo API error"})
	}
	defer resp.Body.Close()

	// Read and log response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read PayMongo response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read PayMongo response"})
	}
	log.Printf("PayMongo response: status=%d, body=%s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "PayMongo source creation failed",
			"details": string(respBody),
		})
	}

	var sourceResp SourceResponse
	if err := json.Unmarshal(respBody, &sourceResp); err != nil {
		log.Printf("Failed to parse PayMongo response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse PayMongo response"})
	}

	// Verify source status
	if sourceResp.Data.Attributes.Status != "pending" {
		log.Printf("Unexpected source status: %s", sourceResp.Data.Attributes.Status)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Invalid source status",
			"details": string(respBody),
		})
	}

	// Save transaction to PostgreSQL
	txn := model.Transaction{
		UserID:           req.UserID,
		BaseAmount:       req.BaseAmount,
		InterestAmount:   interest,
		TotalAmount:      totalAmount,
		PayMongoSourceID: sourceResp.Data.ID,
		Status:           "pending",
	}
	if err := s.DB.Create(&txn).Error; err != nil {
		log.Printf("Database error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	return c.JSON(fiber.Map{
		"checkout_url": sourceResp.Data.Attributes.Redirect.CheckoutURL,
		"source_id":    sourceResp.Data.ID,
	})
}

// HandleWebhook processes PayMongo webhooks
func (s *PayMongoService) HandleWebhook(c *fiber.Ctx) error {
	// 1. Log raw request
body := c.Body()
    log.Printf("📩 Webhook received: headers=%v, body=%s", c.GetReqHeaders(), string(body))

    signature := c.Get("Paymongo-Signature")
    log.Printf("🔍 Paymongo-Signature: %s", signature)
    if signature == "" {
        log.Println("⚠️ Missing Paymongo-Signature header")
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing signature"})
    }

log.Printf("✅ Bypassing signature validation for debugging")

var webhook WebhookPayload
    if err := c.BodyParser(&webhook); err != nil {
        log.Printf("❌ Invalid webhook payload: %v", err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid webhook"})
    }

if webhook.Data.Attributes.Type != "source.chargeable" {
        log.Printf("ℹ️ Ignoring non-chargeable webhook event: %s", webhook.Data.Attributes.Type)
        return c.SendStatus(fiber.StatusOK)
    }

	// 5. Extract source ID and amount
sourceID := webhook.Data.Attributes.Data.ID
    amount := webhook.Data.Attributes.Data.Attributes.Amount
    log.Printf("🔍 Processing webhook for source: %s, amount: %d", sourceID, amount)

    var txn model.Transaction
    if err := s.DB.Where("pay_mongo_source_id = ?", sourceID).First(&txn).Error; err != nil {
        log.Printf("❌ Transaction not found for source %s: %v", sourceID, err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Transaction not found"})
    }

	// 7. Fetch source status from PayMongo API
	sourceReq, err := http.NewRequest("GET", fmt.Sprintf("https://api.paymongo.com/v1/sources/%s", sourceID), nil)
	if err != nil {
		log.Printf("❌ Failed to create source check request: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal error"})
	}
	sourceReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.SecretKey+":")))

	resp, err := http.DefaultClient.Do(sourceReq)
	if err != nil {
		log.Printf("❌ Failed to call PayMongo API: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "PayMongo API error"})
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Failed to read response body: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Read error"})
	}
	log.Printf("✅ Source response: status=%d, body=%s", resp.StatusCode, string(respBody))

	var sourceResp SourceResponse
	if err := json.Unmarshal(respBody, &sourceResp); err != nil {
		log.Printf("❌ Failed to parse source response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Parse error"})
	}

	// 8. Determine action based on source status
	var paymentID string
	switch sourceResp.Data.Attributes.Status {
	case "paid":
		log.Printf("✅ Source %s already paid, retrieving payment ID", sourceID)
		paymentID, err = s.getPaymentIDForSource(sourceID)
		if err != nil {
			log.Printf("❌ Failed to get payment ID: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to retrieve payment",
				"details": err.Error(),
			})
		}
	case "chargeable":
		log.Printf("💳 Source %s is chargeable, creating payment", sourceID)
		paymentID, err = s.createPayment(sourceID, amount)
		if err != nil {
			log.Printf("❌ Failed to create payment: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Payment creation failed",
				"details": err.Error(),
			})
		}
	default:
		log.Printf("⚠️ Source %s has invalid status: %s", sourceID, sourceResp.Data.Attributes.Status)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid source status",
			"details": sourceResp.Data.Attributes.Status,
		})
	}

	// 9. Update transaction in database
	if err := s.DB.Model(&model.Transaction{}).
		Where("pay_mongo_source_id = ?", sourceID).
		Updates(map[string]interface{}{
			"pay_mongo_payment_id": paymentID,
			"status":               "paid",
			"updated_at":           time.Now(),
		}).Error; err != nil {
		log.Printf("❌ Failed to update transaction: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB update failed"})
	}

	log.Printf("✅ Transaction updated successfully: source_id=%s, payment_id=%s", sourceID, paymentID)
	return c.SendStatus(fiber.StatusOK)
}

// Helper function to create a payment
func (s *PayMongoService) createPayment(sourceID string, amount int) (string, error) {
	paymentReq := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"amount":      amount,
				"source":      map[string]string{"id": sourceID, "type": "source"},
				"currency":    "PHP",
				"description": "Purchase with 0.2% interest",
			},
		},
	}

	body, err := json.Marshal(paymentReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payment request: %v", err)
	}

	client := &http.Client{}
	paymongoReq, err := http.NewRequest("POST", "https://api.paymongo.com/v1/payments", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create PayMongo payment request: %v", err)
	}
	paymongoReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.SecretKey+":")))
	paymongoReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(paymongoReq)
	if err != nil {
		return "", fmt.Errorf("PayMongo payment API error: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read PayMongo payment response: %v", err)
	}
	log.Printf("PayMongo payment response: status=%d, body=%s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("PayMongo payment creation failed: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var paymentResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &paymentResp); err != nil {
		return "", fmt.Errorf("failed to parse payment response: %v", err)
	}

	return paymentResp.Data.ID, nil
}

// Helper function to retrieve payment ID for a paid source
func (s *PayMongoService) getPaymentIDForSource(sourceID string) (string, error) {
	// Query payments associated with the source
	client := &http.Client{}
	paymentsReq, err := http.NewRequest("GET", "https://api.paymongo.com/v1/payments?source_id="+sourceID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create payments request: %v", err)
	}
	paymentsReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.SecretKey+":")))

	resp, err := client.Do(paymentsReq)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve payments: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read payments response: %v", err)
	}
	log.Printf("Payments response: status=%d, body=%s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to retrieve payments: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var paymentsResp struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				Source struct {
					ID string `json:"id"`
				} `json:"source"`
				Status string `json:"status"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &paymentsResp); err != nil {
		return "", fmt.Errorf("failed to parse payments response: %v", err)
	}

	for _, payment := range paymentsResp.Data {
		if payment.Attributes.Source.ID == sourceID && payment.Attributes.Status == "paid" {
			return payment.ID, nil
		}
	}

	return "", fmt.Errorf("no paid payment found for source %s", sourceID)
}

func (s *PayMongoService) GetTransaction(c *fiber.Ctx) error {
    sourceID := c.Params("source_id")
    if sourceID == "" {
        log.Printf("Missing source_id parameter")
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Source ID required"})
    }

    var txn model.Transaction
    if err := s.DB.Where("pay_mongo_source_id = ?", sourceID).First(&txn).Error; err != nil {
        log.Printf("Transaction not found for source %s: %v", sourceID, err)
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Transaction not found"})
    }

    return c.JSON(fiber.Map{
        "user_id":             txn.UserID,
        "base_amount":         txn.BaseAmount,
        "interest_amount":     txn.InterestAmount,
        "total_amount":        txn.TotalAmount,
        "pay_mongo_source_id": txn.PayMongoSourceID,
        "pay_mongo_payment_id": txn.PayMongoPaymentID,
        "status":              txn.Status,
        "created_at":          txn.CreatedAt.Format(time.RFC3339),
        "updated_at":          txn.UpdatedAt.Format(time.RFC3339),
    })
}
// Add to PayMongoService methods
func (s *PayMongoService) GetTransactions(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		log.Printf("Missing user_id parameter")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User ID required"})
	}

	var transactions []model.Transaction
	if err := s.DB.Where("user_id = ?", userID).Find(&transactions).Error; err != nil {
		log.Printf("Failed to fetch transactions: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	return c.JSON(transactions)
}

func (s *PayMongoService) HandleSuccessRedirect(c *fiber.Ctx) error {
    // Set content type first
    c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
    
    rawQuery := c.Request().URI().QueryArgs().String()
    log.Printf("Success redirect received: raw_query=%s", rawQuery)
    
    sourceID := c.Query("id")
    log.Printf("Parsed source_id=%s", sourceID)
    
    if sourceID == "" {
        log.Printf("Missing source ID, attempting to find latest transaction")
        var txn model.Transaction
        if err := s.DB.Where("status = ?", "pending").Order("created_at desc").First(&txn).Error; err == nil {
            sourceID = txn.PayMongoSourceID
            log.Printf("Found latest transaction source_id=%s", sourceID)
        }
        
        if sourceID == "" {
            log.Printf("No valid source ID found")
            return c.Status(fiber.StatusBadRequest).SendString("Missing source ID")
        }
    }
    
    // Use the app scheme for deep linking
    appRedirectURL := fmt.Sprintf("rentxpert://paymentsuccess?id=%s", url.QueryEscape(sourceID))
    webRedirectURL := fmt.Sprintf("https://bd1a-103-72-190-240.ngrok-free.app/success?id=%s", url.QueryEscape(sourceID))
    
    log.Printf("Redirecting to app: %s or web: %s", appRedirectURL, webRedirectURL)
    
    html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="refresh" content="0;url=%s">
    <script>
        // Try deep link first, fallback to web
        window.location.href = "%s";
        setTimeout(function() {
            window.location.href = "%s";
        }, 500);
    </script>
</head>
<body>
    <p>Redirecting to app...</p>
    <a href="%s">Open in App</a> | 
    <a href="%s">Continue on Web</a>
</body>
</html>`, appRedirectURL, appRedirectURL, webRedirectURL, appRedirectURL, webRedirectURL)
    
    return c.SendString(html)
}

func (s *PayMongoService) HandleFailedRedirect(c *fiber.Ctx) error {
    c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
    
    rawQuery := c.Request().URI().QueryArgs().String()
    log.Printf("Failed redirect received: raw_query=%s", rawQuery)
    
    sourceID := c.Query("id")
    log.Printf("Parsed source_id=%s", sourceID)
    
    if sourceID == "" {
        log.Printf("Missing source ID, attempting to find latest failed transaction")
        var txn model.Transaction
        if err := s.DB.Where("status = ?", "failed").Order("created_at desc").First(&txn).Error; err == nil {
            sourceID = txn.PayMongoSourceID
            log.Printf("Found latest failed transaction source_id=%s", sourceID)
        }
    }
    
    // Use the same pattern as success redirect
    appRedirectURL := fmt.Sprintf("rentxpert://paymentfailed?id=%s", url.QueryEscape(sourceID))
    webRedirectURL := fmt.Sprintf("https://bd1a-103-72-190-240.ngrok-free.app/failed?id=%s", url.QueryEscape(sourceID))
    
    log.Printf("Redirecting to app: %s or web: %s", appRedirectURL, webRedirectURL)
    
    html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="refresh" content="0;url=%s">
    <script>
        // Try deep link first, fallback to web
        window.location.href = "%s";
        setTimeout(function() {
            window.location.href = "%s";
        }, 500);
    </script>
</head>
<body>
    <p>Redirecting to app...</p>
    <a href="%s">Open in App</a> | 
    <a href="%s">Continue on Web</a>
</body>
</html>`, appRedirectURL, appRedirectURL, webRedirectURL, appRedirectURL, webRedirectURL)
    
    return c.SendString(html)
}

// func isValidSignature(secret, payload, receivedSignature string) bool {
//     log.Printf("🔍 Validating signature: secret=%s, payload=%s, received=%s", secret, payload, receivedSignature)
//     parts := strings.Split(receivedSignature, ",")
//     var timestamp, signature string
//     for _, part := range parts {
//         if strings.HasPrefix(part, "t=") {
//             timestamp = strings.TrimPrefix(part, "t=")
//         } else if strings.HasPrefix(part, "te=") {
//             signature = strings.TrimPrefix(part, "te=")
//         }
//     }
//     if timestamp == "" || signature == "" {
//         log.Printf("❌ Invalid signature format: %s", receivedSignature)
//         return false
//     }
//     mac := hmac.New(sha256.New, []byte(secret))
//     mac.Write([]byte(timestamp + "." + payload))
//     expected := fmt.Sprintf("%x", mac.Sum(nil)) // Remove sha256= prefix
//     log.Printf("🔍 Computed signature: %s, received signature: %s", expected, signature)
//     return hmac.Equal([]byte(expected), []byte(signature))
// }