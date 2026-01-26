package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-rest/internal/handlers"
	"go-rest/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var ErrProductNotFound = errors.New("product not found")

// MockProductRepository is a mock implementation of ProductRepository
type MockProductRepository struct {
	products map[uint]*models.Product
	nextID   uint
}

func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{
		products: make(map[uint]*models.Product),
		nextID:   1,
	}
}

func (m *MockProductRepository) Create(product *models.Product) error {
	product.ID = m.nextID
	m.nextID++
	m.products[product.ID] = product
	return nil
}

func (m *MockProductRepository) GetByID(id uint) (*models.Product, error) {
	if product, ok := m.products[id]; ok {
		return product, nil
	}
	return nil, ErrProductNotFound
}

func (m *MockProductRepository) GetAll(limit, offset int) ([]models.Product, int64, error) {
	products := make([]models.Product, 0, len(m.products))
	for _, p := range m.products {
		products = append(products, *p)
	}
	return products, int64(len(products)), nil
}

func (m *MockProductRepository) Update(product *models.Product) error {
	if _, ok := m.products[product.ID]; ok {
		m.products[product.ID] = product
		return nil
	}
	return ErrProductNotFound
}

func (m *MockProductRepository) Delete(id uint) error {
	if _, ok := m.products[id]; ok {
		delete(m.products, id)
		return nil
	}
	return ErrProductNotFound
}

func (m *MockProductRepository) GetMultipleIDs(ids []uint) ([]models.Product, error) {
	var products []models.Product
	for _, id := range ids {
		if product, ok := m.products[id]; ok {
			products = append(products, *product)
		}
	}
	return products, nil
}

func (m *MockProductRepository) BulkCreate(products []*models.Product) error {
	for _, product := range products {
		if err := m.Create(product); err != nil {
			return err
		}
	}
	return nil
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockRepo := NewMockProductRepository()
	productHandler := handlers.NewProductHandler(mockRepo)

	router.POST("/products", productHandler.CreateProduct)
	router.GET("/products/:id", productHandler.GetProduct)
	router.GET("/products", productHandler.GetProducts)
	router.PUT("/products/:id", productHandler.UpdateProduct)
	router.DELETE("/products/:id", productHandler.DeleteProduct)

	return router
}

// TestCreateProduct tests the create product endpoint
func TestCreateProduct(t *testing.T) {
	router := setupTestRouter()

	reqBody := map[string]interface{}{
		"name":        "Test Product",
		"description": "Test Description",
		"price":       100.50,
		"stock":       10,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "Test Product", response["name"])
	assert.Equal(t, 100.50, response["price"])
	assert.Equal(t, float64(10), response["stock"])
}

// TestCreateProductValidation tests validation on create product
func TestCreateProductValidation(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name       string
		reqBody    map[string]interface{}
		expectCode int
	}{
		{
			name: "Missing name",
			reqBody: map[string]interface{}{
				"description": "Test Description",
				"price":       100.50,
				"stock":       10,
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Zero price",
			reqBody: map[string]interface{}{
				"name":        "Test Product",
				"description": "Test Description",
				"price":       0,
				"stock":       10,
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Negative stock",
			reqBody: map[string]interface{}{
				"name":        "Test Product",
				"description": "Test Description",
				"price":       100.50,
				"stock":       -1,
			},
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectCode, w.Code)
		})
	}
}

// TestGetProductTests tests getting a product
func TestGetProduct(t *testing.T) {
	router := setupTestRouter()

	// First create a product
	createReqBody := map[string]interface{}{
		"name":        "Test Product",
		"description": "Test Description",
		"price":       100.50,
		"stock":       10,
	}
	body, _ := json.Marshal(createReqBody)
	createReq, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
	createReq.Header.Set("Content-Type", "application/json")

	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var createResponse map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResponse)
	productID := int(createResponse["id"].(float64))

	// Now get the product
	req, _ := http.NewRequest("GET", "/products/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "Test Product", response["name"])
	assert.Equal(t, float64(productID), response["id"])
}

// TestGetProductNotFound tests getting a non-existent product
func TestGetProductNotFound(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/products/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestGetProducts tests getting all products
func TestGetProducts(t *testing.T) {
	router := setupTestRouter()

	// Create some products
	for i := 1; i <= 3; i++ {
		reqBody := map[string]interface{}{
			"name":        "Test Product",
			"description": "Test Description",
			"price":       float64(i * 10),
			"stock":       i,
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Get all products
	req, _ := http.NewRequest("GET", "/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(3), response["total"])
	assert.NotNil(t, response["data"])
}

// TestUpdateProduct tests updating a product
func TestUpdateProduct(t *testing.T) {
	router := setupTestRouter()

	// First create a product
	createReqBody := map[string]interface{}{
		"name":        "Test Product",
		"description": "Test Description",
		"price":       100.50,
		"stock":       10,
	}
	body, _ := json.Marshal(createReqBody)
	createReq, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
	createReq.Header.Set("Content-Type", "application/json")

	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	// Update the product
	updateReqBody := map[string]interface{}{
		"name":        "Updated Product",
		"description": "Updated Description",
		"price":       200.50,
		"stock":       5,
	}
	updateBody, _ := json.Marshal(updateReqBody)
	updateReq, _ := http.NewRequest("PUT", "/products/1", bytes.NewBuffer(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")

	updateW := httptest.NewRecorder()
	router.ServeHTTP(updateW, updateReq)

	assert.Equal(t, http.StatusOK, updateW.Code)

	var response map[string]interface{}
	json.Unmarshal(updateW.Body.Bytes(), &response)

	assert.Equal(t, "Updated Product", response["name"])
	assert.Equal(t, 200.50, response["price"])
}

// TestDeleteProduct tests deleting a product
func TestDeleteProduct(t *testing.T) {
	router := setupTestRouter()

	// First create a product
	createReqBody := map[string]interface{}{
		"name":        "Test Product",
		"description": "Test Description",
		"price":       100.50,
		"stock":       10,
	}
	body, _ := json.Marshal(createReqBody)
	createReq, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
	createReq.Header.Set("Content-Type", "application/json")

	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	// Delete the product
	deleteReq, _ := http.NewRequest("DELETE", "/products/1", nil)
	deleteW := httptest.NewRecorder()
	router.ServeHTTP(deleteW, deleteReq)

	assert.Equal(t, http.StatusNoContent, deleteW.Code)
}
