package models

import (
	"time"

	"github.com/google/uuid"
)

// Response represents a standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// Customer represents a customer in the system
type Customer struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Code      string    `json:"code" gorm:"type:varchar(32);uniqueIndex;not null"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null"`
	Phone     string    `json:"phone" gorm:"type:varchar(20);index"`
	Email     string    `json:"email" gorm:"type:varchar(255)"`
	Version   int       `json:"version" gorm:"default:1"`
	IsActive  bool      `json:"is_active" gorm:"default:true;index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relations
	Orders []Order `json:"orders,omitempty" gorm:"foreignKey:CustomerID"`
}

// Order represents an order in the system
type Order struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerID uuid.UUID  `json:"customer_id" gorm:"type:uuid;not null;index"`
	Item       string     `json:"item" gorm:"type:varchar(255);not null"`
	Amount     float64    `json:"amount" gorm:"type:numeric(12,2);not null"`
	OrderedAt  time.Time  `json:"ordered_at" gorm:"index"`
	Status     string     `json:"status" gorm:"type:varchar(20);default:'pending';index"`
	SMSSentAt  *time.Time `json:"sms_sent_at,omitempty"`
	Version    int        `json:"version" gorm:"default:1"`
	IsActive   bool       `json:"is_active" gorm:"default:true;index"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relations
	Customer Customer `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
}

// History tables for audit trail
type CustomerHistory struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;not null"`
	Code      string    `json:"code" gorm:"type:varchar(32);not null"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null"`
	Phone     string    `json:"phone" gorm:"type:varchar(20)"`
	Email     string    `json:"email" gorm:"type:varchar(255)"`
	Version   int       `json:"version"`
	ValidFrom time.Time `json:"valid_from"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
	ChangedBy string    `json:"changed_by" gorm:"type:varchar(100)"`
}

type OrderHistory struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;not null"`
	CustomerID uuid.UUID  `json:"customer_id" gorm:"type:uuid;not null"`
	Item       string     `json:"item" gorm:"type:varchar(255);not null"`
	Amount     float64    `json:"amount" gorm:"type:numeric(12,2);not null"`
	OrderedAt  time.Time  `json:"ordered_at"`
	Status     string     `json:"status" gorm:"type:varchar(20)"`
	SMSSentAt  *time.Time `json:"sms_sent_at,omitempty"`
	Version    int        `json:"version"`
	ValidFrom  time.Time  `json:"valid_from"`
	ValidTo    *time.Time `json:"valid_to,omitempty"`
	ChangedBy  string     `json:"changed_by" gorm:"type:varchar(100)"`
}

// Request/Response models
type CreateCustomerRequest struct {
	Code  string `json:"code" binding:"required,min=2,max=32"`
	Name  string `json:"name" binding:"required,min=2,max=255"`
	Phone string `json:"phone" binding:"required,min=10,max=20"`
	Email string `json:"email" binding:"omitempty,email"`
}

type UpdateCustomerRequest struct {
	Name  string `json:"name" binding:"omitempty,min=2,max=255"`
	Phone string `json:"phone" binding:"omitempty,min=10,max=20"`
	Email string `json:"email" binding:"omitempty,email"`
}

type CreateOrderRequest struct {
	CustomerID uuid.UUID  `json:"customer_id" binding:"required"`
	Item       string     `json:"item" binding:"required,min=2,max=255"`
	Amount     float64    `json:"amount" binding:"required,gt=0"`
	OrderedAt  *time.Time `json:"ordered_at,omitempty"`
}

type ListCustomersQuery struct {
	Code   string `form:"code"`
	Name   string `form:"name"`
	Phone  string `form:"phone"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int    `form:"offset" binding:"omitempty,min=0"`
}

type ListOrdersQuery struct {
	CustomerID uuid.UUID `form:"customer_id"`
	Status     string    `form:"status"`
	Limit      int       `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset     int       `form:"offset" binding:"omitempty,min=0"`
}
