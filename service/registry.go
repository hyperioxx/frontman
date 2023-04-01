package service

import (
	"context"
	"github.com/Frontman-Labs/frontman/config"
	"sync"
)

// ServiceRegistry holds the methods to interact with the backend service registry
type ServiceRegistry interface {
	AddService(service *BackendService) error
	UpdateService(service *BackendService) error
	RemoveService(name string) error
	GetServices() []*BackendService
	GetTrie() *RoutingTrie
}

type baseRegistry struct {
	mutex       *sync.RWMutex
	services    []*BackendService
	routingTrie *RoutingTrie
}

func NewServiceRegistry(ctx context.Context, serviceType string, config *config.Config) (ServiceRegistry, error) {
	var (
		reg ServiceRegistry
		mu  sync.RWMutex
		err error
	)

	baseReg := baseRegistry{
		mutex: &mu,
		routingTrie: &RoutingTrie{
			mutex: &mu,
		},
	}

	switch serviceType {
	case "redis":
		redisClient, err := NewRedisClient(ctx, config.GlobalConfig.RedisURI)
		if err != nil {
			return nil, err
		}
		reg, err = NewRedisRegistry(ctx, redisClient, config.GlobalConfig.RedisNamespace, &baseReg)
		if err != nil {
			return nil, err
		}
	case "yaml":
		reg, err = NewYAMLServiceRegistry(config.GlobalConfig.ServicesFile, &baseReg)
		if err != nil {
			return nil, err
		}
	case "mongo":
		mongoClient, err := NewMongoClient(ctx, config.GlobalConfig.MongoURI)
		if err != nil {
			return nil, err
		}
		reg, err = NewMongoServiceRegistry(ctx, mongoClient, config.GlobalConfig.MongoDatabaseName, config.GlobalConfig.MongoCollectionName, &baseReg)
		if err != nil {
			return nil, err
		}
	case "memory": // For testing
		reg = NewMemoryServiceRegistry(&baseReg)
	default:
		return nil, ErrUnsupportedServiceType{serviceType: serviceType}
	}

	// Initialise routing trie
	baseReg.routingTrie.BuildRoutes(reg.GetServices())

	return reg, nil
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

	r.routingTrie.BuildRoutes(r.services)

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

			r.routingTrie.BuildRoutes(r.services)
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

			r.routingTrie.BuildRoutes(r.services)
			return nil
		}
	}

	return ErrServiceNotFound{Name: name}
}

func (r *baseRegistry) GetTrie() *RoutingTrie {
	return r.routingTrie
}
