package handlers

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"smartcv-backend/internal/database"
	"time"

	"github.com/gin-gonic/gin"
)

// MidtransConfig holds Midtrans configuration
type MidtransConfig struct {
	ServerKey  string
	ClientKey  string
	IsSandbox  bool
}

var midtransConfig MidtransConfig

// InitMidtrans initializes Midtrans configuration
func InitMidtrans() {
	midtransConfig = MidtransConfig{
		ServerKey:  os.Getenv("MIDTRANS_SERVER_KEY"),
		ClientKey:  os.Getenv("MIDTRANS_CLIENT_KEY"),
		IsSandbox:  os.Getenv("MIDTRANS_IS_SANDBOX") == "true",
	}
}

// CreditPackage represents a credit purchase package
type CreditPackage struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Credits     int     `json:"credits"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

// GetCreditPackages returns available credit packages
func GetCreditPackages(c *gin.Context) {
	packages := []CreditPackage{
		{
			ID:          "starter",
			Name:        "Starter Pack",
			Credits:     5,
			Price:       4.99,
			Description: "Perfect for trying out",
		},
		{
			ID:          "basic",
			Name:        "Basic Pack",
			Credits:     15,
			Price:       12.99,
			Description: "Best value for regular use",
		},
		{
			ID:          "pro",
			Name:        "Pro Pack",
			Credits:     50,
			Price:       39.99,
			Description: "For power users",
		},
	}
	c.JSON(http.StatusOK, packages)
}

// PurchaseCreditsRequest represents a credit purchase request
type PurchaseCreditsRequest struct {
	PackageID string `json:"package_id" binding:"required"`
}

// PurchaseCreditsMidtrans initiates a Midtrans transaction
func PurchaseCreditsMidtrans(c *gin.Context) {
	userID := getUserID(c)

	var req PurchaseCreditsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get package details
	var pkg CreditPackage
	switch req.PackageID {
	case "starter":
		pkg = CreditPackage{ID: "starter", Name: "Starter Pack", Credits: 5, Price: 4.99}
	case "basic":
		pkg = CreditPackage{ID: "basic", Name: "Basic Pack", Credits: 15, Price: 12.99}
	case "pro":
		pkg = CreditPackage{ID: "pro", Name: "Pro Pack", Credits: 50, Price: 39.99}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package ID"})
		return
	}

	// Generate order ID
	orderID := fmt.Sprintf("ORD-%d-%d", userID, getCurrentTimestamp())

	// Create Midtrans transaction request
	snapURL := "https://app.sandbox.midtrans.com/snap/v1/transactions"
	if !midtransConfig.IsSandbox {
		snapURL = "https://app.midtrans.com/snap/v1/transactions"
	}

	// Prepare request body
	reqBody := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":     orderID,
			"gross_amount": pkg.Price,
		},
		"customer_details": map[string]interface{}{
			"user_id": userID,
		},
		"item_details": []map[string]interface{}{
			{
				"id":       pkg.ID,
				"price":    pkg.Price,
				"quantity": 1,
				"name":     pkg.Name,
			},
		},
	}

	reqJSON, _ := json.Marshal(reqBody)

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", snapURL, bytes.NewReader(reqJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Set headers
	authKey := base64Encode(midtransConfig.ServerKey + ":")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+authKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to payment gateway"})
		return
	}
	defer resp.Body.Close()

	// Parse response
	var snapResp struct {
		Token       string `json:"token"`
		RedirectURL string `json:"redirect_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&snapResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	// Save transaction to database
	_, err = database.DB.Exec(`
		INSERT INTO credit_transactions (user_id, amount, transaction_type, description, midtrans_order_id)
		VALUES ($1, $2, 'purchase', $3, $4)
	`, userID, pkg.Credits, pkg.Name, orderID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":        snapResp.Token,
		"redirect_url": snapResp.RedirectURL,
		"order_id":     orderID,
	})
}

// MidtransWebhook handles Midtrans webhook notifications
func MidtransWebhook(c *gin.Context) {
	var notification struct {
		OrderID            string `json:"order_id"`
		StatusCode         string `json:"status_code"`
		GrossAmount        string `json:"gross_amount"`
		TransactionStatus  string `json:"transaction_status"`
		TransactionID      string `json:"transaction_id"`
		FraudStatus        string `json:"fraud_status"`
		PaymentType        string `json:"payment_type"`
		SignatureKey       string `json:"signature_key"`
	}

	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification"})
		return
	}

	// Verify signature
	expectedSig := generateSignatureKey(notification.OrderID, notification.StatusCode, notification.GrossAmount, midtransConfig.ServerKey)
	if notification.SignatureKey != expectedSig {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Get transaction from database
	var userID int
	var amount int
	err := database.DB.QueryRow(`
		SELECT user_id, amount FROM credit_transactions WHERE midtrans_order_id = $1
	`, notification.OrderID).Scan(&userID, &amount)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Update transaction status
	_, err = database.DB.Exec(`
		UPDATE credit_transactions 
		SET midtrans_transaction_id = $1, description = description || ' | ' || $2
		WHERE midtrans_order_id = $3
	`, notification.TransactionID, notification.TransactionStatus, notification.OrderID)

	// Add credits if payment successful
	if notification.TransactionStatus == "capture" || notification.TransactionStatus == "settlement" {
		if notification.FraudStatus == "accept" || notification.FraudStatus == "" {
			// Add credits to user
			_, err = database.DB.Exec(`
				UPDATE users SET credits = credits + $1 WHERE id = $2
			`, amount, userID)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update credits"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Helper functions
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func generateSignatureKey(orderID, statusCode, grossAmount, serverKey string) string {
	data := orderID + statusCode + grossAmount + serverKey
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
