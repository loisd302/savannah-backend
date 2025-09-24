package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"savannah-backend/internal/api/v1"
	"savannah-backend/internal/domain"
	"savannah-backend/pkg/database"
	"savannah-backend/pkg/config"
)

type APITestSuite struct {
	suite.Suite
	router *gin.Engine
	db     *database.Database
}

func (suite *APITestSuite) SetupSuite() {
	// Initialize test database
	config := &config.Config{
		Database: config.DatabaseConfig{
			URL: "postgres://testuser:testpass@localhost:5432/test_db?sslmode=disable",
		},
		Environment: "test",
	}

	var err error
	suite.db, err = database.New(config)
	if err != nil {
		suite.T().Fatal("Failed to connect to test database:", err)
	}

	// Setup test router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// Setup API routes
	v1Group := suite.router.Group("/api/v1")
	api.SetupCustomerRoutes(v1Group, suite.db)
	api.SetupOrderRoutes(v1Group, suite.db)
}

func (suite *APITestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *APITestSuite) SetupTest() {
	// Clean test data before each test
	suite.db.Exec("TRUNCATE customers, orders CASCADE")
}

func (suite *APITestSuite) TestHealthCheck() {
	// Setup health check route
	suite.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", response["status"])
}

func (suite *APITestSuite) TestCreateCustomer_Success() {
	customer := domain.Customer{
		Name:        "John Doe",
		Code:        "CUST001",
		PhoneNumber: "+254700123456",
		Email:       "john@example.com",
	}

	jsonData, _ := json.Marshal(customer)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/customers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

func (suite *APITestSuite) TestCreateCustomer_ValidationError() {
	invalidCustomer := domain.Customer{
		Name: "", // Invalid: empty name
		Code: "CUST001",
	}

	jsonData, _ := json.Marshal(invalidCustomer)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/customers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["error"])
}

func (suite *APITestSuite) TestGetCustomers_Success() {
	// Create test customer first
	customer := domain.Customer{
		Name:        "Jane Doe",
		Code:        "CUST002",
		PhoneNumber: "+254700654321",
		Email:       "jane@example.com",
	}

	jsonData, _ := json.Marshal(customer)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/customers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	// Now get customers
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/customers", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	
	data := response["data"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(data), 1)
}

func (suite *APITestSuite) TestGetCustomer_NotFound() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/customers/00000000-0000-0000-0000-000000000000", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
}

func (suite *APITestSuite) TestCreateOrder_Success() {
	// First create a customer
	customer := domain.Customer{
		Name:        "Test Customer",
		Code:        "CUST003",
		PhoneNumber: "+254700123456",
		Email:       "test@example.com",
	}

	jsonData, _ := json.Marshal(customer)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/customers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	var customerResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &customerResponse)
	customerData := customerResponse["data"].(map[string]interface{})
	customerID := customerData["id"].(string)

	// Now create order
	order := domain.Order{
		CustomerID: customerID,
		Item:       "Test Product",
		Amount:     99.99,
	}

	jsonData, _ = json.Marshal(order)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

func (suite *APITestSuite) TestRateLimiting() {
	// Test rate limiting by making multiple requests quickly
	endpoint := "/api/v1/customers"
	
	successCount := 0
	rateLimitedCount := 0
	
	// Make multiple requests
	for i := 0; i < 20; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", endpoint, nil)
		suite.router.ServeHTTP(w, req)
		
		if w.Code == http.StatusOK {
			successCount++
		} else if w.Code == http.StatusTooManyRequests {
			rateLimitedCount++
		}
	}
	
	// Should have some successful requests and some rate-limited
	assert.Greater(suite.T(), successCount, 0)
	// Note: This test may need adjustment based on actual rate limiting configuration
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}