package frontman

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func getServicesHandler(bs *BackendServices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		services := bs.GetServices()
		jsonData, err := json.Marshal(services)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
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
		if service.Path == "" {
			http.Error(w, "Path is a required field", http.StatusBadRequest)
			return
		}

		// If the scheme is not specified, default to "http"
		if service.Scheme == "" {
			service.Scheme = "http"
		}

		// Check that at least one upstream target is specified
		if len(service.UpstreamTargets) < 1 {
			http.Error(w, "At least one upstream target is required", http.StatusBadRequest)
			return
		}

		// If no timeout is specified, default to 10 seconds
		if service.Timeout == 0 {
			service.Timeout = 10 
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
		if service.UpstreamTargets == nil || len(service.UpstreamTargets) == 0 || service.Path == "" {
			http.Error(w, "UpstreamTargets and path are required fields", http.StatusBadRequest)
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
