package frontman

import (
	"net/http"
	"net/http/httputil"
	"net/url"

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
        name := r.FormValue("name")
        url := r.FormValue("url")
        service := &BackendService{
            Name: name,
            URL:  url,
        }
        bs.AddService(service)
        w.Write([]byte("Added service " + name + ": " + url + "\n"))
    }
}

func removeServiceHandler(bs *BackendServices) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        name := mux.Vars(r)["name"]
        bs.RemoveService(name) 
        w.Write([]byte("Removed service " + name + "\n"))
    }
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

        // Create a new reverse proxy instance for the backend service
        proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: backendService.URL})

        // Update the request URL and headers
        r.URL.Path = "/" + vars["path"]
        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
        r.Host = backendService.URL

        // Serve the reverse proxy request
        proxy.ServeHTTP(w, r)
    }
}