package docs

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupSwaggerRoutes sets up Swagger documentation routes
func SetupSwaggerRoutes(router *gin.Engine) {
	// OpenAPI JSON endpoint (serve at standard swagger.json path)
	router.GET("/swagger.json", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, swaggerSpec)
	})
	
	// Also serve at custom path for backward compatibility
	router.GET("/api/openapi.json", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, swaggerSpec)
	})
	
	// Swagger UI endpoint with custom URL configuration
	url := ginSwagger.URL("/swagger.json") // The url pointing to API definition
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	
	// API documentation redirect
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/docs/index.html")
	})

// swaggerSpec contains the OpenAPI 3.0 specification
const swaggerSpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Savannah Backend API",
    "description": "Enterprise-grade RESTful API for customer and order management with SMS notifications",
    "version": "1.0.0",
    "contact": {
      "name": "API Support",
      "url": "https://github.com/jmukavana/savannah-backend",
      "email": "jmukavana@github.com"
    },
    "license": {
      "name": "MIT",
      "url": "https://github.com/jmukavana/savannah-backend/blob/main/LICENSE"
    }
  },
  "servers": [
    {
      "url": "http://localhost:8080",
      "description": "Local development server"
    },
    {
      "url": "https://api-dev.savannah.com",
      "description": "Development server"
    },
    {
      "url": "https://api.savannah.com",
      "description": "Production server"
    }
  ],
  "security": [
    {
      "BearerAuth": []
    }
  ],
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT",
        "description": "JWT token obtained from OIDC provider"
      }
    },
    "schemas": {
      "Customer": {
        "type": "object",
        "required": ["name", "code"],
        "properties": {
          "id": {
            "type": "string",
            "format": "uuid",
            "description": "Unique identifier for the customer",
            "example": "123e4567-e89b-12d3-a456-426614174000"
          },
          "name": {
            "type": "string",
            "description": "Customer full name",
            "example": "John Doe",
            "minLength": 2,
            "maxLength": 100
          },
          "code": {
            "type": "string",
            "description": "Unique customer code",
            "example": "CUST001",
            "pattern": "^[A-Z0-9]{4,20}$"
          },
          "phone_number": {
            "type": "string",
            "description": "Customer phone number in international format",
            "example": "+254700123456",
            "pattern": "^\\+[1-9]\\d{1,14}$"
          },
          "email": {
            "type": "string",
            "format": "email",
            "description": "Customer email address",
            "example": "john@example.com"
          },
          "created_at": {
            "type": "string",
            "format": "date-time",
            "description": "Customer creation timestamp",
            "example": "2024-01-01T00:00:00Z"
          },
          "updated_at": {
            "type": "string",
            "format": "date-time",
            "description": "Customer last update timestamp",
            "example": "2024-01-01T00:00:00Z"
          }
        }
      },
      "Order": {
        "type": "object",
        "required": ["customer_id", "item", "amount"],
        "properties": {
          "id": {
            "type": "string",
            "format": "uuid",
            "description": "Unique identifier for the order",
            "example": "123e4567-e89b-12d3-a456-426614174000"
          },
          "customer_id": {
            "type": "string",
            "format": "uuid",
            "description": "ID of the customer who placed the order",
            "example": "123e4567-e89b-12d3-a456-426614174000"
          },
          "item": {
            "type": "string",
            "description": "Name or description of the ordered item",
            "example": "Premium Service Package",
            "minLength": 1,
            "maxLength": 200
          },
          "amount": {
            "type": "number",
            "format": "double",
            "description": "Order amount in currency units",
            "example": 99.99,
            "minimum": 0
          },
          "status": {
            "type": "string",
            "enum": ["pending", "confirmed", "completed", "cancelled"],
            "description": "Current status of the order",
            "example": "pending"
          },
          "created_at": {
            "type": "string",
            "format": "date-time",
            "description": "Order creation timestamp",
            "example": "2024-01-01T00:00:00Z"
          },
          "updated_at": {
            "type": "string",
            "format": "date-time",
            "description": "Order last update timestamp",
            "example": "2024-01-01T00:00:00Z"
          }
        }
      },
      "ApiResponse": {
        "type": "object",
        "properties": {
          "success": {
            "type": "boolean",
            "description": "Indicates if the request was successful",
            "example": true
          },
          "message": {
            "type": "string",
            "description": "Human-readable response message",
            "example": "Operation completed successfully"
          },
          "data": {
            "description": "Response data (varies by endpoint)"
          },
          "error": {
            "type": "string",
            "description": "Error message (present only when success is false)",
            "example": null
          }
        }
      },
      "ErrorResponse": {
        "type": "object",
        "properties": {
          "success": {
            "type": "boolean",
            "example": false
          },
          "message": {
            "type": "string",
            "example": "Request failed"
          },
          "error": {
            "type": "string",
            "example": "Validation error: name is required"
          }
        }
      },
      "HealthResponse": {
        "type": "object",
        "properties": {
          "status": {
            "type": "string",
            "enum": ["healthy", "unhealthy", "degraded"],
            "example": "healthy"
          },
          "timestamp": {
            "type": "string",
            "format": "date-time",
            "example": "2024-01-01T00:00:00Z"
          },
          "uptime": {
            "type": "string",
            "example": "1h30m45s"
          },
          "version": {
            "type": "string",
            "example": "1.0.0"
          },
          "components": {
            "type": "object",
            "properties": {
              "database": {
                "$ref": "#/components/schemas/ComponentHealth"
              },
              "redis": {
                "$ref": "#/components/schemas/ComponentHealth"
              },
              "sms_service": {
                "$ref": "#/components/schemas/ComponentHealth"
              }
            }
          }
        }
      },
      "ComponentHealth": {
        "type": "object",
        "properties": {
          "status": {
            "type": "string",
            "enum": ["healthy", "unhealthy", "degraded"],
            "example": "healthy"
          },
          "message": {
            "type": "string",
            "example": "Database is healthy"
          },
          "last_checked": {
            "type": "string",
            "format": "date-time",
            "example": "2024-01-01T00:00:00Z"
          },
          "duration": {
            "type": "string",
            "example": "5ms"
          }
        }
      }
    },
    "responses": {
      "BadRequest": {
        "description": "Bad Request - Invalid input parameters",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "Unauthorized": {
        "description": "Unauthorized - Authentication required",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "Forbidden": {
        "description": "Forbidden - Insufficient permissions",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "NotFound": {
        "description": "Not Found - Resource does not exist",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "InternalServerError": {
        "description": "Internal Server Error",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      }
    }
  },
  "paths": {
    "/health": {
      "get": {
        "summary": "Health Check",
        "description": "Get the health status of the API and its dependencies",
        "operationId": "getHealth",
        "tags": ["Health"],
        "security": [],
        "responses": {
          "200": {
            "description": "Health status retrieved successfully",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/HealthResponse"
                }
              }
            }
          },
          "503": {
            "description": "Service unavailable - System is unhealthy",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/HealthResponse"
                }
              }
            }
          }
        }
      }
    },
    "/health/ready": {
      "get": {
        "summary": "Readiness Check",
        "description": "Check if the service is ready to receive traffic (Kubernetes readiness probe)",
        "operationId": "getReadiness",
        "tags": ["Health"],
        "security": [],
        "responses": {
          "200": {
            "description": "Service is ready"
          },
          "503": {
            "description": "Service is not ready"
          }
        }
      }
    },
    "/health/live": {
      "get": {
        "summary": "Liveness Check",
        "description": "Check if the service is alive (Kubernetes liveness probe)",
        "operationId": "getLiveness",
        "tags": ["Health"],
        "security": [],
        "responses": {
          "200": {
            "description": "Service is alive"
          }
        }
      }
    },
    "/metrics": {
      "get": {
        "summary": "Prometheus Metrics",
        "description": "Get Prometheus metrics for monitoring",
        "operationId": "getMetrics",
        "tags": ["Monitoring"],
        "security": [],
        "responses": {
          "200": {
            "description": "Metrics retrieved successfully",
            "content": {
              "text/plain": {
                "schema": {
                  "type": "string"
                }
              }
            }
          }
        }
      }
    },
    "/api/v1/customers": {
      "get": {
        "summary": "List Customers",
        "description": "Retrieve a list of customers with optional pagination",
        "operationId": "listCustomers",
        "tags": ["Customers"],
        "parameters": [
          {
            "name": "limit",
            "in": "query",
            "description": "Maximum number of customers to return",
            "schema": {
              "type": "integer",
              "minimum": 1,
              "maximum": 100,
              "default": 20
            }
          },
          {
            "name": "offset",
            "in": "query",
            "description": "Number of customers to skip",
            "schema": {
              "type": "integer",
              "minimum": 0,
              "default": 0
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Customers retrieved successfully",
            "content": {
              "application/json": {
                "schema": {
                  "allOf": [
                    {
                      "$ref": "#/components/schemas/ApiResponse"
                    },
                    {
                      "properties": {
                        "data": {
                          "type": "array",
                          "items": {
                            "$ref": "#/components/schemas/Customer"
                          }
                        }
                      }
                    }
                  ]
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "401": {
            "$ref": "#/components/responses/Unauthorized"
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      },
      "post": {
        "summary": "Create Customer",
        "description": "Create a new customer",
        "operationId": "createCustomer",
        "tags": ["Customers"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["name", "code"],
                "properties": {
                  "name": {
                    "type": "string",
                    "minLength": 2,
                    "maxLength": 100,
                    "example": "John Doe"
                  },
                  "code": {
                    "type": "string",
                    "pattern": "^[A-Z0-9]{4,20}$",
                    "example": "CUST001"
                  },
                  "phone_number": {
                    "type": "string",
                    "pattern": "^\\+[1-9]\\d{1,14}$",
                    "example": "+254700123456"
                  },
                  "email": {
                    "type": "string",
                    "format": "email",
                    "example": "john@example.com"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Customer created successfully",
            "content": {
              "application/json": {
                "schema": {
                  "allOf": [
                    {
                      "$ref": "#/components/schemas/ApiResponse"
                    },
                    {
                      "properties": {
                        "data": {
                          "$ref": "#/components/schemas/Customer"
                        }
                      }
                    }
                  ]
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "401": {
            "$ref": "#/components/responses/Unauthorized"
          },
          "409": {
            "description": "Conflict - Customer with this code already exists",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      }
    },
    "/api/v1/customers/{id}": {
      "get": {
        "summary": "Get Customer",
        "description": "Retrieve a specific customer by ID",
        "operationId": "getCustomer",
        "tags": ["Customers"],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "description": "Customer ID",
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Customer retrieved successfully",
            "content": {
              "application/json": {
                "schema": {
                  "allOf": [
                    {
                      "$ref": "#/components/schemas/ApiResponse"
                    },
                    {
                      "properties": {
                        "data": {
                          "$ref": "#/components/schemas/Customer"
                        }
                      }
                    }
                  ]
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "401": {
            "$ref": "#/components/responses/Unauthorized"
          },
          "404": {
            "$ref": "#/components/responses/NotFound"
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      },
      "put": {
        "summary": "Update Customer",
        "description": "Update an existing customer",
        "operationId": "updateCustomer",
        "tags": ["Customers"],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "description": "Customer ID",
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "name": {
                    "type": "string",
                    "minLength": 2,
                    "maxLength": 100,
                    "example": "John Doe"
                  },
                  "phone_number": {
                    "type": "string",
                    "pattern": "^\\+[1-9]\\d{1,14}$",
                    "example": "+254700123456"
                  },
                  "email": {
                    "type": "string",
                    "format": "email",
                    "example": "john@example.com"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Customer updated successfully",
            "content": {
              "application/json": {
                "schema": {
                  "allOf": [
                    {
                      "$ref": "#/components/schemas/ApiResponse"
                    },
                    {
                      "properties": {
                        "data": {
                          "$ref": "#/components/schemas/Customer"
                        }
                      }
                    }
                  ]
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "401": {
            "$ref": "#/components/responses/Unauthorized"
          },
          "404": {
            "$ref": "#/components/responses/NotFound"
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      },
      "delete": {
        "summary": "Delete Customer",
        "description": "Delete a customer (soft delete)",
        "operationId": "deleteCustomer",
        "tags": ["Customers"],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "description": "Customer ID",
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Customer deleted successfully",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ApiResponse"
                }
              }
            }
          },
          "401": {
            "$ref": "#/components/responses/Unauthorized"
          },
          "404": {
            "$ref": "#/components/responses/NotFound"
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      }
    },
    "/api/v1/orders": {
      "get": {
        "summary": "List Orders",
        "description": "Retrieve a list of orders with optional filtering",
        "operationId": "listOrders",
        "tags": ["Orders"],
        "parameters": [
          {
            "name": "customer_id",
            "in": "query",
            "description": "Filter orders by customer ID",
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          },
          {
            "name": "status",
            "in": "query",
            "description": "Filter orders by status",
            "schema": {
              "type": "string",
              "enum": ["pending", "confirmed", "completed", "cancelled"]
            }
          },
          {
            "name": "limit",
            "in": "query",
            "description": "Maximum number of orders to return",
            "schema": {
              "type": "integer",
              "minimum": 1,
              "maximum": 100,
              "default": 20
            }
          },
          {
            "name": "offset",
            "in": "query",
            "description": "Number of orders to skip",
            "schema": {
              "type": "integer",
              "minimum": 0,
              "default": 0
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Orders retrieved successfully",
            "content": {
              "application/json": {
                "schema": {
                  "allOf": [
                    {
                      "$ref": "#/components/schemas/ApiResponse"
                    },
                    {
                      "properties": {
                        "data": {
                          "type": "array",
                          "items": {
                            "$ref": "#/components/schemas/Order"
                          }
                        }
                      }
                    }
                  ]
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "401": {
            "$ref": "#/components/responses/Unauthorized"
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      },
      "post": {
        "summary": "Create Order",
        "description": "Create a new order and trigger SMS notification",
        "operationId": "createOrder",
        "tags": ["Orders"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["customer_id", "item", "amount"],
                "properties": {
                  "customer_id": {
                    "type": "string",
                    "format": "uuid",
                    "example": "123e4567-e89b-12d3-a456-426614174000"
                  },
                  "item": {
                    "type": "string",
                    "minLength": 1,
                    "maxLength": 200,
                    "example": "Premium Service Package"
                  },
                  "amount": {
                    "type": "number",
                    "format": "double",
                    "minimum": 0,
                    "example": 99.99
                  }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Order created successfully",
            "content": {
              "application/json": {
                "schema": {
                  "allOf": [
                    {
                      "$ref": "#/components/schemas/ApiResponse"
                    },
                    {
                      "properties": {
                        "data": {
                          "$ref": "#/components/schemas/Order"
                        }
                      }
                    }
                  ]
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "401": {
            "$ref": "#/components/responses/Unauthorized"
          },
          "404": {
            "description": "Customer not found"
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      }
    },
    "/api/v1/orders/{id}": {
      "get": {
        "summary": "Get Order",
        "description": "Retrieve a specific order by ID",
        "operationId": "getOrder",
        "tags": ["Orders"],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "description": "Order ID",
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Order retrieved successfully",
            "content": {
              "application/json": {
                "schema": {
                  "allOf": [
                    {
                      "$ref": "#/components/schemas/ApiResponse"
                    },
                    {
                      "properties": {
                        "data": {
                          "$ref": "#/components/schemas/Order"
                        }
                      }
                    }
                  ]
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "401": {
            "$ref": "#/components/responses/Unauthorized"
          },
          "404": {
            "$ref": "#/components/responses/NotFound"
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      }
    }
  },
  "tags": [
    {
      "name": "Health",
      "description": "Health check and monitoring endpoints"
    },
    {
      "name": "Monitoring",
      "description": "Monitoring and metrics endpoints"
    },
    {
      "name": "Customers",
      "description": "Customer management operations"
    },
    {
      "name": "Orders",
      "description": "Order management operations"
    }
  ]
}`
