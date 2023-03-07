package service

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"time"

	"github.com/hyperioxx/frontman/config"
	"github.com/hyperioxx/frontman/loadbalancer"
	"github.com/hyperioxx/frontman/oauth"
)

// ServiceRegistry holds the methods to interact with the backend service registry
type ServiceRegistry interface {
	AddService(service *BackendService) error
	UpdateService(service *BackendService) error
	RemoveService(name string) error
	GetServices() []*BackendService
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
		return nil, fmt.Errorf("unsupported service type: %s", serviceType)
	}
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
	LoadBalancerPolicy LoadBalancerPolicy `json:"loadBalancerPolicy,omitempty" yaml:"loadBalancerPolicy,omitempty"`
	loadBalancer       loadbalancer.LoadBalancer
	provider           oauth.OAuthProvider
}

type LoadBalancerPolicy struct {
	Type    string        `json:"type" yaml:"type"`
	Options PolicyOptions `json:"options,omitempty" yaml:"options,omitempty"`
}

type PolicyOptions struct {
	WeightedOptions WeightedOptions `json:"weighted,omitempty" yaml:"weighted,omitempty"`
}

type WeightedOptions struct {
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
