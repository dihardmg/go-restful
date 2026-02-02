package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-rest/internal/models"
	"go-rest/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProductHandler handles product HTTP requests
type ProductHandler struct {
	repo repository.ProductRepository
}

const (
	// TraceIDKey is the context key for trace ID
	TraceIDKey = "trace_id"
)

// NewProductHandler creates a new product handler
func NewProductHandler(repo repository.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

// CreateProductRequest represents request body for creating a product
type CreateProductRequest struct {
	Name        string      `json:"name" example:"Gaming Laptop"`
	Description string      `json:"description" example:"High-performance gaming laptop"`
	Price       interface{} `json:"price" example:"1500.00"`
	Stock       interface{} `json:"stock" example:"10"`
}

// UpdateProductRequest represents request body for updating a product
type UpdateProductRequest struct {
	Name        string      `json:"name" example:"Gaming Laptop"`
	Description string      `json:"description" example:"High-performance gaming laptop"`
	Price       interface{} `json:"price" example:"1800.00"`
	Stock       interface{} `json:"stock" example:"5"`
}

// ProductResponse represents response body for product
type ProductResponse struct {
	ID          uint       `json:"id" example:"1"`
	Name        string     `json:"name" example:"Gaming Laptop"`
	Description string     `json:"description" example:"High-performance gaming laptop"`
	Price       float64    `json:"price" example:"1500.00"`
	Stock       int        `json:"stock" example:"10"`
	CreatedAt   time.Time  `json:"created_at" example:"2026-01-28T13:35:27.813525Z"`
	UpdatedAt   *time.Time `json:"updated_at" example:"2026-01-28T13:35:27.813525Z"`
}

// ErrorDetail represents error detail
type ErrorDetail struct {
	Code    string `json:"code" example:"NOT_FOUND"`
	Message string `json:"message" example:"Product with id 999 not found"`
}

// ErrorMeta represents metadata for error response
type ErrorMeta struct {
	Timestamp string `json:"timestamp" example:"2026-01-16T14:23:41.761Z"`
	TraceID   string `json:"trace_id" example:"b79b7627-fc60-4a43-8215-efb4739f7d1d"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   ErrorDetail            `json:"error"`
	Details *ValidationErrorDetails `json:"details,omitempty"`
	Meta    ErrorMeta               `json:"meta"`
}

// ValidationErrorDetails represents validation error details
type ValidationErrorDetails struct {
	Name  []string `json:"name,omitempty"`
	Price []string `json:"price,omitempty"`
	Stock []string `json:"stock,omitempty"`
}

// MetaResponse represents metadata for paginated response
type MetaResponse struct {
	TraceID    string `json:"trace_id" example:"1234555xxxx"`
	Total      int64  `json:"total" example:"14"`
	Page       int    `json:"page" example:"1"`
	Limit      int    `json:"limit" example:"10"`
	TotalPages int    `json:"total_pages" example:"2"`
}

// SuccessMeta represents metadata for success response
type SuccessMeta struct {
	TraceID string `json:"trace_id" example:"815ad63d-26ff-4940-83dd-c74a115187b9"`
}

// PaginatedResponse represents paginated response
type PaginatedResponse struct {
	Data []ProductResponse `json:"data"`
	Meta MetaResponse      `json:"meta"`
}

// SingleProductResponse represents single product response with data and meta
type SingleProductResponse struct {
	Data ProductResponse `json:"data"`
	Meta SuccessMeta     `json:"meta"`
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with the provided information. Validation rules:<br>- name: required, min 3 characters<br>- price: required, must be number, must be greater than 0<br>- stock: required, must be number, cannot be negative (can be 0)
// @Tags products
// @Accept json
// @Produce json
// @Param product body CreateProductRequest true "Product information"
// @Success 201 {object} SingleProductResponse
// @Failure 400 {object} ErrorResponse "Bad Request - Invalid JSON format"
// @Failure 422 {object} ErrorResponse "Validation Error - Check 'details' field for specific validation errors per field"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponseWithErr(c, "BAD_REQUEST", err))
		return
	}

	// Validate all fields and collect errors
	if validationErrors := validateCreateProduct(req); len(validationErrors.Name) > 0 ||
		len(validationErrors.Price) > 0 || len(validationErrors.Stock) > 0 {
		c.JSON(http.StatusUnprocessableEntity, newValidationErrorResponse(c, validationErrors))
		return
	}

	// Convert interface{} to proper types
	price, stock, err := convertCreateProductRequest(&req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, newErrorResponse(c, "VALIDATION_ERROR", err.Error()))
		return
	}

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       price,
		Stock:       stock,
	}

	if err := h.repo.Create(product); err != nil {
		c.JSON(http.StatusInternalServerError, newErrorResponse(c, "INTERNAL_ERROR", "Failed to create product"))
		return
	}

	// BACKGROUND PROCESSING: Non-blocking goroutines for async operations
	// These operations run in parallel and don't block the HTTP response
	// Performance: User gets immediate response while background tasks complete
	go func() {
		// Log product creation (analytics)
		h.logProductCreation(product)

		// Index for search (e.g., Elasticsearch, Meilisearch)
		h.indexProduct(product)

		// Warm cache (e.g., Redis)
		h.warmProductCache(product.ID)
	}()

	c.JSON(http.StatusCreated, SingleProductResponse{
		Data: ProductResponse{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       product.Stock,
			CreatedAt:   product.CreatedAt,
			UpdatedAt:   product.UpdatedAt,
		},
		Meta: SuccessMeta{
			TraceID: getTraceID(c),
		},
	})
}

// GetProduct godoc
// @Summary Get a product by ID
// @Description Retrieve a single product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} SingleProductResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(c, "BAD_REQUEST", "Invalid product ID"))
		return
	}

	product, err := h.repo.GetByID(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, newErrorResponse(c, "NOT_FOUND", "Product not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, newErrorResponse(c, "INTERNAL_ERROR", "Failed to retrieve product"))
		return
	}

	c.JSON(http.StatusOK, SingleProductResponse{
		Data: ProductResponse{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       product.Stock,
			CreatedAt:   product.CreatedAt,
			UpdatedAt:   product.UpdatedAt,
		},
		Meta: SuccessMeta{
			TraceID: getTraceID(c),
		},
	})
}

// GetProducts godoc
// @Summary Get all products
// @Description Retrieve a paginated list of all products
// @Tags products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /products [get]
func (h *ProductHandler) GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	products, total, err := h.repo.GetAll(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, newErrorResponse(c, "INTERNAL_ERROR", "Failed to retrieve products"))
		return
	}

	// Convert models.Product to ProductResponse
	productResponses := make([]ProductResponse, len(products))
	for i, p := range products {
		productResponses[i] = ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
	}

	// Calculate total pages
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Data: productResponses,
		Meta: MetaResponse{
			TraceID:    getTraceID(c),
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update an existing product with the provided information. Validation rules:<br>- name: required, min 3 characters<br>- price: required, must be number, must be greater than 0<br>- stock: required, must be number, cannot be negative (can be 0)
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param product body UpdateProductRequest true "Product information"
// @Success 200 {object} SingleProductResponse
// @Failure 400 {object} ErrorResponse "Bad Request - Invalid JSON format"
// @Failure 404 {object} ErrorResponse "Product Not Found"
// @Failure 422 {object} ErrorResponse "Validation Error - Check 'details' field for specific validation errors per field"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(c, "BAD_REQUEST", "Invalid product ID"))
		return
	}

	var req UpdateProductRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponseWithErr(c, "BAD_REQUEST", err))
		return
	}

	// Validate all fields and collect errors
	if validationErrors := validateUpdateProduct(req); len(validationErrors.Name) > 0 ||
		len(validationErrors.Price) > 0 || len(validationErrors.Stock) > 0 {
		c.JSON(http.StatusUnprocessableEntity, newValidationErrorResponse(c, validationErrors))
		return
	}

	// Convert interface{} to proper types
	price, stock, err := convertUpdateProductRequest(&req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, newErrorResponse(c, "VALIDATION_ERROR", err.Error()))
		return
	}

	// Check if product exists
	product, err := h.repo.GetByID(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, newErrorResponse(c, "NOT_FOUND", "Product not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, newErrorResponse(c, "INTERNAL_ERROR", "Failed to retrieve product"))
		return
	}

	// Update product fields
	product.Name = req.Name
	product.Description = req.Description
	product.Price = price
	product.Stock = stock

	// Set UpdatedAt to current time
	now := time.Now()
	product.UpdatedAt = &now

	if err := h.repo.Update(product); err != nil {
		c.JSON(http.StatusInternalServerError, newErrorResponse(c, "INTERNAL_ERROR", "Failed to update product"))
		return
	}

	c.JSON(http.StatusOK, SingleProductResponse{
		Data: ProductResponse{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       product.Stock,
			CreatedAt:   product.CreatedAt,
			UpdatedAt:   product.UpdatedAt,
		},
		Meta: SuccessMeta{
			TraceID: getTraceID(c),
		},
	})
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(c, "BAD_REQUEST", "Invalid product ID"))
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, newErrorResponse(c, "NOT_FOUND", "Product not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, newErrorResponse(c, "INTERNAL_ERROR", "Failed to delete product"))
		return
	}

	c.Status(http.StatusNoContent)
}

// BulkCreateRequest represents request body for bulk creating products
type BulkCreateRequest struct {
	Products []CreateProductRequest `json:"products" binding:"required"`
}

// BulkCreate godoc
// @Summary Bulk create products
// @Description Create multiple products in parallel using goroutines
// @Tags products
// @Accept json
// @Produce json
// @Param products body BulkCreateRequest true "Array of products"
// @Success 201 {object} []ProductResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/bulk [post]
func (h *ProductHandler) BulkCreate(c *gin.Context) {
	var req BulkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponseWithErr(c, "BAD_REQUEST", err))
		return
	}

	if len(req.Products) == 0 {
		c.JSON(http.StatusBadRequest, newErrorResponse(c, "BAD_REQUEST", "Products array cannot be empty"))
		return
	}

	if len(req.Products) > 100 {
		c.JSON(http.StatusBadRequest, newErrorResponse(c, "BAD_REQUEST", "Cannot create more than 100 products at once"))
		return
	}

	// Convert request to models
	products := make([]*models.Product, len(req.Products))
	for i, p := range req.Products {
		price, stock, err := convertCreateProductRequest(&p)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, newErrorResponse(c, "VALIDATION_ERROR", fmt.Sprintf("Product %d: %s", i+1, err.Error())))
			return
		}

		products[i] = &models.Product{
			Name:        p.Name,
			Description: p.Description,
			Price:       price,
			Stock:       stock,
		}
	}

	// Bulk create using goroutines in repository
	if err := h.repo.BulkCreate(products); err != nil {
		c.JSON(http.StatusInternalServerError, newErrorResponse(c, "INTERNAL_ERROR", "Failed to create products"))
		return
	}

	// Convert to response format
	responses := make([]ProductResponse, len(products))
	for i, p := range products {
		responses[i] = ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
	}

	c.JSON(http.StatusCreated, responses)
}

// GetMultipleProducts godoc
// @Summary Get multiple products by IDs
// @Description Retrieve multiple products by their IDs in parallel using goroutines
// @Tags products
// @Accept json
// @Produce json
// @Param ids query string true "Comma-separated product IDs" example(1,2,3)
// @Success 200 {object} []ProductResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/multiple [get]
func (h *ProductHandler) GetMultipleProducts(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr == "" {
		c.JSON(http.StatusBadRequest, newErrorResponse(c, "BAD_REQUEST", "IDs parameter is required"))
		return
	}

	// Parse comma-separated IDs
	var ids []uint
	for _, idStr := range strings.Split(idsStr, ",") {
		id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, newErrorResponse(c, "BAD_REQUEST", "Invalid ID format: "+idStr))
			return
		}
		ids = append(ids, uint(id))
	}

	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, newErrorResponse(c, "BAD_REQUEST", "At least one ID is required"))
		return
	}

	if len(ids) > 50 {
		c.JSON(http.StatusBadRequest, newErrorResponse(c, "BAD_REQUEST", "Cannot fetch more than 50 products at once"))
		return
	}

	// Fetch products in parallel using goroutines
	products, err := h.repo.GetMultipleIDs(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, newErrorResponse(c, "INTERNAL_ERROR", "Failed to retrieve products"))
		return
	}

	// Convert to response format
	responses := make([]ProductResponse, len(products))
	for i, p := range products {
		responses[i] = ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, responses)
}

// logProductCreation logs product creation asynchronously (non-blocking)
func (h *ProductHandler) logProductCreation(product *models.Product) {
	// In production, this would send to analytics service, logging system, etc.
	// For now, we'll just simulate it
	// TODO: Integrate with logging/analytics service
}

// indexProduct indexes product for search (non-blocking)
func (h *ProductHandler) indexProduct(product *models.Product) {
	// In production, this would update search index (Elasticsearch, Meilisearch, etc.)
	// TODO: Integrate with search service
}

// warmProductCache warms cache for frequently accessed products (non-blocking)
func (h *ProductHandler) warmProductCache(productID uint) {
	// In production, this would update Redis cache or similar
	// TODO: Integrate with caching layer
}

// getTraceID retrieves trace ID from Gin context
func getTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(TraceIDKey); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return uuid.New().String()
}

// generateTraceID generates a unique trace ID for logging and monitoring
func generateTraceID() string {
	return uuid.New().String()
}

// newErrorResponse creates a new error response with code, message, and trace ID
func newErrorResponse(c *gin.Context, code, message string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
		Meta: ErrorMeta{
			Timestamp: time.Now().Format(time.RFC3339Nano),
			TraceID:   getTraceID(c),
		},
	}
}

// newErrorResponseWithErr creates a new error response from error object
func newErrorResponseWithErr(c *gin.Context, code string, err error) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: err.Error(),
		},
		Meta: ErrorMeta{
			Timestamp: time.Now().Format(time.RFC3339Nano),
			TraceID:   getTraceID(c),
		},
	}
}

// newValidationErrorResponse creates a new validation error response with field-level details
func newValidationErrorResponse(c *gin.Context, details ValidationErrorDetails) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid request data",
		},
		Details: &details,
		Meta: ErrorMeta{
			Timestamp: time.Now().Format(time.RFC3339Nano),
			TraceID:   getTraceID(c),
		},
	}
}

// validateCreateProduct validates product creation request and returns all validation errors
func validateCreateProduct(req CreateProductRequest) ValidationErrorDetails {
	var details ValidationErrorDetails

	// Validate name field
	if req.Name == "" {
		details.Name = append(details.Name, "must not be blank")
	} else if len(req.Name) < 3 {
		details.Name = append(details.Name, "min length is 3")
	}

	// Validate price field - check type first
	if req.Price == nil {
		details.Price = append(details.Price, "is required")
	} else {
		switch v := req.Price.(type) {
		case float64:
			if v <= 0 {
				details.Price = append(details.Price, "must be greater than 0")
			}
		case string:
			details.Price = append(details.Price, "must be a number, not string")
		default:
			details.Price = append(details.Price, "must be a number")
		}
	}

	// Validate stock field - check type first
	if req.Stock == nil {
		details.Stock = append(details.Stock, "is required")
	} else {
		switch v := req.Stock.(type) {
		case float64:
			if v < 0 {
				details.Stock = append(details.Stock, "must not be negative")
			}
		case string:
			details.Stock = append(details.Stock, "must be a number, not string")
		default:
			details.Stock = append(details.Stock, "must be a number")
		}
	}

	return details
}

// validateUpdateProduct validates product update request and returns all validation errors
func validateUpdateProduct(req UpdateProductRequest) ValidationErrorDetails {
	var details ValidationErrorDetails

	// Validate name field
	if req.Name == "" {
		details.Name = append(details.Name, "must not be blank")
	} else if len(req.Name) < 3 {
		details.Name = append(details.Name, "min length is 3")
	}

	// Validate price field - check type first
	if req.Price == nil {
		details.Price = append(details.Price, "is required")
	} else {
		switch v := req.Price.(type) {
		case float64:
			if v <= 0 {
				details.Price = append(details.Price, "must be greater than 0")
			}
		case string:
			details.Price = append(details.Price, "must be a number, not string")
		default:
			details.Price = append(details.Price, "must be a number")
		}
	}

	// Validate stock field - check type first
	if req.Stock == nil {
		details.Stock = append(details.Stock, "is required")
	} else {
		switch v := req.Stock.(type) {
		case float64:
			if v < 0 {
				details.Stock = append(details.Stock, "must not be negative")
			}
		case string:
			details.Stock = append(details.Stock, "must be a number, not string")
		default:
			details.Stock = append(details.Stock, "must be a number")
		}
	}

	return details
}

// convertCreateProductRequest converts interface{} values to proper types
func convertCreateProductRequest(req *CreateProductRequest) (price float64, stock int, err error) {
	// Convert price
	if req.Price != nil {
		switch v := req.Price.(type) {
		case float64:
			price = v
		case int:
			price = float64(v)
		case json.Number:
			f, err := v.Float64()
			if err != nil {
				return 0, 0, fmt.Errorf("invalid price format")
			}
			price = f
		default:
			return 0, 0, fmt.Errorf("price must be a number")
		}
	}

	// Convert stock
	if req.Stock != nil {
		switch v := req.Stock.(type) {
		case float64:
			stock = int(v)
		case int:
			stock = v
		case json.Number:
			i, err := v.Int64()
			if err != nil {
				return 0, 0, fmt.Errorf("invalid stock format")
			}
			stock = int(i)
		default:
			return 0, 0, fmt.Errorf("stock must be a number")
		}
	}

	return price, stock, nil
}

// convertUpdateProductRequest converts interface{} values to proper types
func convertUpdateProductRequest(req *UpdateProductRequest) (price float64, stock int, err error) {
	// Convert price
	if req.Price != nil {
		switch v := req.Price.(type) {
		case float64:
			price = v
		case int:
			price = float64(v)
		case json.Number:
			f, err := v.Float64()
			if err != nil {
				return 0, 0, fmt.Errorf("invalid price format")
			}
			price = f
		default:
			return 0, 0, fmt.Errorf("price must be a number")
		}
	}

	// Convert stock
	if req.Stock != nil {
		switch v := req.Stock.(type) {
		case float64:
			stock = int(v)
		case int:
			stock = v
		case json.Number:
			i, err := v.Int64()
			if err != nil {
				return 0, 0, fmt.Errorf("invalid stock format")
			}
			stock = int(i)
		default:
			return 0, 0, fmt.Errorf("stock must be a number")
		}
	}

	return price, stock, nil
}
