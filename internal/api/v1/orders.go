package v1

import (
	"log"
	"net/http"
	"time"

	"backend/internal/repositories"
	"backend/internal/services"
	"backend/pkg/models"
	"backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderHandler struct {
	orderRepo    *repositories.OrderRepository
	customerRepo *repositories.CustomerRepository
	smsService   *services.SMSService
}

func NewOrderHandler(orderRepo *repositories.OrderRepository, customerRepo *repositories.CustomerRepository, smsService *services.SMSService) *OrderHandler {
	return &OrderHandler{
		orderRepo:    orderRepo,
		customerRepo: customerRepo,
		smsService:   smsService,
	}
}

// CreateOrder handles POST /v1/orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err.Error())
		return
	}

	// Verify customer exists
	customer, err := h.customerRepo.GetByID(req.CustomerID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.BadRequestResponse(c, "Customer not found", err.Error())
		} else {
			utils.InternalServerErrorResponse(c, "Failed to verify customer", err.Error())
		}
		return
	}

	// Set ordered_at if not provided
	orderedAt := time.Now()
	if req.OrderedAt != nil {
		orderedAt = *req.OrderedAt
	}

	// Create order
	order := &models.Order{
		CustomerID: req.CustomerID,
		Item:       req.Item,
		Amount:     req.Amount,
		OrderedAt:  orderedAt,
		Status:     "pending",
		Version:    1,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := h.orderRepo.Create(order); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create order", err.Error())
		return
	}

	// Load customer relationship for response
	order.Customer = *customer

	// Queue SMS job for background processing
	if err := h.smsService.QueueSMS(c.Request.Context(), order); err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to queue SMS for order %s: %v", order.ID, err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Order created successfully",
		"data":    order,
	})
}

// GetOrder handles GET /v1/orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid order ID", err.Error())
		return
	}

	order, err := h.orderRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFoundResponse(c, "Order not found")
		} else {
			utils.InternalServerErrorResponse(c, "Failed to retrieve order", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, "Order retrieved successfully", order)
}

// ListOrders handles GET /v1/orders with query parameters
func (h *OrderHandler) ListOrders(c *gin.Context) {
	var query models.ListOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequestResponse(c, "Invalid query parameters", err.Error())
		return
	}

	// Set default pagination values
	if query.Limit == 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	orders, total, err := h.orderRepo.List(&query)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve orders", err.Error())
		return
	}

	response := gin.H{
		"orders": orders,
		"pagination": gin.H{
			"total":  total,
			"limit":  query.Limit,
			"offset": query.Offset,
		},
	}

	utils.SuccessResponse(c, "Orders retrieved successfully", response)
}

// GetCustomerOrders handles GET /v1/customers/:id/orders
func (h *OrderHandler) GetCustomerOrders(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid customer ID", err.Error())
		return
	}

	// Verify customer exists
	_, err = h.customerRepo.GetByID(customerID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFoundResponse(c, "Customer not found")
		} else {
			utils.InternalServerErrorResponse(c, "Failed to verify customer", err.Error())
		}
		return
	}

	orders, err := h.orderRepo.GetByCustomerID(customerID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve customer orders", err.Error())
		return
	}

	utils.SuccessResponse(c, "Customer orders retrieved successfully", orders)
}