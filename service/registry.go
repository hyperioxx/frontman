package service

import (
	"context"
	"github.com/Frontman-Labs/frontman/config"
	"github.com/Frontman-Labs/frontman/gateway"
	"sync"
)

// ServiceRegistry holds the methods to interact with the backend service registry
type ServiceRegistry interface {
	AddService(service *BackendService) error
	UpdateService(service *BackendService) error
	RemoveService(name string) error
	GetServices() []*BackendService
}

type baseRegistry struct {
	mutex       sync.RWMutex
	services    []*BackendService
	routingTrie *gateway.RoutingTrie
}

func NewServiceRegistry(ctx context.Context, serviceType string, config *config.Config) (ServiceRegistry, error) {
	switch serviceType {
	case "redis":
		redisClient, err := NewRedisClient(ctx, config.GlobalConfig.RedisURI)
		if err != nil {
			return nil, err
		}
		redisBackendServices, err := NewRedisRegistry(ctx, redisClient, config.GlobalConfig.RedisNamespace)
		if err != nil {
			return nil, err
		}
		return redisBackendServices, nil
	case "yaml":
		yamlBackendServices, err := NewYAMLServiceRegistry(config.GlobalConfig.ServicesFile)
		if err != nil {
			return nil, err
		}
		return yamlBackendServices, nil
	case "mongo":
		mongoClient, err := NewMongoClient(ctx, config.GlobalConfig.MongoURI)
		if err != nil {
			return nil, err
		}
		mongoBackendServices, err := NewMongoServiceRegistry(ctx, mongoClient, config.GlobalConfig.MongoDatabaseName, config.GlobalConfig.MongoCollectionName)
		if err != nil {
			return nil, err
		}
		return mongoBackendServices, nil
	default:
		return nil, ErrUnsupportedServiceType{serviceType: serviceType}
	}
}

func (r *baseRegistry) getServices() []*BackendService {
	services := make([]*BackendService, len(r.services))
	copy(services, r.services)
	return services
}

func (r *baseRegistry) addService(service *BackendService, apply func() error) error {
	old := r.getServices()

	for _, s := range r.services {
		if s.Name == service.Name {
			return ErrServiceExists{Name: service.Name}
		}
	}

	r.services = append(r.services, service)
	err := apply()
	if err != nil {
		r.services = old
		return err
	}

	return nil
}

func (r *baseRegistry) updateService(service *BackendService, apply func() error) error {
	old := r.getServices()

	for i, s := range r.services {
		if s.Name == service.Name {
			r.services[i] = service
			err := apply()
			if err != nil {
				r.services = old
				return err
			}

			return nil
		}
	}

	return ErrServiceNotFound{Name: service.Name}
}

func (r *baseRegistry) removeService(name string, apply func() error) error {
	old := r.getServices()

	for i, s := range r.services {
		if s.Name == name {
			r.services = append(r.services[:i], r.services[i+1:]...)
			err := apply()
			if err != nil {
				r.services = old
				return err
			}

			return nil
		}
	}

	return ErrServiceNotFound{Name: name}
}
