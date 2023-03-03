package frontman

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func reverseProxyHandler(bs *BackendServices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the service name from the request path

		services := bs.GetServices()

		// Find the backend service that matches the domain and path
		var backendService *BackendService
		for _, s := range services {
			if s.Domain != "" && r.Host == s.Domain && strings.HasPrefix(r.URL.Path, s.Path) {
				backendService = s
				break
			}
			if s.Domain == "" && strings.HasPrefix(r.URL.Path, s.Path) {
				backendService = s
				break
			}
		}

		// If the backend service was not found, return a 404 error
		if backendService == nil {
			http.NotFound(w, r)
			return
		}

		// Create a new target URL with the service path and scheme

		targetURL, err := url.Parse(backendService.Scheme + "://" + backendService.URL + backendService.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create a new transport with the specified options
		transport := &http.Transport{
			MaxIdleConns:        backendService.MaxIdleConns,
			IdleConnTimeout:     backendService.MaxIdleTime,
			TLSHandshakeTimeout: backendService.Timeout,
		}

		// Create a new client with the specified transport
		client := &http.Client{
			Transport: transport,
		}

		// Create a new handler that strips the service path from the request URL
		handler := http.StripPrefix(backendService.Path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Copy the headers from the original request
			headers := make(http.Header)
			for k, v := range r.Header {
				headers[k] = v
			}

			// Remove the X-Forwarded-For header to prevent spoofing
			headers.Del("X-Forwarded-For")

			// Log a message indicating that the request is being sent to the target service
			log.Printf("Sending request to %s: %s %s", backendService.URL, r.Method, r.URL.Path)

			// Send the request to the target service using the client with the specified transport
			resp, err := client.Do(&http.Request{
				Method:        r.Method,
				URL:           targetURL.ResolveReference(r.URL),
				Proto:         r.Proto,
				ProtoMajor:    r.ProtoMajor,
				ProtoMinor:    r.ProtoMinor,
				Header:        headers,
				Body:          r.Body,
				ContentLength: r.ContentLength,
				Host:          targetURL.Host,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}

			defer resp.Body.Close()

			// Log a message indicating that the response has been received from the target service
			log.Printf("Response received from %s: %d %s", backendService.URL, resp.StatusCode, resp.Status)

			// Copy the response headers back to the client
			for k, v := range resp.Header {
				w.Header()[k] = v
			}

			// Set the status code and body of the response
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
		}))

		// Serve the request with the modified handler
		handler.ServeHTTP(w, r)
	}
}
