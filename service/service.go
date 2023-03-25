package service

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Frontman-Labs/frontman/auth"
	"github.com/Frontman-Labs/frontman/config"
	"github.com/Frontman-Labs/frontman/loadbalancer"
	"github.com/Frontman-Labs/frontman/oauth"
)

// ServiceRegistry holds the methods to interact with the backend service registry
type ServiceRegistry interface {
	AddService(service *BackendService) error
	UpdateService(service *BackendService) error
	RemoveService(name string) error
	GetServices() []*BackendService
}

type baseRegistry struct {
	services []*BackendService
	mutex    sync.RWMutex
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

// BackendService holds the details of a backend service
type BackendService struct {
	Name               string             `json:"name" yaml:"name"`
	Scheme             string             `json:"scheme" yaml:"scheme"`
	UpstreamTargets    []string           `json:"upstreamTargets" yaml:"upstreamTargets"`
	Path               string             `json:"path,omitempty" yaml:"path,omitempty"`
	Domain             string             `json:"domain" yaml:"domain"`
	HealthCheck        string             `json:"healthCheck" yaml:"healthCheck"`
	RetryAttempts      int                `json:"retryAttempts,omitempty" yaml:"retryAttempts,omitempty"`
	Timeout            time.Duration      `json:"timeout" yaml:"timeout"`
	MaxIdleConns       int                `json:"maxIdleConns,omitempty" yaml:"maxIdleConns,omitempty"`
	MaxIdleTime        time.Duration      `json:"maxIdleTime" yaml:"maxIdleTime"`
	StripPath          bool               `json:"stripPath,omitempty" yaml:"stripPath,omitempty"`
	AuthConfig         *config.AuthConfig `json:"auth,omitempty" yaml:"auth,omitempty"`
	LoadBalancerPolicy LoadBalancerPolicy `json:"loadBalancerPolicy,omitempty" yaml:"loadBalancerPolicy,omitempty"`

	loadBalancer   loadbalancer.LoadBalancer
	provider       oauth.OAuthProvider
	tokenValidator *auth.TokenValidator
}

type LoadBalancerPolicy struct {
	Type    string        `json:"type" yaml:"type"`
	Options PolicyOptions `json:"options,omitempty" yaml:"options,omitempty"`
}

type PolicyOptions struct {
	Weights []int `json:"weights,omitempty" yaml:"weights,omitempty"`
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

func (bs *BackendService) setTokenValidator() {
	if bs.AuthConfig == nil {
		return
	}
	validator, err := auth.GetTokenValidator(*bs.AuthConfig)
	if err != nil {
		log.Printf("Error adding auth to backend service: %s: %s", bs.Name, err.Error())
	} else {
		bs.tokenValidator = &validator
	}
}

func (bs *BackendService) GetTokenValidator() auth.TokenValidator {
	if bs.AuthConfig != nil && bs.tokenValidator == nil {
		// Token validator has not been instantiated for this backend service
		// Instantiating here to avoid having to call setTokenValidator on each update/add
		bs.setTokenValidator()
	}
	return *bs.tokenValidator
}

func (bs *BackendService) GetUserDataHeader() string {
	if bs.AuthConfig.UserDataHeader != "" {
		return bs.AuthConfig.UserDataHeader
	}
	return "user"
}

func (bs *BackendService) GetLoadBalancer() loadbalancer.LoadBalancer {
	return bs.loadBalancer
}

func (bs *BackendService) setLoadBalancer() {
	switch bs.LoadBalancerPolicy.Type {
	case loadbalancer.Random:
		bs.loadBalancer = loadbalancer.NewRandomLoadBalancer()
	case loadbalancer.RoundRobin:
		bs.loadBalancer = loadbalancer.NewRoundRobinLoadBalancer()
	case loadbalancer.WeightedRoundRobin:
		bs.loadBalancer = loadbalancer.NewWRoundRobinLoadBalancer(bs.LoadBalancerPolicy.Options.Weights)
	case loadbalancer.LeastConnection:
		bs.loadBalancer = loadbalancer.NewLeastConnLoadBalancer(bs.UpstreamTargets, nil)
	case loadbalancer.WeightedLeastConnection:
		bs.loadBalancer = loadbalancer.NewLeastConnLoadBalancer(bs.UpstreamTargets, bs.LoadBalancerPolicy.Options.Weights)
	default:
		bs.loadBalancer = loadbalancer.NewRoundRobinLoadBalancer()
	}
}

func (bs *BackendService) Init() {
	bs.setTokenValidator()
	bs.setLoadBalancer()
}
