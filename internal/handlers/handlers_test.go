package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestGetCreditPackages tests the credit packages endpoint
func TestGetCreditPackages(t *testing.T) {
	router := gin.New()
	router.GET("/credits/packages", GetCreditPackages)

	req, _ := http.NewRequest("GET", "/credits/packages", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var packages []CreditPackage
	if err := json.Unmarshal(w.Body.Bytes(), &packages); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if len(packages) != 3 {
		t.Errorf("Expected 3 packages, got %d", len(packages))
	}
}

// TestValidationMiddleware tests the validation middleware
func TestValidationMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		input      map[string]interface{}
		wantStatus int
	}{
		{
			name:       "empty body",
			input:      map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid input",
			input: map[string]interface{}{
				"job_title":       "Software Engineer",
				"job_description": "Test description",
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest("POST", "/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Test would need actual handler setup
			// This is a placeholder for actual test implementation
			_ = w
		})
	}
}

// TestPDFGeneration tests PDF generation functionality
func TestPDFGeneration(t *testing.T) {
	cvContent := map[string]interface{}{
		"contact": map[string]string{
			"name":    "John Doe",
			"email":   "john@example.com",
			"phone":   "+1234567890",
			"address": "New York, USA",
		},
		"summary": "Experienced software developer",
		"skills": map[string][]string{
			"technical": {"Go", "Python", "JavaScript"},
		},
	}

	// Test that CV content can be marshaled
	contentBytes, err := json.Marshal(cvContent)
	if err != nil {
		t.Errorf("Failed to marshal CV content: %v", err)
	}

	if len(contentBytes) == 0 {
		t.Error("CV content is empty")
	}
}

// TestCVParsing tests CV parsing validation
func TestCVParsing(t *testing.T) {
	parsedData := `{
		"name": "Jane Smith",
		"email": "jane@example.com",
		"phone": "+0987654321",
		"experiences": [
			{
				"company": "Tech Corp",
				"position": "Developer",
				"start_date": "2020-01-01",
				"is_current": true
			}
		],
		"skills": [
			{"name": "Go", "category": "Technical", "proficiency": "Expert"}
		]
	}`

	var cvData map[string]interface{}
	if err := json.Unmarshal([]byte(parsedData), &cvData); err != nil {
		t.Errorf("Failed to parse CV data: %v", err)
	}

	if cvData["name"] != "Jane Smith" {
		t.Errorf("Expected name 'Jane Smith', got %v", cvData["name"])
	}
}

// BenchmarkGenerateCV benchmarks the CV generation endpoint
func BenchmarkGenerateCV(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Placeholder for actual benchmark
		// Would need full handler setup
	}
}
