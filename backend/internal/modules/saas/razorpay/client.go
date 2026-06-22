package razorpay

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

// Config holds Razorpay API credentials.
type Config struct {
	KeyID         string
	KeySecret     string
	WebhookSecret string
	BaseURL       string
}

// Client is the Razorpay API client.
type Client struct {
	cfg    Config
	client *http.Client
}

func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.razorpay.com/v1"
	}
	return &Client{
		cfg: cfg,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// KeyID returns the Razorpay key ID for frontend checkout.
func (c *Client) KeyID() string {
	return c.cfg.KeyID
}

// CreateCustomer creates a Razorpay customer.
type Customer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"contact"`
}

func (c *Client) CreateCustomer(ctx context.Context, name, email, phone string) (*Customer, error) {
	payload := map[string]string{
		"name":    name,
		"email":   email,
		"contact": phone,
	}

	var customer Customer
	err := c.doRequest(ctx, "POST", "/customers", payload, &customer)
	if err != nil {
		return nil, fmt.Errorf("create customer: %w", err)
	}
	return &customer, nil
}

// CreateSubscription creates a Razorpay subscription.
type Subscription struct {
	ID         string `json:"id"`
	PlanID     string `json:"plan_id"`
	CustomerID string `json:"customer_id"`
	Status     string `json:"status"`
	ShortURL   string `json:"short_url"`
	CurrentEnd int64  `json:"current_end"`
}

type CreateSubscriptionInput struct {
	PlanID     string `json:"plan_id"`
	CustomerID string `json:"customer_id"`
	TotalCount int    `json:"total_count"`
	Quantity   int    `json:"quantity"`
}

func (c *Client) CreateSubscription(ctx context.Context, input CreateSubscriptionInput) (*Subscription, error) {
	var sub Subscription
	err := c.doRequest(ctx, "POST", "/subscriptions", input, &sub)
	if err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}
	return &sub, nil
}

// CancelSubscription cancels a Razorpay subscription.
func (c *Client) CancelSubscription(ctx context.Context, subscriptionID string, cancelAtEnd bool) error {
	payload := map[string]bool{"cancel_at_cycle_end": cancelAtEnd}
	return c.doRequest(ctx, "POST", "/subscriptions/"+subscriptionID+"/cancel", payload, nil)
}

// CreateOrder creates a Razorpay order for one-time payments.
type Order struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"` // in paise
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

func (c *Client) CreateOrder(ctx context.Context, amountPaise int64, currency, receipt string) (*Order, error) {
	payload := map[string]interface{}{
		"amount":   amountPaise,
		"currency": currency,
		"receipt":  receipt,
	}

	var order Order
	err := c.doRequest(ctx, "POST", "/orders", payload, &order)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	return &order, nil
}

// VerifyPaymentSignature verifies Razorpay webhook/payment signatures.
func (c *Client) VerifyPaymentSignature(orderID, paymentID, signature string) bool {
	data := orderID + "|" + paymentID
	return c.verifySignature(data, signature, c.cfg.KeySecret)
}

// VerifyWebhookSignature verifies webhook event signatures.
func (c *Client) VerifyWebhookSignature(body []byte, signature string) bool {
	return c.verifySignature(string(body), signature, c.cfg.WebhookSecret)
}

func (c *Client) verifySignature(data, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// GetSubscription fetches subscription details.
func (c *Client) GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	var sub Subscription
	err := c.doRequest(ctx, "GET", "/subscriptions/"+subscriptionID, nil, &sub)
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	return &sub, nil
}

// FetchPayment fetches payment details.
type Payment struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
	Method   string `json:"method"`
	OrderID  string `json:"order_id"`
	Email    string `json:"email"`
}

func (c *Client) FetchPayment(ctx context.Context, paymentID string) (*Payment, error) {
	var payment Payment
	err := c.doRequest(ctx, "GET", "/payments/"+paymentID, nil, &payment)
	if err != nil {
		return nil, fmt.Errorf("fetch payment: %w", err)
	}
	return &payment, nil
}

// CreateRefund issues a refund.
func (c *Client) CreateRefund(ctx context.Context, paymentID string, amountPaise int64) error {
	payload := map[string]interface{}{"amount": amountPaise}
	return c.doRequest(ctx, "POST", "/payments/"+paymentID+"/refund", payload, nil)
}

func (c *Client) doRequest(ctx context.Context, method, path string, payload interface{}, result interface{}) error {
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}
	}

	// Retry with exponential backoff for transient failures (5xx, network errors)
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		var reqBody io.Reader
		if body != nil {
			reqBody = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, reqBody)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.SetBasicAuth(c.cfg.KeyID, c.cfg.KeySecret)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("do request: %w", err)
			continue // Retry on network error
		}

		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		// Don't retry client errors (4xx) — only server errors (5xx) and network issues
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("razorpay API error (status %d): %s", resp.StatusCode, string(respBody))
			continue // Retry on server error
		}

		if resp.StatusCode >= 400 {
			return fmt.Errorf("razorpay API error (status %d): %s", resp.StatusCode, string(respBody))
		}

		if result != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("unmarshal response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("razorpay request failed after %d retries: %w", maxRetries, lastErr)
}
