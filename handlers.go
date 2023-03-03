package frontman

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

func refreshConnections(bs *BackendServices, clients map[string]*http.Client, clientLock *sync.Mutex) {
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
			}
		}
	}
}

func gatewayHandler(bs *BackendServices) http.HandlerFunc {
	var clients map[string]*http.Client = make(map[string]*http.Client)
	var clientLock sync.Mutex
	var currentTargetIndex int
	go refreshConnections(bs, clients, &clientLock)

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

		// Get the target index to use for this request
		targetIndex := getNextTargetIndex(backendService, currentTargetIndex)
		currentTargetIndex = targetIndex

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

		// Get or create a new client for this service

		clientLock.Lock()
		client, ok := clients[upstreamTarget]
		clientLock.Unlock()
		if !ok {
			http.Error(w, "Client not found", http.StatusInternalServerError)
			return
		}

		// Copy the headers from the original request
		headers := make(http.Header)
		for k, v := range r.Header {
			headers[k] = v
		}

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

		// Log a message indicating that the response has been received from the target service
		log.Printf("Response received from %s: %d %s", upstreamTarget, resp.StatusCode, resp.Status)

		// Copy the response headers back to the client
		for k, v := range resp.Header {
			w.Header()[k] = v
		}

		// Set the status code and body of the response
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
