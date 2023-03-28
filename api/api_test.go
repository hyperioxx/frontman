package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Frontman-Labs/frontman/loadbalancer"
	"github.com/Frontman-Labs/frontman/service"
	"github.com/julienschmidt/httprouter"
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
	handler := getServicesHandler(backendServices)
	handler(rr, req, nil)

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
		LoadBalancerPolicy: service.LoadBalancerPolicy{
			Type:    loadbalancer.WeightedRoundRobin,
			Options: service.PolicyOptions{Weights: []int{3}},
		},
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
	handler := addServiceHandler(backendServices)
	handler(rr, req, nil)

	// Check the status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Check the response body
	expected := "{\"name\":\"test_service\",\"scheme\":\"http\",\"upstreamTargets\":[\"http://localhost:8080\"],\"path\":\"/api/test\",\"domain\":\"localhost\",\"healthCheck\":\"/health\",\"retryAttempts\":3,\"timeout\":10,\"maxIdleConns\":100,\"maxIdleTime\":30,\"stripPath\":true,\"loadBalancerPolicy\":{\"type\":\"weighted_round_robin\",\"options\":{\"weights\":[3]}}}\n"
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
	handler := removeServiceHandler(backendServices)
	handler(rr, req, httprouter.Params{{"name", "test"}})

	// Check the status code
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	// Check the response body
	expected := "service with name 'test' not found\n"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
