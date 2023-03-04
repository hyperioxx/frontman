package frontman

import (
	"context"
	"encoding/json"
	"errors"

	"log"
	"net/http"

	"sync"
	"time"

	"github.com/go-redis/redis/v9"
)

// BackendService holds the details of a backend service
type BackendService struct {
	Name            string        `json:"name"`
	Scheme          string        `json:"scheme"`
	UpstreamTargets []string      `json:"upstreamTargets"`
	Path            string        `json:"path"`
	Domain          string        `json:"domain"`
	HealthCheck     string        `json:"healthCheck"`
	RetryAttempts   int           `json:"retryAttempts"`
	Timeout         time.Duration `json:"timeout"`
	MaxIdleConns    int           `json:"maxIdleConns"`
	MaxIdleTime     time.Duration `json:"maxIdleTime"`
	StripPath       bool          `json:"stripPath"`
	Provider        OAuthProvider
}

// BackendServices holds a map of backend services
type BackendServices struct {
	redis    *redis.Client
	ctx      context.Context
	services map[string]*BackendService
	sync.RWMutex
}

// NewBackendServices creates a new BackendServices instance with a Redis client connection
func NewBackendServices(ctx context.Context, redisClient *redis.Client) (*BackendServices, error) {
	bs := &BackendServices{redis: redisClient, ctx: ctx}
	if err := bs.loadServices(); err != nil {
		return nil, err
	}
	return bs, nil
}

// ensureListExists creates the services list if it does not exist
func (bs *BackendServices) ensureListExists() error {
	return bs.redis.Do(bs.ctx, "PING").Err()
}

// loadServices retrieves the list of backend services from Redis
func (bs *BackendServices) loadServices() error {
	services, err := bs.redis.LRange(bs.ctx, "services", 0, -1).Result()
	if err != nil {
		return err
	}

	bs.Lock()
	defer bs.Unlock()
	bs.services = make(map[string]*BackendService)
	for _, service := range services {
		var backendService BackendService
		err := json.Unmarshal([]byte(service), &backendService)
		if err != nil {
			return err
		}
		bs.services[backendService.Name] = &backendService
	}
	return nil
}

// AddService adds a new backend service to Redis
func (bs *BackendServices) AddService(service *BackendService) error {
	bs.Lock()
	defer bs.Unlock()

	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return err
	}

	_, err = bs.redis.RPush(bs.ctx, "services", serviceJSON).Result()
	if err != nil {
		return err
	}
	bs.services[service.Name] = service

	return nil
}

// UpdateService updates an existing backend service in Redis
func (bs *BackendServices) UpdateService(service *BackendService) error {
	bs.Lock()
	defer bs.Unlock()

	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return err
	}

	if _, ok := bs.services[service.Name]; !ok {
		return errors.New("service not found")
	}

	bs.services[service.Name] = service
	err = bs.redis.LSet(bs.ctx, "services", int64(len(bs.services)-1), serviceJSON).Err()
	if err != nil {
		return err
	}

	return nil
}

// RemoveService removes a backend service from Redis
func (bs *BackendServices) RemoveService(name string) error {
	bs.Lock()
	defer bs.Unlock()

	if _, ok := bs.services[name]; !ok {
		return errors.New("service not found")
	}

	delete(bs.services, name)
	err := bs.redis.LRem(bs.ctx, "services", 0, name).Err()
	if err != nil {
		return err
	}

	return nil
}

// GetServices returns a copy of the current list of backend services
func (bs *BackendServices) GetServices() []*BackendService {
	bs.RLock()
	defer bs.RUnlock()
	services := make([]*BackendService, 0, len(bs.services))
	for _, service := range bs.services {
		services = append(services, service)
	}
	return services
}

// GetHealthCheck performs a health check on the backend service and returns true if it is healthy.
func (bs *BackendService) GetHealthCheck() bool {
	resp, err := http.Get(bs.HealthCheck)
	if err != nil {
		log.Printf("Error performing health check for service %s: %s", bs.Name, err.Error())
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return true
	}

	log.Printf("Service %s health check failed with status code %d", bs.Name, resp.StatusCode)
	return false
}
