package handlers

import (
	"net/http"
	"strconv"

	"backend/pkg/database"
	"backend/pkg/models"
	"backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related requests
type UserHandler struct {
	// In a real application, you would inject dependencies like database connection here
}

// NewUserHandler creates a new UserHandler
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// GetUsers handles GET /users
func (h *UserHandler) GetUsers(c *gin.Context) {
	var users []models.User
	if err := database.GetDB().Find(&users).Error; err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve users", err.Error())
		return
	}

	utils.SuccessResponse(c, "Users retrieved successfully", users)
}

// GetUser handles GET /users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", err.Error())
		return
	}

	var user models.User
	if err := database.GetDB().First(&user, uint(id)).Error; err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	utils.SuccessResponse(c, "User retrieved successfully", user)
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := utils.ParseJSON(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err.Error())
		return
	}

	// Create user in database
	user := models.User{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := database.GetDB().Create(&user).Error; err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create user", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User created successfully",
		"data":    user,
	})
}

// ProductHandler handles product-related requests
type ProductHandler struct {
	// In a real application, you would inject dependencies here
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

// GetProducts handles GET /products
func (h *ProductHandler) GetProducts(c *gin.Context) {
	var products []models.Product
	if err := database.GetDB().Find(&products).Error; err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve products", err.Error())
		return
	}

	utils.SuccessResponse(c, "Products retrieved successfully", products)
}

// GetProduct handles GET /products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err.Error())
		return
	}

	var product models.Product
	if err := database.GetDB().First(&product, uint(id)).Error; err != nil {
		utils.NotFoundResponse(c, "Product not found")
		return
	}

	utils.SuccessResponse(c, "Product retrieved successfully", product)
}

// CreateProduct handles POST /products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := utils.ParseJSON(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err.Error())
		return
	}

	// Create product in database
	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
	}

	if err := database.GetDB().Create(&product).Error; err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create product", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Product created successfully",
		"data":    product,
	})
}
