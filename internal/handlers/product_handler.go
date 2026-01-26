package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"go-rest/internal/models"
	"go-rest/internal/repository"

	"github.com/gin-gonic/gin"
)

// ProductHandler handles product HTTP requests
type ProductHandler struct {
	repo repository.ProductRepository
}

// NewProductHandler creates a new product handler
func NewProductHandler(repo repository.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

// CreateProductRequest represents request body for creating a product
type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required" example:"Gaming Laptop"`
	Description string  `json:"description" example:"High-performance gaming laptop"`
	Price       float64 `json:"price" binding:"required,gt=0" example:"1500.00"`
	Stock       int     `json:"stock" binding:"gte=0" example:"10"`
}

// UpdateProductRequest represents request body for updating a product
type UpdateProductRequest struct {
	Name        string  `json:"name" binding:"required" example:"Gaming Laptop"`
	Description string  `json:"description" example:"High-performance gaming laptop"`
	Price       float64 `json:"price" binding:"required,gt=0" example:"1800.00"`
	Stock       int     `json:"stock" binding:"gte=0" example:"5"`
}

// ProductResponse represents response body for product
type ProductResponse struct {
	ID          uint    `json:"id" example:"1"`
	Name        string  `json:"name" example:"Gaming Laptop"`
	Description string  `json:"description" example:"High-performance gaming laptop"`
	Price       float64 `json:"price" example:"1500.00"`
	Stock       int     `json:"stock" example:"10"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error string `json:"error" example:"Product not found"`
}

// PaginatedResponse represents paginated response
type PaginatedResponse struct {
	Data  []models.Product `json:"data"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with the provided information
// @Tags products
// @Accept json
// @Produce json
// @Param product body CreateProductRequest true "Product information"
// @Success 201 {object} ProductResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	if err := h.repo.Create(product); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create product"})
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

	c.JSON(http.StatusCreated, ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
	})
}

// GetProduct godoc
// @Summary Get a product by ID
// @Description Retrieve a single product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} ProductResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid product ID"})
		return
	}

	product, err := h.repo.GetByID(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve product"})
		return
	}

	c.JSON(http.StatusOK, ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve products"})
		return
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:  products,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update an existing product with the provided information
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param product body UpdateProductRequest true "Product information"
// @Success 200 {object} ProductResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid product ID"})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Check if product exists
	product, err := h.repo.GetByID(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve product"})
		return
	}

	// Update product fields
	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock

	if err := h.repo.Update(product); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid product ID"})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete product"})
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if len(req.Products) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Products array cannot be empty"})
		return
	}

	if len(req.Products) > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Cannot create more than 100 products at once"})
		return
	}

	// Convert request to models
	products := make([]*models.Product, len(req.Products))
	for i, p := range req.Products {
		products[i] = &models.Product{
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
		}
	}

	// Bulk create using goroutines in repository
	if err := h.repo.BulkCreate(products); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create products"})
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "IDs parameter is required"})
		return
	}

	// Parse comma-separated IDs
	var ids []uint
	for _, idStr := range strings.Split(idsStr, ",") {
		id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format: " + idStr})
			return
		}
		ids = append(ids, uint(id))
	}

	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "At least one ID is required"})
		return
	}

	if len(ids) > 50 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Cannot fetch more than 50 products at once"})
		return
	}

	// Fetch products in parallel using goroutines
	products, err := h.repo.GetMultipleIDs(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve products"})
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
