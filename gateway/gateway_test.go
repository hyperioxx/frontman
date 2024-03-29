package gateway

import (
	"context"
	"github.com/Frontman-Labs/frontman/loadbalancer"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/Frontman-Labs/frontman/config"
	"github.com/Frontman-Labs/frontman/log"
	"github.com/Frontman-Labs/frontman/plugins"
	"github.com/Frontman-Labs/frontman/service"
)

type testPlugin struct {
	preRequest   func(*http.Request, service.ServiceRegistry, *config.Config) plugins.PluginError
	postResponse func(*http.Response, service.ServiceRegistry, *config.Config) plugins.PluginError
	close        func() plugins.PluginError
}

func (p *testPlugin) Name() string {
	return ""
}

func (p *testPlugin) PreRequest(r *http.Request, s service.ServiceRegistry, c *config.Config) plugins.PluginError {
	if p.preRequest != nil {
		return p.preRequest(r, s, c)
	}
	return nil
}

func (p *testPlugin) PostResponse(resp *http.Response, s service.ServiceRegistry, c *config.Config) plugins.PluginError {
	if p.postResponse != nil {
		return p.postResponse(resp, s, c)
	}
	return nil
}

func (p *testPlugin) Close() plugins.PluginError {
	if p.close != nil {
		return p.close()
	}
	return nil
}

type mockHTTPClient struct {
	mockResponse *http.Response
	mockErr      error
	requestURL   string
}

func (m *mockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	m.requestURL = req.URL.String()
	return m.mockResponse, m.mockErr
}

func TestGatewayHandler(t *testing.T) {
	testCases := []struct {
		name                       string
		domain                     string
		path                       string
		scheme                     string
		stripPath                  bool
		maxIdleConns               int
		maxIdleTime                int
		timeout                    int
		upstreamTargets            []string
		rewriteMatch               string
		rewriteReplace             string
		requestURL                 string
		expectedStatusCode         int
		expectedHeader             string
		expectedUpstreamRequestURL string
	}{
		{
			name:                       "Test Case 1",
			domain:                     "test.com",
			path:                       "/api",
			scheme:                     "https",
			stripPath:                  true,
			maxIdleConns:               100,
			maxIdleTime:                10,
			timeout:                    5,
			upstreamTargets:            []string{"https://httpbin.org"},
			requestURL:                 "https://test.com/api/anything?test",
			expectedStatusCode:         http.StatusOK,
			expectedHeader:             "plugin",
			expectedUpstreamRequestURL: "https://httpbin.org/anything?test",
		},
		{
			name:                       "Test Case 2 - No matching backend service",
			domain:                     "test.com",
			path:                       "/api",
			scheme:                     "https",
			stripPath:                  true,
			maxIdleConns:               100,
			maxIdleTime:                10,
			timeout:                    5,
			upstreamTargets:            []string{"https://httpbin.org"},
			requestURL:                 "https://test.com/notfound",
			expectedStatusCode:         http.StatusNotFound,
			expectedHeader:             "",
			expectedUpstreamRequestURL: "",
		},
		{
			name:                       "Test Case 3 - Invalid upstream target URL",
			domain:                     "test.com",
			path:                       "/api",
			scheme:                     "https",
			stripPath:                  true,
			maxIdleConns:               100,
			maxIdleTime:                10,
			timeout:                    5,
			upstreamTargets:            []string{"https://httpbin.or"},
			requestURL:                 "https://test.com/api/anythin?test",
			expectedStatusCode:         http.StatusBadGateway,
			expectedHeader:             "plugin",
			expectedUpstreamRequestURL: "https://httpbin.or/anythin?test",
		},
		{
			name:                       "Test Case 5 - Backend service with no domain",
			domain:                     "",
			path:                       "/api",
			scheme:                     "https",
			stripPath:                  true,
			maxIdleConns:               100,
			maxIdleTime:                10,
			timeout:                    5,
			upstreamTargets:            []string{"https://httpbin.org"},
			requestURL:                 "https://localhost/api/anything?test",
			expectedStatusCode:         http.StatusOK,
			expectedHeader:             "plugin",
			expectedUpstreamRequestURL: "https://httpbin.org/anything?test",
		},
		{
			name:                       "Test Case 6 - Backend service with domain",
			domain:                     "test.com",
			path:                       "/api",
			scheme:                     "https",
			stripPath:                  true,
			maxIdleConns:               100,
			maxIdleTime:                10,
			timeout:                    5,
			upstreamTargets:            []string{"https://httpbin.org"},
			requestURL:                 "https://test.com/api/anything?test",
			expectedStatusCode:         http.StatusOK,
			expectedHeader:             "plugin",
			expectedUpstreamRequestURL: "https://httpbin.org/anything?test",
		},
		{
			name:                       "Test Case 7 - StripPath is false",
			domain:                     "",
			path:                       "/api",
			scheme:                     "https",
			stripPath:                  false,
			maxIdleConns:               100,
			maxIdleTime:                10,
			timeout:                    5,
			upstreamTargets:            []string{"https://httpbin.org"},
			requestURL:                 "https://localhost/api/anything/test?test",
			expectedStatusCode:         http.StatusNotFound,
			expectedHeader:             "plugin",
			expectedUpstreamRequestURL: "https://httpbin.org/api/anything/test?test",
		},
		{
			name:                       "Test Case 8 - Multiple backend targets with localhost domain",
			domain:                     "localhost",
			path:                       "/api",
			scheme:                     "http",
			stripPath:                  true,
			maxIdleConns:               100,
			maxIdleTime:                10,
			timeout:                    5,
			upstreamTargets:            []string{"http://localhost:8000", "http://localhost:8001", "http://localhost:8002"},
			requestURL:                 "http://localhost/api/anything?test",
			expectedStatusCode:         http.StatusOK,
			expectedHeader:             "plugin",
			expectedUpstreamRequestURL: "http://localhost:8000/anything?test",
		},
		{
			name:                       "Test Case 9 - Test Root Path /",
			domain:                     "localhost",
			path:                       "/api",
			scheme:                     "http",
			stripPath:                  true,
			maxIdleConns:               100,
			maxIdleTime:                10,
			timeout:                    5,
			upstreamTargets:            []string{"http://localhost:8000", "http://localhost:8001", "http://localhost:8002"},
			requestURL:                 "http://localhost/api",
			expectedStatusCode:         http.StatusOK,
			expectedHeader:             "plugin",
			expectedUpstreamRequestURL: "http://localhost:8000",
		},
		{
			name:                       "Test Case 10 - Query parameters",
			domain:                     "test.com",
			path:                       "/api",
			scheme:                     "https",
			stripPath:                  true,
			maxIdleConns:               100,
			maxIdleTime:                10,
			timeout:                    5,
			upstreamTargets:            []string{"https://httpbin.org"},
			requestURL:                 "https://test.com/api/anything?foo=bar&baz=qux",
			expectedStatusCode:         http.StatusOK,
			expectedHeader:             "plugin",
			expectedUpstreamRequestURL: "https://httpbin.org/anything?foo=bar&baz=qux",
		},
		{
			name:                       "Test Case 11 - URL Rewrite",
			domain:                     "test.com",
			path:                       "/",
			scheme:                     "https",
			stripPath:                  false,
			maxIdleConns:               100,
			maxIdleTime:                10,
			timeout:                    5,
			upstreamTargets:            []string{"https://httpbin.org"},
			rewriteMatch:               "/api/old/(.*)",
			rewriteReplace:             "/api/new/$1",
			requestURL:                 "https://test.com/api/old/anything?test",
			expectedStatusCode:         http.StatusOK,
			expectedHeader:             "plugin",
			expectedUpstreamRequestURL: "https://httpbin.org/api/new/anything?test",
		},
	}

	for _, tc := range testCases {

		// Create a mock HTTP client with the desired response

		bs := &service.BackendService{
			Name:            tc.name,
			Domain:          tc.domain,
			Path:            tc.path,
			Scheme:          tc.scheme,
			StripPath:       tc.stripPath,
			MaxIdleConns:    tc.maxIdleConns,
			MaxIdleTime:     time.Duration(tc.maxIdleTime),
			Timeout:         time.Duration(tc.timeout),
			UpstreamTargets: tc.upstreamTargets,
			LoadBalancerPolicy: service.LoadBalancerPolicy{
				Type: loadbalancer.RoundRobin,
			},
			RewriteMatch:   tc.rewriteMatch,
			RewriteReplace: tc.rewriteReplace,
		}

		bs.Init()

		mockClient := &mockHTTPClient{
			mockResponse: &http.Response{
				StatusCode: tc.expectedStatusCode,
				Header:     make(http.Header),
			},
			mockErr: nil,
		}

		bs.GetHttpClient().Transport = mockClient

		reg, _ := service.NewServiceRegistry(context.Background(), "memory", nil)
		reg.AddService(bs)

		req := httptest.NewRequest("GET", tc.requestURL, nil)
		w := httptest.NewRecorder()

		plugin := &testPlugin{
			preRequest: func(r *http.Request, s service.ServiceRegistry, c *config.Config) plugins.PluginError {
				r.Header.Set("X-Plugin-Header", "plugin")
				return nil
			},
			postResponse: func(resp *http.Response, s service.ServiceRegistry, c *config.Config) plugins.PluginError {
				resp.Header.Set("X-Plugin-Header", "plugin")
				return nil
			},
		}

		logger, err := log.NewZapLogger("info")
		if err != nil {
			t.Errorf("could not create logger due to: %s", err)
		}
		handler := NewAPIGateway(reg, []plugins.FrontmanPlugin{plugin}, &config.Config{}, logger)
		handler.ServeHTTP(w, req)

		// Check the response status code
		if w.Code != tc.expectedStatusCode {
			t.Errorf("[%s] Expected status code %d, got %d", tc.name, tc.expectedStatusCode, w.Code)
		}

		// Check the response headers (plugin and backend)
		if w.Header().Get("X-Plugin-Header") != tc.expectedHeader {
			t.Errorf("[%s] Expected header X-Plugin-Header to be set to '%s', got '%s'", tc.name, tc.expectedHeader, w.Header().Get("X-Plugin-Header"))
		}

		if mockClient.requestURL != tc.expectedUpstreamRequestURL {
			t.Errorf("[%s] Expected upstream request URL to be '%s', got '%s'", tc.name, tc.expectedUpstreamRequestURL, mockClient.requestURL)
		}
	}
}

func TestFindBackendService(t *testing.T) {
	testCases := []struct {
		name           string
		services       []*service.BackendService
		request        *http.Request
		expectedOutput *service.BackendService
	}{
		{
			name: "Matching path with no domain",
			services: []*service.BackendService{
				{
					Domain: "",
					Path:   "/api",
				},
				{
					Domain: "",
					Path:   "/admin",
				},
			},
			request: &http.Request{
				URL: &url.URL{
					Path: "/api/get",
				},
			},
			expectedOutput: &service.BackendService{
				Domain: "",
				Path:   "/api",
			},
		},
		{
			name: "Matching path and domain",
			services: []*service.BackendService{
				{
					Domain: "example.com",
					Path:   "/api",
				},
				{
					Domain: "example.com",
					Path:   "/admin",
				},
			},
			request: &http.Request{
				Host: "example.com",
				URL: &url.URL{
					Path: "/api/get",
				},
			},
			expectedOutput: &service.BackendService{
				Domain: "example.com",
				Path:   "/api",
			},
		},
		{
			name: "No matching backend service",
			services: []*service.BackendService{
				{
					Domain: "",
					Path:   "/api",
				},
				{
					Domain: "example.com",
					Path:   "/admin",
				},
			},
			request: &http.Request{
				URL: &url.URL{
					Path: "/dashboard",
				},
			},
			expectedOutput: nil,
		},
	}

	reg, _ := service.NewServiceRegistry(context.Background(), "memory", nil)

	for _, tc := range testCases {
		for _, s := range tc.services {
			reg.AddService(s)
		}

		t.Run(tc.name, func(t *testing.T) {
			reg.GetTrie().BuildRoutes(tc.services)
			result := reg.GetTrie().FindBackendService(tc.request)
			if !reflect.DeepEqual(result, tc.expectedOutput) {
				t.Errorf("Unexpected result: got %#v, want %#v", result, tc.expectedOutput)
			}
		})
	}
}

func BenchmarkGatewayHandler(b *testing.B) {
	bs := &service.BackendService{
		Name:            "test",
		Domain:          "test.com",
		Path:            "/api",
		Scheme:          "https",
		StripPath:       true,
		MaxIdleConns:    100,
		MaxIdleTime:     time.Duration(10) * time.Second,
		Timeout:         time.Duration(5) * time.Second,
		UpstreamTargets: []string{"httpbin.org"},
	}

	bs.Init()

	reg, _ := service.NewServiceRegistry(context.Background(), "memory", nil)
	reg.AddService(bs)

	req, _ := http.NewRequest("GET", "https://test.com/api/anything?test", nil)
	w := httptest.NewRecorder()

	plugin := &testPlugin{
		preRequest: func(r *http.Request, s service.ServiceRegistry, c *config.Config) plugins.PluginError {
			r.Header.Set("X-Plugin-Header", "plugin")
			return nil
		},
		postResponse: func(resp *http.Response, s service.ServiceRegistry, c *config.Config) plugins.PluginError {
			resp.Header.Set("X-Plugin-Header", "plugin")
			return nil
		},
	}

	logger, err := log.NewZapLogger("info")
	if err != nil {
		b.Errorf("could not create logger due to: %s", err)
	}
	handler := NewAPIGateway(reg, []plugins.FrontmanPlugin{plugin}, &config.Config{}, logger)

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
	}
}
