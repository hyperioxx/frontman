package frontman

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/hyperioxx/frontman/config"
	"github.com/hyperioxx/frontman/plugins"
	"github.com/hyperioxx/frontman/service"
)

func refreshClients(bs *service.BackendService, clients map[string]*http.Client, clientLock *sync.Mutex) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			clientLock.Lock()

			// Update the transport settings for each client
			for _, client := range clients {
				transport := client.Transport.(*http.Transport)
				transport.MaxIdleConns = bs.MaxIdleConns
				transport.IdleConnTimeout = bs.MaxIdleTime * time.Second
				transport.TLSHandshakeTimeout = bs.Timeout * time.Second
			}

			clientLock.Unlock()
		}
	}
}

func refreshConnections(bs service.ServiceRegistry, clients map[string]*http.Client, clientLock *sync.Mutex) {

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			services := bs.GetServices()

			// Remove clients that are no longer needed
			clientLock.Lock()
			for url := range clients {
				found := false
				for _, s := range services {
					for _, t := range s.UpstreamTargets {
						if t == url {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					delete(clients, url)
				}
			}
			clientLock.Unlock()

			// Add or update clients for each service
			for _, s := range services {
				for _, t := range s.UpstreamTargets {
					clientLock.Lock()
					_, ok := clients[t]
					if !ok {
						transport := &http.Transport{
							MaxIdleConns:        s.MaxIdleConns,
							IdleConnTimeout:     s.MaxIdleTime * time.Second,
							TLSHandshakeTimeout: s.Timeout * time.Second,
						}
						client := &http.Client{
							Transport: transport,
						}
						clients[t] = client
					} else {
						clients[t].Transport.(*http.Transport).MaxIdleConns = s.MaxIdleConns
						clients[t].Transport.(*http.Transport).IdleConnTimeout = s.MaxIdleTime * time.Second
						clients[t].Transport.(*http.Transport).TLSHandshakeTimeout = s.Timeout * time.Second
					}
					clientLock.Unlock()
				}
				refreshClients(s, clients, clientLock)
			}
		}
	}
}

func findBackendService(services []*service.BackendService, r *http.Request) *service.BackendService {
	for _, s := range services {
		if s.Domain == "" && strings.HasPrefix(r.URL.Path, s.Path) {
			return s
		}
		if s.Domain != "" && r.Host == s.Domain && strings.HasPrefix(r.URL.Path, s.Path) {
			return s
		}
	}
	return nil
}

func getNextTargetIndex(backendService *service.BackendService, currentIndex int) int {
	numTargets := len(backendService.UpstreamTargets)
	if numTargets == 0 {
		return -1
	}
	if currentIndex >= numTargets-1 {
		return 0
	}
	return currentIndex + 1
}

func getClientForBackendService(bs service.BackendService, target string, clients map[string]*http.Client, clientLock *sync.Mutex) (*http.Client, error) {
	clientLock.Lock()
	defer clientLock.Unlock()

	// Check if the client for this target already exists
	if client, ok := clients[target]; ok {
		return client, nil
	}

	// Create a new transport with the specified settings
	transport := &http.Transport{
		MaxIdleConns:        bs.MaxIdleConns,
		IdleConnTimeout:     bs.MaxIdleTime * time.Second,
		TLSHandshakeTimeout: bs.Timeout * time.Second,
	}

	// Create a new HTTP client with the transport
	client := &http.Client{
		Transport: transport,
	}

	// Add the client to the map of clients
	clients[target] = client

	return client, nil
}

func copyHeaders(dst, src http.Header) {
	for k, v := range src {
		dst[k] = v
	}
}

func gatewayHandler(bs service.ServiceRegistry, plugs []plugins.FrontmanPlugin, conf *config.Config) http.HandlerFunc {
	// Create a map to store HTTP clients for each backend service
	var clients map[string]*http.Client = make(map[string]*http.Client)
	var clientLock sync.Mutex
	var currentTargetIndex int

	// Start a goroutine to refresh HTTP connections to each backend service
	go refreshConnections(bs, clients, &clientLock)

	return func(w http.ResponseWriter, r *http.Request) {
		for _, plugin := range plugs {
			if err := plugin.PreRequest(r, bs, conf); err != nil {
				log.Printf("Plugin error: %v", err)
				http.Error(w, err.Error(), err.StatusCode())
				return
			}
		}
		// Find the backend service that matches the request
		backendService := findBackendService(bs.GetServices(), r)

		// If the backend service was not found, return a 404 error
		if backendService == nil {
			http.NotFound(w, r)
			return
		}

		// Get the target index to use for this request
		targetIndex := getNextTargetIndex(backendService, currentTargetIndex)

		// Get the upstream target URL for this request
		upstreamTarget := backendService.UpstreamTargets[targetIndex]

		// Create a new target URL with the service path and scheme
		targetURL, err := url.Parse(backendService.Scheme + "://" + upstreamTarget + backendService.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Strip the service path from the request URL if required
		if backendService.StripPath {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, backendService.Path)
		}

		// Get or create a new client for this backend service
		client, err := getClientForBackendService(*backendService, upstreamTarget, clients, &clientLock)
		headers := make(http.Header)
		// Copy the headers from the original request
		copyHeaders(headers, r.Header)

		// Remove the X-Forwarded-For header to prevent spoofing
		headers.Del("X-Forwarded-For")

		// Log a message indicating that the request is being sent to the target service
		log.Printf("Sending request to %s: %s %s", upstreamTarget, r.Method, r.URL.Path)

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

		for _, plugin := range plugs {
			if err := plugin.PostResponse(resp, bs, conf); err != nil {
				log.Printf("Plugin error: %v", err)
				http.Error(w, err.Error(), err.StatusCode())
				return
			}
		}

		// Log a message indicating that the response has been received from the target service
		log.Printf("Response received from %s: %d %s", upstreamTarget, resp.StatusCode, resp.Status)

		// Copy the response headers back to the client
		copyHeaders(w.Header(), resp.Header)

		// Set the status code and body of the response
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
