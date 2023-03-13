package frontman

import (
	"encoding/json"
	"fmt"
	"github.com/Frontman-Labs/frontman/loadbalancer"
	"net/http"
	"net/url"

	"github.com/Frontman-Labs/frontman/service"
	"github.com/gorilla/mux"
)

func getServicesHandler(bs service.ServiceRegistry) http.HandlerFunc {
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

func getHealthHandler(bs service.ServiceRegistry) http.HandlerFunc {
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

func addServiceHandler(bs service.ServiceRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body as a BackendService object
		var service service.BackendService
		err := json.NewDecoder(r.Body).Decode(&service)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate service
		err = validateService(&service)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Add the service to the list of backend services
		bs.AddService(&service)

		// Write a response to the HTTP client indicating that the service was added successfully
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(service)
	}
}

func updateServiceHandler(bs service.ServiceRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body as a BackendService object
		var service service.BackendService
		err := json.NewDecoder(r.Body).Decode(&service)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = validateService(&service)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
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

func removeServiceHandler(bs service.ServiceRegistry) http.HandlerFunc {
	type Response struct {
		Message string `json:"message,omitempty"`
		Error   string `json:"error,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		err := bs.RemoveService(name)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Message: "",
				Error:   "missing service name",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Message: "Removed service " + name,
			Error:   "",
		})
	}
}

func validateService(service *service.BackendService) error {
	// Validate that the required fields are present
	if service.Path == "" {
		return fmt.Errorf("path is a required field")
	}

	// Validate that at least one upstream target is specified and that each target is a valid URL
	if len(service.UpstreamTargets) < 1 {
		return fmt.Errorf("at least one upstream target is required")
	}
	for _, target := range service.UpstreamTargets {
		u, err := url.Parse(target)
		if err != nil {
			return fmt.Errorf("Invalid upstream target: " + target)
		}
		if u.Scheme == "" {
			return fmt.Errorf("Upstream target " + target + " must include a scheme (e.g., 'http' or 'https')")
		}
	}

	// If the scheme is not specified, default to "http"
	if service.Scheme == "" {
		service.Scheme = "http"
	}

	// If no timeout is specified, default to 10 seconds
	if service.Timeout == 0 {
		service.Timeout = 10
	}

	// If no policy type is specified, default to round-robin
	if service.LoadBalancerPolicy.Type == "" {
		service.LoadBalancerPolicy.Type = loadbalancer.RoundRobin
	}

	// Validate load-balancer policy and set
	err := validateLoadBalancerPolicy(service)
	if err != nil {
		return err
	}

	service.Init()

	return nil
}

func validateLoadBalancerPolicy(s *service.BackendService) error {
	switch s.LoadBalancerPolicy.Type {
	case loadbalancer.RoundRobin:
	case loadbalancer.WeightedRoundRobin:
		if len(s.LoadBalancerPolicy.Options.Weights) != len(s.UpstreamTargets) {
			return fmt.Errorf("mismatched lengts of weights and targets")
		}

		for _, w := range s.LoadBalancerPolicy.Options.Weights {
			if w <= 0 {
				return fmt.Errorf("weightes must be greater than zero")
			}
		}
	default:
		return fmt.Errorf("unknown load-balancer policy: %s", s.LoadBalancerPolicy.Type)
	}

	return nil
}
