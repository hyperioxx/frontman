package frontman

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetServicesHandler(t *testing.T) {
    // Create a mock BackendServices instance
    mockServices := []*BackendService{
        {
            Name:            "test-service-1",
            Scheme:          "http",
            UpstreamTargets: []string{"http://localhost:8080"},
            Path:            "/test1",
            Domain:          "example.com",
            HealthCheck:     "http://localhost:8080/health",
            RetryAttempts:   3,
            Timeout:         5,
            MaxIdleConns:    100,
            MaxIdleTime:     30,
            StripPath:       false,
        },
        {
            Name:            "test-service-2",
            Scheme:          "http",
            UpstreamTargets: []string{"http://localhost:8081"},
            Path:            "/test2",
            Domain:          "example.com",
            HealthCheck:     "http://localhost:8081/health",
            RetryAttempts:   3,
            Timeout:         5,
            MaxIdleConns:    100,
            MaxIdleTime:     30,
            StripPath:       false,
        },
    }
    mockBackendServices := &BackendServices{
        services: make(map[string]*BackendService),
    }
    for _, service := range mockServices {
        mockBackendServices.services[service.Name] = service
    }

    // Create a mock HTTP request and response
    req, err := http.NewRequest("GET", "/services", nil)
    if err != nil {
        t.Fatal(err)
    }
    rr := httptest.NewRecorder()

    // Call the getServicesHandler function and pass in the mock BackendServices instance
    handler := http.HandlerFunc(getServicesHandler(mockBackendServices))
    handler.ServeHTTP(rr, req)

    // Check that the HTTP response status code is 200 OK
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v, expected %v", status, http.StatusOK)
    }

    // Check that the response body contains the expected JSON data
    expected := `[{"name":"test-service-1","scheme":"http","upstreamTargets":["http://localhost:8080"],"path":"/test1","domain":"example.com","healthCheck":"http://localhost:8080/health","retryAttempts":3,"timeout":5,"maxIdleConns":100,"maxIdleTime":30,"stripPath":false,"Provider":null},{"name":"test-service-2","scheme":"http","upstreamTargets":["http://localhost:8081"],"path":"/test2","domain":"example.com","healthCheck":"http://localhost:8081/health","retryAttempts":3,"timeout":5,"maxIdleConns":100,"maxIdleTime":30,"stripPath":false,"Provider":null}]`
    if rr.Body.String() != expected {
        t.Errorf("handler returned unexpected body: got %v, expected %v", rr.Body.String(), expected)
    }
}

