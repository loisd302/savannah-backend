package v1

import (
	"net/http"
	"time"

	"backend/internal/repositories"
	"backend/pkg/models"
	"backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerHandler struct {
	customerRepo *repositories.CustomerRepository
}

func NewCustomerHandler(customerRepo *repositories.CustomerRepository) *CustomerHandler {
	return &CustomerHandler{
		customerRepo: customerRepo,
	}
}

// CreateCustomer handles POST /v1/customers
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req models.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err.Error())
		return
	}

	// Check if customer code already exists
	exists, err := h.customerRepo.Exists(req.Code)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to check customer existence", err.Error())
		return
	}
	if exists {
		utils.BadRequestResponse(c, "Customer code already exists", map[string]string{"code": "already taken"})
		return
	}

	// Create customer
	customer := &models.Customer{
		Code:      req.Code,
		Name:      req.Name,
		Phone:     req.Phone,
		Email:     req.Email,
		Version:   1,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.customerRepo.Create(customer); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create customer", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Customer created successfully",
		"data":    customer,
	})
}

// GetCustomer handles GET /v1/customers/:id
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid customer ID", err.Error())
		return
	}

	customer, err := h.customerRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFoundResponse(c, "Customer not found")
		} else {
			utils.InternalServerErrorResponse(c, "Failed to retrieve customer", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, "Customer retrieved successfully", customer)
}

// ListCustomers handles GET /v1/customers with query parameters
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	var query models.ListCustomersQuery
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

	customers, total, err := h.customerRepo.List(&query)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve customers", err.Error())
		return
	}

	response := gin.H{
		"customers": customers,
		"pagination": gin.H{
			"total":  total,
			"limit":  query.Limit,
			"offset": query.Offset,
		},
	}

	utils.SuccessResponse(c, "Customers retrieved successfully", response)
}

// UpdateCustomer handles PUT /v1/customers/:id
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid customer ID", err.Error())
		return
	}

	var req models.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err.Error())
		return
	}

	// Get existing customer
	customer, err := h.customerRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFoundResponse(c, "Customer not found")
		} else {
			utils.InternalServerErrorResponse(c, "Failed to retrieve customer", err.Error())
		}
		return
	}

	// Update fields if provided
	if req.Name != "" {
		customer.Name = req.Name
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}
	if req.Email != "" {
		customer.Email = req.Email
	}
	customer.UpdatedAt = time.Now()

	if err := h.customerRepo.Update(customer); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update customer", err.Error())
		return
	}

	utils.SuccessResponse(c, "Customer updated successfully", customer)
}

// DeleteCustomer handles DELETE /v1/customers/:id
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid customer ID", err.Error())
		return
	}

	// Check if customer exists
	_, err = h.customerRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFoundResponse(c, "Customer not found")
		} else {
			utils.InternalServerErrorResponse(c, "Failed to retrieve customer", err.Error())
		}
		return
	}

	if err := h.customerRepo.Delete(id); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete customer", err.Error())
		return
	}

	c.JSON(http.StatusNoContent, nil)
}