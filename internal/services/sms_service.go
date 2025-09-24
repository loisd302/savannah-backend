package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"backend/pkg/models"
	"github.com/google/uuid"
)

// SMSConfig holds SMS service configuration
type SMSConfig struct {
	Username    string
	APIKey      string
	Shortcode   string
	BaseURL     string
	IsSandbox   bool
	RetryLimit  int
	RetryDelay  time.Duration
}

// SMSRequest represents the request payload for Africa's Talking SMS API
type SMSRequest struct {
	Username string `json:"username"`
	To       string `json:"to"`
	Message  string `json:"message"`
	From     string `json:"from,omitempty"`
}

// SMSResponse represents the response from Africa's Talking SMS API
type SMSResponse struct {
	SMSMessageData SMSMessageData `json:"SMSMessageData"`
}

type SMSMessageData struct {
	Message    string        `json:"Message"`
	Recipients []SMSRecipient `json:"Recipients"`
}

type SMSRecipient struct {
	StatusCode   int    `json:"statusCode"`
	Number       string `json:"number"`
	Status       string `json:"status"`
	Cost         string `json:"cost"`
	MessageId    string `json:"messageId"`
	MessageParts int    `json:"messageParts"`
}

// SMSJob represents a background SMS job
type SMSJob struct {
	ID           uuid.UUID `json:"id"`
	OrderID      uuid.UUID `json:"order_id"`
	CustomerID   uuid.UUID `json:"customer_id"`
	Phone        string    `json:"phone"`
	Message      string    `json:"message"`
	Status       string    `json:"status"` // pending, sent, failed
	Attempts     int       `json:"attempts"`
	MaxAttempts  int       `json:"max_attempts"`
	LastError    string    `json:"last_error,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastAttempt  time.Time `json:"last_attempt"`
	ScheduledFor time.Time `json:"scheduled_for"`
}

// SMSService handles SMS operations
type SMSService struct {
	config     *SMSConfig
	httpClient *http.Client
	jobQueue   JobQueue
}

// JobQueue interface for job queuing
type JobQueue interface {
	Enqueue(ctx context.Context, job *SMSJob) error
	Dequeue(ctx context.Context) (*SMSJob, error)
	UpdateJob(ctx context.Context, job *SMSJob) error
	RetryJob(ctx context.Context, job *SMSJob, delay time.Duration) error
	MarkCompleted(ctx context.Context, jobID uuid.UUID) error
	MarkFailed(ctx context.Context, jobID uuid.UUID, error string) error
}

// NewSMSService creates a new SMS service
func NewSMSService(config *SMSConfig, jobQueue JobQueue) *SMSService {
	return &SMSService{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		jobQueue: jobQueue,
	}
}

// QueueSMS queues an SMS job for background processing
func (s *SMSService) QueueSMS(ctx context.Context, order *models.Order) error {
	if order.Customer.Phone == "" {
		return fmt.Errorf("customer phone number is required")
	}

	message := s.buildOrderSMSMessage(order)
	
	job := &SMSJob{
		ID:           uuid.New(),
		OrderID:      order.ID,
		CustomerID:   order.CustomerID,
		Phone:        order.Customer.Phone,
		Message:      message,
		Status:       "pending",
		Attempts:     0,
		MaxAttempts:  s.config.RetryLimit,
		CreatedAt:    time.Now(),
		ScheduledFor: time.Now(),
	}

	return s.jobQueue.Enqueue(ctx, job)
}

// ProcessSMSJobs processes pending SMS jobs
func (s *SMSService) ProcessSMSJobs(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			job, err := s.jobQueue.Dequeue(ctx)
			if err != nil {
				log.Printf("Failed to dequeue SMS job: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}
			
			if job == nil {
				time.Sleep(1 * time.Second)
				continue
			}

			if err := s.processSMSJob(ctx, job); err != nil {
				log.Printf("Failed to process SMS job %s: %v", job.ID, err)
			}
		}
	}
}

// processSMSJob processes a single SMS job
func (s *SMSService) processSMSJob(ctx context.Context, job *SMSJob) error {
	job.Attempts++
	job.LastAttempt = time.Now()

	// Send SMS
	response, err := s.sendSMS(ctx, job.Phone, job.Message)
	if err != nil {
		job.LastError = err.Error()
		
		// Check if we should retry
		if job.Attempts < job.MaxAttempts {
			delay := time.Duration(job.Attempts*job.Attempts) * s.config.RetryDelay
			log.Printf("SMS job %s failed (attempt %d/%d), retrying in %v: %v", 
				job.ID, job.Attempts, job.MaxAttempts, delay, err)
			return s.jobQueue.RetryJob(ctx, job, delay)
		}

		// Mark as failed
		job.Status = "failed"
		s.jobQueue.MarkFailed(ctx, job.ID, err.Error())
		log.Printf("SMS job %s permanently failed after %d attempts: %v", 
			job.ID, job.Attempts, err)
		return nil
	}

	// Check response status
	if len(response.SMSMessageData.Recipients) > 0 {
		recipient := response.SMSMessageData.Recipients[0]
		if recipient.StatusCode == 101 || recipient.StatusCode == 100 {
			// Success
			job.Status = "sent"
			s.jobQueue.MarkCompleted(ctx, job.ID)
			log.Printf("SMS job %s completed successfully: %s", job.ID, recipient.Status)
		} else {
			// API returned error
			errorMsg := fmt.Sprintf("SMS API error: %s (code: %d)", recipient.Status, recipient.StatusCode)
			job.LastError = errorMsg
			
			if job.Attempts < job.MaxAttempts {
				delay := time.Duration(job.Attempts*job.Attempts) * s.config.RetryDelay
				return s.jobQueue.RetryJob(ctx, job, delay)
			}
			
			job.Status = "failed"
			s.jobQueue.MarkFailed(ctx, job.ID, errorMsg)
			log.Printf("SMS job %s failed with API error: %s", job.ID, errorMsg)
		}
	}

	return s.jobQueue.UpdateJob(ctx, job)
}

// sendSMS sends an SMS using Africa's Talking API
func (s *SMSService) sendSMS(ctx context.Context, phone, message string) (*SMSResponse, error) {
	// Format phone number (ensure it starts with country code)
	if len(phone) > 0 && phone[0] != '+' && !s.isInternationalFormat(phone) {
		phone = "+254" + phone // Default to Kenya country code for sandbox
	}

	smsRequest := SMSRequest{
		Username: s.config.Username,
		To:       phone,
		Message:  message,
		From:     s.config.Shortcode,
	}

	jsonData, err := json.Marshal(smsRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SMS request: %w", err)
	}

	url := s.config.BaseURL + "/messaging"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apiKey", s.config.APIKey)

	log.Printf("Sending SMS to %s: %s", phone, message)
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("SMS API returned status %d: %s", resp.StatusCode, string(body))
	}

	var smsResponse SMSResponse
	if err := json.Unmarshal(body, &smsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SMS response: %w", err)
	}

	return &smsResponse, nil
}

// buildOrderSMSMessage builds the SMS message for an order
func (s *SMSService) buildOrderSMSMessage(order *models.Order) string {
	return fmt.Sprintf(
		"Hello %s! Your order for %s (Amount: %.2f) has been received. Order ID: %s. Thank you!",
		order.Customer.Name,
		order.Item,
		order.Amount,
		order.ID,
	)
}

// isInternationalFormat checks if phone number is in international format
func (s *SMSService) isInternationalFormat(phone string) bool {
	return len(phone) >= 10 && (phone[:3] == "254" || phone[:4] == "2547")
}

// GetSMSJobStats returns statistics about SMS jobs
func (s *SMSService) GetSMSJobStats(ctx context.Context) (map[string]int64, error) {
	// This would be implemented based on your job queue backend
	// For now, return empty stats
	return map[string]int64{
		"pending": 0,
		"sent":    0,
		"failed":  0,
	}, nil
}