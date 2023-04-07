package api

import (
	"encoding/json"
	"fmt"
	"github.com/Frontman-Labs/frontman/loadbalancer"
	"net/http"
	"net/url"

	"github.com/Frontman-Labs/frontman/service"
	"github.com/julienschmidt/httprouter"
)

func NewServicesRouter(backendServices service.ServiceRegistry) *httprouter.Router {
	router := httprouter.New()

	router.GET("/api/services", getServicesHandler(backendServices))
	router.POST("/api/services", addServiceHandler(backendServices))
	router.DELETE("/api/services/:name", removeServiceHandler(backendServices))
	router.PUT("/api/services/:name", updateServiceHandler(backendServices))
	router.GET("/api/health", getHealthHandler(backendServices))

	return router
}

func getServicesHandler(bs service.ServiceRegistry) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		services := bs.GetServices()
		jsonData, err := json.Marshal(services)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		prepareHeaders(w, http.StatusOK)
		w.Write(jsonData)
	}
}

func getHealthHandler(bs service.ServiceRegistry) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		services := bs.GetServices()
		healthStatus := make(map[string]bool)
		for _, service := range services {
			healthStatus[service.Name] = service.GetHealthCheck()
		}

		prepareHeaders(w, http.StatusOK)
		json.NewEncoder(w).Encode(healthStatus)
	}
}

func addServiceHandler(bs service.ServiceRegistry) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		err = bs.AddService(&service)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write a response to the HTTP client indicating that the service was added successfully
		prepareHeaders(w, http.StatusCreated)
		json.NewEncoder(w).Encode(service)
	}
}

func updateServiceHandler(bs service.ServiceRegistry) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		name := params.ByName("name")
		// Parse the request body as a BackendService object
		var service service.BackendService
		err := json.NewDecoder(r.Body).Decode(&service)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		service.Name = name

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
		prepareHeaders(w, http.StatusOK)
		json.NewEncoder(w).Encode(service)
	}
}

func removeServiceHandler(bs service.ServiceRegistry) httprouter.Handle {
	type Response struct {
		Message string `json:"message,omitempty"`
		Error   string `json:"error,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		name := params.ByName("name")
		err := bs.RemoveService(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		prepareHeaders(w, http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Message: "Removed service " + name,
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
	case loadbalancer.Random:
	case loadbalancer.LeastConnection:
	case loadbalancer.RoundRobin:
	case loadbalancer.WeightedRoundRobin, loadbalancer.WeightedLeastConnection:
		if len(s.LoadBalancerPolicy.Options.Weights) != len(s.UpstreamTargets) {
			return fmt.Errorf("mismatched lengths of weights and targets")
		}

		for _, w := range s.LoadBalancerPolicy.Options.Weights {
			if w <= 0 {
				return fmt.Errorf("weights must be greater than zero")
			}
		}
	default:
		return fmt.Errorf("unknown load-balancer policy: %s", s.LoadBalancerPolicy.Type)
	}

	return nil
}

func prepareHeaders(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
}
