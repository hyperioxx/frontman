package service

import (
	"context"
	"encoding/json"
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
	baseRegistry
	redisClient *redis.Client
	namespace   string
	ctx         context.Context
}

// NewRedisRegistry creates a new RedisRegistry instance with a Redis client connection
func NewRedisRegistry(ctx context.Context, redisClient *redis.Client, namespace string) (*RedisRegistry, error) {
	r := &RedisRegistry{redisClient: redisClient, ctx: ctx, namespace: namespace}
	if err := r.loadServices(); err != nil {
		return nil, err
	}

	return r, nil
}

// AddService adds a new backend service to Redis
func (r *RedisRegistry) AddService(service *BackendService) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return err
	}

	err = r.addService(service, func() error {
		_, err = r.redisClient.RPush(r.ctx, "services", serviceJSON).Result()
		return err
	})

	return err
}

// UpdateService updates an existing backend service in Redis
func (r *RedisRegistry) UpdateService(service *BackendService) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return err
	}

	err = r.updateService(service, func() error {
		return r.redisClient.LSet(r.ctx, "services", int64(len(r.services)-1), serviceJSON).Err()
	})

	return err
}

// RemoveService removes a backend service from Redis
func (r *RedisRegistry) RemoveService(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	err := r.removeService(name, func() error {
		return r.redisClient.LRem(r.ctx, "services", 0, name).Err()
	})

	return err
}

// GetServices returns a copy of the current list of backend services
func (r *RedisRegistry) GetServices() []*BackendService {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.getServices()
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

	for _, service := range services {
		var backendService BackendService

		err = json.Unmarshal([]byte(service), &backendService)
		if err != nil {
			return err
		}

		backendService.Init()
		r.services = append(r.services, &backendService)
	}

	return nil
}
