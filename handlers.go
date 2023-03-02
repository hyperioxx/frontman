package frontman

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gorilla/mux"
)


func getServicesHandler(bs *BackendServices) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        services := bs.GetServices()
        for _, service := range services {
            w.Write([]byte(service.Name + ": " + service.URL + "\n"))
        }
    }
}

func getHealthHandler(bs *BackendServices) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        services := bs.GetServices()
        healthStatus := make(map[string]bool)
        for _, service := range services {
            healthStatus[service.Name] = service.GetHealthCheck()
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(healthStatus)
    }
}

func addServiceHandler(bs *BackendServices) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse the request body as a BackendService object
        var service BackendService
        err := json.NewDecoder(r.Body).Decode(&service)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Validate that the required fields are present
        if service.URL == "" || service.Path == "" {
            http.Error(w, "URL and path are required fields", http.StatusBadRequest)
            return
        }

        // If the scheme is not specified, default to "http"
        if service.Scheme == "" {
            service.Scheme = "http"
        }
 

        // Add the service to the list of backend services
        bs.AddService(&service)
	
        // Write a response to the HTTP client indicating that the service was added successfully
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(service)
    }
}

func updateServiceHandler(bs *BackendServices) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse the request body as a BackendService object
        var service BackendService
        err := json.NewDecoder(r.Body).Decode(&service)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Validate that the required fields are present
        if service.URL == "" || service.Path == "" {
            http.Error(w, "URL and path are required fields", http.StatusBadRequest)
            return
        }

        // If the scheme is not specified, default to "http"
        if service.Scheme == "" {
            service.Scheme = "http"
        }

        // Update the service in the list of backend services
        err = bs.UpdateService(&service)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Write a response to the HTTP client indicating that the service was updated successfully
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(service)
    }
}



func removeServiceHandler(bs *BackendServices) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        name := mux.Vars(r)["name"]
        bs.RemoveService(name) 
        w.Write([]byte("Removed service " + name + "\n"))
    }
}


func forwardRequest(w http.ResponseWriter, r *http.Request, targetURL *url.URL) {
    targetURL.Path = path.Join(targetURL.Path, r.URL.Path)

    // Copy the headers from the original request
    headers := make(http.Header)
    for k, v := range r.Header {
        headers[k] = v
    }

    // Remove the X-Forwarded-For header to prevent spoofing
    headers.Del("X-Forwarded-For")

    // Log a message indicating that the request is being sent to the target service
    log.Printf("Sending request to %s: %s %s", targetURL.Host, r.Method, targetURL.Path)

    // Send the request to the target service
    resp, err := http.DefaultClient.Do(&http.Request{
        Method:        r.Method,
        URL:           targetURL,
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
    log.Printf("Response received from %s: %d %s", targetURL.Host, resp.StatusCode, resp.Status)

    // Copy the response headers back to the client
    for k, v := range resp.Header {
        w.Header()[k] = v
    }

    // Set the status code and body of the response
    w.WriteHeader(resp.StatusCode)
    io.Copy(w, resp.Body)
}


func reverseProxyHandler(bs *BackendServices) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get the service name from the request path

		services := bs.GetServices()
    
        // Find the backend service that matches the domain and path
        var backendService *BackendService
        for _, s := range services {
	        fmt.Println(services)
            if s.Domain != "" && r.Host == s.Domain && strings.HasPrefix(r.URL.Path, s.Path) {
				fmt.Println("TEST2")
                backendService = s
                break
            }
            if s.Domain == "" && strings.HasPrefix(r.URL.Path, s.Path) {
				fmt.Println("TEST3")
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
		fmt.Println(backendService)
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
        handler := http.StripPrefix("/"+ backendService.Path , http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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




