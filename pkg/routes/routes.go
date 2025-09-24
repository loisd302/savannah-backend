package routes

import (
	"backend/internal/api/v1"
	"backend/internal/auth"
	"backend/internal/repositories"
	"backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(router *gin.Engine, db *gorm.DB, oidcProvider *auth.OIDCProvider, smsService *services.SMSService) {
	// Initialize repositories
	customerRepo := repositories.NewCustomerRepository(db)
	orderRepo := repositories.NewOrderRepository(db)

	// Initialize handlers
	customerHandler := v1.NewCustomerHandler(customerRepo)
	orderHandler := v1.NewOrderHandler(orderRepo, customerRepo, smsService)

	// API v1 routes
	api := router.Group("/api/v1")
	{
		// Customer routes
		customers := api.Group("/customers")
		{
			// Public routes (with basic auth)
			customers.POST("/", oidcProvider.RequireScopes("customers:write"), customerHandler.CreateCustomer)
			customers.GET("/", oidcProvider.RequireScopes("customers:read"), customerHandler.ListCustomers)
			customers.GET("/:id", oidcProvider.RequireScopes("customers:read"), customerHandler.GetCustomer)
			customers.PUT("/:id", oidcProvider.RequireScopes("customers:write"), customerHandler.UpdateCustomer)
			customers.DELETE("/:id", oidcProvider.RequireRoles("admin"), customerHandler.DeleteCustomer)
			
			// Customer orders
			customers.GET("/:id/orders", oidcProvider.RequireScopes("orders:read"), orderHandler.GetCustomerOrders)
		}

		// Order routes
		orders := api.Group("/orders")
		{
			orders.POST("/", oidcProvider.RequireScopes("orders:write"), orderHandler.CreateOrder)
			orders.GET("/", oidcProvider.RequireScopes("orders:read"), orderHandler.ListOrders)
			orders.GET("/:id", oidcProvider.RequireScopes("orders:read"), orderHandler.GetOrder)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(oidcProvider.RequireRoles("admin"))
		{
			admin.GET("/stats", func(c *gin.Context) {
				// Get SMS job stats
				smsStats, _ := smsService.GetSMSJobStats(c.Request.Context())
				
				c.JSON(200, gin.H{
					"message": "Admin statistics",
					"stats": gin.H{
						"sms_jobs": smsStats,
					},
				})
			})
		}
	}

	// API documentation route
	router.GET("/docs", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Savannah Backend API Documentation",
			"version": "1.0.0",
			"endpoints": gin.H{
				"health":            "GET /health",
				"customers":         "GET|POST /api/v1/customers (auth: customers:read|write)",
				"customer_by_id":    "GET|PUT|DELETE /api/v1/customers/:id",
				"customer_orders":   "GET /api/v1/customers/:id/orders",
				"orders":            "GET|POST /api/v1/orders (auth: orders:read|write)",
				"order_by_id":       "GET /api/v1/orders/:id",
				"admin_stats":       "GET /api/v1/admin/stats (role: admin)",
			},
			"authentication": gin.H{
				"type":   "OIDC Bearer Token",
				"header": "Authorization: Bearer <access_token>",
				"scopes": []string{"customers:read", "customers:write", "orders:read", "orders:write"},
				"roles":  []string{"admin"},
			},
		})
	})
}
