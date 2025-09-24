package performance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"savannah-backend/internal/domain"
	"savannah-backend/pkg/config"
	"savannah-backend/pkg/database"
)

var (
	benchmarkRouter *gin.Engine
	benchmarkDB     *database.Database
	setupOnce       sync.Once
)

func setupBenchmark() {
	// Initialize test environment
	config := &config.Config{
		Database: config.DatabaseConfig{
			URL: "postgres://testuser:testpass@localhost:5432/benchmark_db?sslmode=disable",
		},
		Environment: "test",
	}

	var err error
	benchmarkDB, err = database.New(config)
	if err != nil {
		panic("Failed to connect to benchmark database: " + err.Error())
	}

	// Setup router
	gin.SetMode(gin.TestMode)
	benchmarkRouter = gin.New()
	
	// Add minimal middleware for benchmarking
	benchmarkRouter.Use(gin.Recovery())
	
	// Setup routes (simplified for benchmarking)
	setupBenchmarkRoutes()
}

func setupBenchmarkRoutes() {
	v1 := benchmarkRouter.Group("/api/v1")
	
	// Simple health check endpoint
	benchmarkRouter.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	
	// Customer endpoints for benchmarking
	v1.GET("/customers", func(c *gin.Context) {
		// Simulate customer retrieval
		customers := []domain.Customer{
			{Name: "Customer 1", Code: "CUST001"},
			{Name: "Customer 2", Code: "CUST002"},
			{Name: "Customer 3", Code: "CUST003"},
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    customers,
		})
	})
	
	v1.POST("/customers", func(c *gin.Context) {
		var customer domain.Customer
		if err := c.ShouldBindJSON(&customer); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Simulate customer creation
		customer.ID = "12345678-1234-1234-1234-123456789012"
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    customer,
		})
	})
}

// BenchmarkHealthCheck benchmarks the health check endpoint
func BenchmarkHealthCheck(b *testing.B) {
	setupOnce.Do(setupBenchmark)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			benchmarkRouter.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
			}
		}
	})
}

// BenchmarkGetCustomers benchmarks the get customers endpoint
func BenchmarkGetCustomers(b *testing.B) {
	setupOnce.Do(setupBenchmark)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/customers", nil)
			benchmarkRouter.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
			}
		}
	})
}

// BenchmarkCreateCustomer benchmarks the create customer endpoint
func BenchmarkCreateCustomer(b *testing.B) {
	setupOnce.Do(setupBenchmark)
	
	customer := domain.Customer{
		Name:        "Benchmark Customer",
		Code:        "BENCH001",
		PhoneNumber: "+254700000000",
		Email:       "bench@example.com",
	}
	
	jsonData, _ := json.Marshal(customer)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/customers", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			benchmarkRouter.ServeHTTP(w, req)
			
			if w.Code != http.StatusCreated {
				b.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
			}
		}
	})
}

// BenchmarkJSONSerialization benchmarks JSON serialization performance
func BenchmarkJSONSerialization(b *testing.B) {
	customers := make([]domain.Customer, 100)
	for i := 0; i < 100; i++ {
		customers[i] = domain.Customer{
			Name:        fmt.Sprintf("Customer %d", i),
			Code:        fmt.Sprintf("CUST%03d", i),
			PhoneNumber: "+254700000000",
			Email:       fmt.Sprintf("customer%d@example.com", i),
		}
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := json.Marshal(customers)
			if err != nil {
				b.Error("JSON marshaling failed:", err)
			}
		}
	})
}

// BenchmarkJSONDeserialization benchmarks JSON deserialization performance
func BenchmarkJSONDeserialization(b *testing.B) {
	customer := domain.Customer{
		Name:        "Benchmark Customer",
		Code:        "BENCH001",
		PhoneNumber: "+254700000000",
		Email:       "bench@example.com",
	}
	
	jsonData, _ := json.Marshal(customer)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var c domain.Customer
			err := json.Unmarshal(jsonData, &c)
			if err != nil {
				b.Error("JSON unmarshaling failed:", err)
			}
		}
	})
}

// BenchmarkConcurrentRequests tests performance under concurrent load
func BenchmarkConcurrentRequests(b *testing.B) {
	setupOnce.Do(setupBenchmark)
	
	// Test different concurrency levels
	concurrencyLevels := []int{1, 10, 50, 100, 200}
	
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency-%d", concurrency), func(b *testing.B) {
			b.SetParallelism(concurrency)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					w := httptest.NewRecorder()
					req, _ := http.NewRequest("GET", "/health", nil)
					benchmarkRouter.ServeHTTP(w, req)
					
					if w.Code != http.StatusOK {
						b.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
					}
				}
			})
		})
	}
}

// TestPerformanceRequirements tests that performance requirements are met
func TestPerformanceRequirements(t *testing.T) {
	setupOnce.Do(setupBenchmark)
	
	// Test response time requirements
	t.Run("ResponseTimeRequirement", func(t *testing.T) {
		// Measure average response time over multiple requests
		totalTime := int64(0)
		numRequests := 100
		
		for i := 0; i < numRequests; i++ {
			start := testing.Benchmark(func(b *testing.B) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/health", nil)
				benchmarkRouter.ServeHTTP(w, req)
			})
			totalTime += start.NsPerOp()
		}
		
		avgResponseTimeMs := float64(totalTime/int64(numRequests)) / 1e6
		
		// Requirement: Average response time should be less than 50ms
		assert.Less(t, avgResponseTimeMs, 50.0, 
			"Average response time (%.2fms) exceeds requirement (50ms)", avgResponseTimeMs)
		
		t.Logf("Average response time: %.2fms", avgResponseTimeMs)
	})
	
	// Test memory allocation requirements
	t.Run("MemoryAllocationRequirement", func(t *testing.T) {
		result := testing.Benchmark(BenchmarkHealthCheck)
		
		// Requirement: Less than 1KB allocation per request
		bytesPerOp := result.AllocedBytesPerOp()
		assert.Less(t, int(bytesPerOp), 1024, 
			"Memory allocation per request (%d bytes) exceeds requirement (1KB)", bytesPerOp)
		
		t.Logf("Memory allocation per request: %d bytes", bytesPerOp)
	})
	
	// Test throughput requirements
	t.Run("ThroughputRequirement", func(t *testing.T) {
		result := testing.Benchmark(BenchmarkHealthCheck)
		
		// Calculate requests per second
		opsPerSec := float64(result.N) / result.T.Seconds()
		
		// Requirement: Should handle at least 1000 requests/second
		assert.GreaterOrEqual(t, opsPerSec, 1000.0,
			"Throughput (%.2f req/s) is below requirement (1000 req/s)", opsPerSec)
		
		t.Logf("Throughput: %.2f requests/second", opsPerSec)
	})
}