package service

import (
	"context"
	"encoding/json"
	"errors"

	"sync"

	"github.com/go-redis/redis/v9"
)

func NewRedisClient(ctx context.Context, uri string) (*redis.Client, error) {
	opt, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// RedisRegistry implements the ServiceRegistry interface using Redis as a backend storage
type RedisRegistry struct {
	redisClient *redis.Client
	namespace   string
	ctx         context.Context
	services    map[string]*BackendService
	sync.RWMutex
}

// NewRedisRegistry creates a new RedisRegistry instance with a Redis client connection
func NewRedisRegistry(ctx context.Context, redisClient *redis.Client, namespace string) (*RedisRegistry, error) {
	r := &RedisRegistry{redisClient: redisClient, ctx: ctx, namespace: namespace}
	if err := r.loadServices(); err != nil {
		return nil, err
	}
	return r, nil
}

// ensureListExists creates the services list if it does not exist
func (r *RedisRegistry) ensureListExists() error {
	return r.redisClient.Do(r.ctx, "PING").Err()
}

// loadServices retrieves the list of backend services from Redis
func (r *RedisRegistry) loadServices() error {
	services, err := r.redisClient.LRange(r.ctx, "services", 0, -1).Result()
	if err != nil {
		return err
	}

	r.Lock()
	defer r.Unlock()
	r.services = make(map[string]*BackendService)
	for _, service := range services {
		var backendService BackendService
		err := json.Unmarshal([]byte(service), &backendService)
		if err != nil {
			return err
		}
		backendService.Init()
		r.services[backendService.Name] = &backendService
	}
	return nil
}

// AddService adds a new backend service to Redis
func (r *RedisRegistry) AddService(service *BackendService) error {
	r.Lock()
	defer r.Unlock()

	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return err
	}

	_, err = r.redisClient.RPush(r.ctx, "services", serviceJSON).Result()
	if err != nil {
		return err
	}
	r.services[service.Name] = service

	return nil
}

// UpdateService updates an existing backend service in Redis
func (r *RedisRegistry) UpdateService(service *BackendService) error {
	r.Lock()
	defer r.Unlock()

	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return err
	}

	if _, ok := r.services[service.Name]; !ok {
		return errors.New("service not found")
	}

	r.services[service.Name] = service
	err = r.redisClient.LSet(r.ctx, "services", int64(len(r.services)-1), serviceJSON).Err()
	if err != nil {
		return err
	}

	return nil
}

// RemoveService removes a backend service from Redis
func (r *RedisRegistry) RemoveService(name string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.services[name]; !ok {
		return errors.New("service not found")
	}

	delete(r.services, name)
	err := r.redisClient.LRem(r.ctx, "services", 0, name).Err()
	if err != nil {
		return err
	}

	return nil
}

// GetServices returns a copy of the current list of backend services
func (r *RedisRegistry) GetServices() []*BackendService {
	r.RLock()
	defer r.RUnlock()
	services := make([]*BackendService, 0, len(r.services))
	for _, service := range r.services {
		services = append(services, service)
	}
	return services
}
