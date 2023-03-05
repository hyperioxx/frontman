package frontman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hyperioxx/frontman/service"
)

// TestGetServicesHandler tests the getServicesHandler function
func TestGetServicesHandler(t *testing.T) {
	// Create a new request
	req, err := http.NewRequest("GET", "/api/services", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new recorder to record the response
	rr := httptest.NewRecorder()

	// Create a new backend service registry
	backendServices := service.NewMemoryServiceRegistry()

	// Call the handler function
	handler := http.HandlerFunc(getServicesHandler(backendServices))
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := `[]`
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

// TestAddServiceHandler tests the addServiceHandler function
func TestAddServiceHandler(t *testing.T) {
	// Define a sample backend service
	bs := &service.BackendService{
		Name:            "test_service",
		Scheme:          "http",
		UpstreamTargets: []string{"http://localhost:8080"},
		Path:            "/api/test",
		Domain:          "localhost",
		HealthCheck:     "/health",
		RetryAttempts:   3,
		Timeout:         10,
		MaxIdleConns:    100,
		MaxIdleTime:     30,
		StripPath:       true,
	}

	// Marshal the backend service into JSON
	bsJSON, err := json.Marshal(bs)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new request with the JSON body
	req, err := http.NewRequest("POST", "/api/services", bytes.NewBuffer(bsJSON))
	if err != nil {
		t.Fatal(err)
	}

	// Create a new recorder to record the response
	rr := httptest.NewRecorder()

	// Create a new backend service registry
	backendServices := service.NewMemoryServiceRegistry()

	// Call the handler function
	handler := http.HandlerFunc(addServiceHandler(backendServices))
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Check the response body
	expected := "{\"name\":\"test_service\",\"scheme\":\"http\",\"upstreamTargets\":[\"http://localhost:8080\"],\"path\":\"/api/test\",\"domain\":\"localhost\",\"healthCheck\":\"/health\",\"retryAttempts\":3,\"timeout\":10,\"maxIdleConns\":100,\"maxIdleTime\":30,\"stripPath\":true}\n"
	if rr.Body.String() != expected {
		fmt.Println(rr.Body.String())
		fmt.Println(expected)
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	// Check that the backend service was added
	if len(backendServices.GetServices()) != 1 {
		t.Errorf("Expected 1 service to be added to the backend service registry, but got %v", len(backendServices.GetServices()))
	}
}

// TestRemoveServiceHandler tests the removeServiceHandler function
func TestRemoveServiceHandler(t *testing.T) {
	// Create a new request
	req, err := http.NewRequest("DELETE", "/api/services/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new recorder to record the response
	rr := httptest.NewRecorder()

	// Create a new backend service registry
	backendServices := service.NewMemoryServiceRegistry()

	// Call the handler function
	handler := http.HandlerFunc(removeServiceHandler(backendServices))
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	// Check the response body
	expected := "{\"error\":\"missing service name\"}\n"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
