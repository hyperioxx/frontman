package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Frontman-Labs/frontman/config"
	"github.com/Frontman-Labs/frontman/log"
	"github.com/Frontman-Labs/frontman/plugins"
	"github.com/Frontman-Labs/frontman/service"
)

type APIGateway struct {
	reg        service.ServiceRegistry
	plugs      []plugins.FrontmanPlugin
	conf       *config.Config
	clients    map[string]*http.Client
	clientLock *sync.Mutex
	log        log.Logger
}

func NewAPIGateway(bs service.ServiceRegistry, plugs []plugins.FrontmanPlugin, conf *config.Config, clients map[string]*http.Client, logger log.Logger, lock *sync.Mutex) *APIGateway {
	return &APIGateway{
		reg:        bs,
		plugs:      plugs,
		conf:       conf,
		clients:    clients,
		clientLock: lock,
		log:        logger,
	}
}

func (g *APIGateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, plugin := range g.plugs {
		if err := plugin.PreRequest(req, g.reg, g.conf); err != nil {
			g.log.Errorf("Plugin error: %v", err)
			http.Error(w, err.Error(), err.StatusCode())
			return
		}
	}

	// Find the backend service that matches the request
	backendService := g.reg.GetTrie().FindBackendService(req)

	// If the backend service was not found, return a 404 error
	if backendService == nil {
		http.NotFound(w, req)
		return
	}

	// Get the upstream target URL for this request
	upstreamTarget := backendService.GetLoadBalancer().ChooseTarget(backendService.UpstreamTargets)

	var urlPath string

	if backendService.StripPath {
		urlPath = strings.TrimPrefix(req.URL.Path, backendService.Path)
	} else {
		urlPath = backendService.Path
	}

	// Create a new target URL with the service path and scheme
    targetURL, err := url.Parse(upstreamTarget + urlPath + "?" + req.URL.RawQuery)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	// Get or create a new client for this backend service
	client, err := getClientForBackendService(*backendService, backendService.Name, g.clients, g.clientLock)

	// Copy the headers from the original request
	headers := make(http.Header)
	copyHeaders(headers, req.Header)

	if backendService.AuthConfig != nil {
		tokenValidator := backendService.GetTokenValidator()
		// Backend service has auth config specified
		claims, err := tokenValidator.ValidateToken(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if claims != nil {
			data, err := json.Marshal(claims)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			headers.Add(backendService.GetUserDataHeader(), string(data))
		}

	}
	// Remove the X-Forwarded-For header to prevent spoofing
	headers.Del("X-Forwarded-For")

	// Log a message indicating that the request is being sent to the target service
	g.log.Infof("Sending request to %s: %s %s", upstreamTarget, req.Method, urlPath)

	// Send the request to the target service using the client with the specified transport
	resp, err := client.Do(&http.Request{
		Method:        req.Method,
		URL:           targetURL,
		Proto:         req.Proto,
		ProtoMajor:    req.ProtoMajor,
		ProtoMinor:    req.ProtoMinor,
		Header:        headers,
		Body:          req.Body,
		ContentLength: req.ContentLength,
		Host:          targetURL.Host,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		g.log.Infof("Error sending request: %v\n", err.Error())
		return
	}

	backendService.GetLoadBalancer().Done(upstreamTarget)

	defer resp.Body.Close()

	for _, plugin := range g.plugs {
		if err := plugin.PostResponse(resp, g.reg, g.conf); err != nil {
			g.log.Infof("Plugin error: %v", err)
			http.Error(w, err.Error(), err.StatusCode())
			return
		}
	}

	// Log a message indicating that the response has been received from the target service
	g.log.Infof("Response received from %s: %d %s", upstreamTarget, resp.StatusCode, resp.Status)

	// Copy the response headers back to the client
	copyHeaders(w.Header(), resp.Header)

	// Set the status code and body of the response
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

}

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

func RefreshConnections(bs service.ServiceRegistry, clients map[string]*http.Client, clientLock *sync.Mutex) {

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			services := bs.GetServices()

			// Remove clients that are no longer needed
			clientLock.Lock()
			for k := range clients {
				found := false
				for _, s := range services {
					for _, t := range s.UpstreamTargets {
						key := fmt.Sprintf("%s_%s", s.Name, t)
						if key == k {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					delete(clients, k)
				}
			}
			clientLock.Unlock()

			// Add or update clients for each service
			for _, s := range services {
				for _, t := range s.UpstreamTargets {
					clientLock.Lock()
					key := fmt.Sprintf("%s_%s", s.Name, t)
					_, ok := clients[key]
					if !ok {
						transport := &http.Transport{
							MaxIdleConns:        s.MaxIdleConns,
							IdleConnTimeout:     s.MaxIdleTime * time.Second,
							TLSHandshakeTimeout: s.Timeout * time.Second,
						}
						client := &http.Client{
							Transport: transport,
						}
						clients[key] = client
					} else {
						clients[key].Transport.(*http.Transport).MaxIdleConns = s.MaxIdleConns
						clients[key].Transport.(*http.Transport).IdleConnTimeout = s.MaxIdleTime * time.Second
						clients[key].Transport.(*http.Transport).TLSHandshakeTimeout = s.Timeout * time.Second
					}
					clientLock.Unlock()
				}
				refreshClients(s, clients, clientLock)
			}
			
		}
	}
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
	key := fmt.Sprintf("%s_%s", bs.Name, target)
	clients[key] = client

	return client, nil
}

func copyHeaders(dst, src http.Header) {
	for k, v := range src {
		dst[k] = v
	}
}
