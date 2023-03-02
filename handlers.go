package frontman

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"

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

func addServiceHandler(bs *BackendServices) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse the request body as a BackendService object
        var service BackendService
        err := json.NewDecoder(r.Body).Decode(&service)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Add the service to the list of backend services
        bs.AddService(&service)

        // Write a response to the HTTP client indicating that the service was added successfully
        w.WriteHeader(http.StatusCreated)
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
        vars := mux.Vars(r)
        service := vars["service"]

        // Get the current list of backend services
        services := bs.GetServices()

        // Find the backend service by name
        var backendService *BackendService
        for _, s := range services {
            if s.Name == service {
                backendService = s
                break
            }
        }

        // If the backend service was not found, return a 404 error
        if backendService == nil {
            http.NotFound(w, r)
            return
        }

        // Create a new target URL with the service name removed from the path
        targetURL, err := url.Parse("http://" + backendService.URL)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Create a new handler that strips the service name from the path
        handler := http.StripPrefix("/"+service, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            forwardRequest(w, r, targetURL)
        }))

        // Serve the request with the modified handler
        handler.ServeHTTP(w, r)
    }
}


